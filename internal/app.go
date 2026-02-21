package internal

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

type App struct {
	localStore    IndexStorage
	remoteStore   RemoteStorage
	remoteFetcher RemoteFetcher
	picker        Picker
	refreshLaunch RefreshLauncher
	pruneStore    PruneStateStorage
	pruneLaunch   PruneLauncher
	cloner        Cloner
}

type IndexStorage interface {
	Load() ([]Entry, error)
	Save([]Entry) error
	Upsert(Entry) error
}

type RemoteStorage interface {
	Load() (Cache, bool, error)
	Save(Cache) error
}

type RemoteFetcher interface {
	FetchAll(context.Context) ([]Repo, error)
}

type Picker interface {
	Pick(string, []Candidate) (Candidate, bool, error)
}

type RefreshLauncher interface {
	Launch()
}

type PruneStateStorage interface {
	Load() (PruneState, bool, error)
	Save(PruneState) error
}

type PruneLauncher interface {
	Launch()
}

type Cloner interface {
	Clone(Repo) (string, error)
}

const (
	RefreshTTL = 24 * time.Hour
	PruneTTL   = 24 * time.Hour
)

func NewApp(localStore IndexStorage, remoteStore RemoteStorage, ghClient RemoteFetcher, picker Picker, refreshLaunch RefreshLauncher, pruneStore PruneStateStorage, pruneLaunch PruneLauncher, cloner Cloner) (*App, error) {
	if localStore == nil {
		return nil, fmt.Errorf("local store required")
	}
	if remoteStore == nil {
		return nil, fmt.Errorf("remote store required")
	}
	if ghClient == nil {
		return nil, fmt.Errorf("gh client required")
	}
	if picker == nil {
		return nil, fmt.Errorf("picker required")
	}
	if refreshLaunch == nil {
		return nil, fmt.Errorf("refresh launcher required")
	}
	if pruneStore == nil {
		return nil, fmt.Errorf("prune state store required")
	}
	if pruneLaunch == nil {
		return nil, fmt.Errorf("prune launcher required")
	}
	if cloner == nil {
		return nil, fmt.Errorf("cloner required")
	}

	return &App{
		localStore:    localStore,
		remoteStore:   remoteStore,
		remoteFetcher: ghClient,
		picker:        picker,
		refreshLaunch: refreshLaunch,
		pruneStore:    pruneStore,
		pruneLaunch:   pruneLaunch,
		cloner:        cloner,
	}, nil
}

func (a *App) Refresh(ctx context.Context) error {
	repos, err := a.remoteFetcher.FetchAll(ctx)
	if err != nil {
		return fmt.Errorf("fetch repos: %w", err)
	}

	cache := Cache{
		FetchedAt: time.Now().UTC(),
		Repos:     repos,
	}

	return a.remoteStore.Save(cache)
}

func (a *App) Track(repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repo path required")
	}

	entry := Entry{
		Name: filepath.Base(repoPath),
		Path: repoPath,
	}

	return a.localStore.Upsert(entry)
}

func (a *App) Query(query string, stdout io.Writer) error {
	entries, err := a.loadEntries()
	if err != nil {
		return err
	}
	cache := a.loadCacheAndMaybeRefresh()
	a.loadPruneAndMaybeStart()

	if len(entries) == 0 && len(cache.Repos) == 0 {
		return fmt.Errorf("no repos tracked yet")
	}

	localMatches := Filter(query, entries)
	remoteMatches := FilterRemote(query, cache.Repos)
	if len(localMatches) == 0 && len(remoteMatches) == 0 {
		return fmt.Errorf("no matches for %q", query)
	}

	candidates := Build(localMatches, remoteMatches)
	selected, err := a.selectCandidate(query, candidates)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil
	}

	if selected.Kind == KindLocal {
		valid, err := CheckDestination(selected.Local.Path)
		if err != nil || !valid {
			updated := removeEntry(entries, selected.Local.Path)
			if saveErr := a.localStore.Save(updated); saveErr != nil {
				return saveErr
			}
			return fmt.Errorf("repo no longer exists: %s", selected.Local.Path)
		}

		_, err = fmt.Fprintln(stdout, selected.Local.Path)
		return err
	}

	dest, err := a.cloner.Clone(selected.Remote)
	if err != nil {
		return err
	}

	if err := a.localStore.Upsert(Entry{Name: filepath.Base(dest), Path: dest}); err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, dest)
	return err
}

func (a *App) Prune() error {
	entries, err := a.localStore.Load()
	if err != nil {
		return err
	}

	kept := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		valid, err := CheckDestination(entry.Path)
		if err != nil || !valid {
			continue
		}
		kept = append(kept, entry)
	}

	if err := a.localStore.Save(kept); err != nil {
		return err
	}

	state := PruneState{LastPrunedAt: time.Now().UTC()}
	return a.pruneStore.Save(state)
}

func (a *App) selectCandidate(query string, candidates []Candidate) (*Candidate, error) {
	if len(candidates) == 1 && candidates[0].Kind == KindLocal {
		selected := candidates[0]
		return &selected, nil
	}

	selected, ok, err := a.picker.Pick(query, candidates)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &selected, nil
}

func (a *App) maybeStartBackgroundRefresh(cache Cache, exists bool) {
	if !ShouldRefresh(cache, exists) {
		return
	}

	a.refreshLaunch.Launch()
}

func (a *App) maybeStartBackgroundPrune(state PruneState, exists bool) {
	if !ShouldPrune(state, exists) {
		return
	}

	a.pruneLaunch.Launch()
}

func (a *App) loadEntries() ([]Entry, error) {
	return a.localStore.Load()
}

func (a *App) loadCacheAndMaybeRefresh() Cache {
	cache, exists, err := a.remoteStore.Load()
	if err != nil {
		cache = Cache{}
		exists = false
	}

	a.maybeStartBackgroundRefresh(cache, exists)
	return cache
}

func (a *App) loadPruneAndMaybeStart() {
	state, exists, err := a.pruneStore.Load()
	if err != nil {
		state = PruneState{}
		exists = false
	}

	a.maybeStartBackgroundPrune(state, exists)
}

func ShouldRefresh(cache Cache, exists bool) bool {
	if !exists {
		return true
	}
	if cache.FetchedAt.IsZero() {
		return true
	}
	if time.Since(cache.FetchedAt) >= RefreshTTL {
		return true
	}
	return false
}

func ShouldPrune(state PruneState, exists bool) bool {
	if !exists {
		return true
	}
	if state.LastPrunedAt.IsZero() {
		return true
	}
	if time.Since(state.LastPrunedAt) >= PruneTTL {
		return true
	}
	return false
}

func removeEntry(entries []Entry, path string) []Entry {
	filtered := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == path {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}
