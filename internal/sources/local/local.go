package local

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Source struct {
	store       *IndexStore
	pruneStore  *PruneStateStore
	pruneLaunch internal.Pruner
	now         func() time.Time
	pruneMu     sync.Mutex
}

type Option func(*options) error

type options struct {
	pruneLaunch internal.Pruner
}

func WithPruneLauncher(pruneLaunch internal.Pruner) Option {
	return func(opts *options) error {
		if pruneLaunch == nil {
			return fmt.Errorf("prune launcher required")
		}

		opts.pruneLaunch = pruneLaunch
		return nil
	}
}

func New(sourceOptions ...Option) (*Source, error) {
	defaultOpts := options{pruneLaunch: newDetachedPruner()}
	for _, option := range sourceOptions {
		if err := option(&defaultOpts); err != nil {
			return nil, err
		}
	}

	store, err := NewIndexStore()
	if err != nil {
		return nil, err
	}
	pruneStore, err := NewPruneStateStore()
	if err != nil {
		return nil, err
	}

	return &Source{
		store:       store,
		pruneStore:  pruneStore,
		pruneLaunch: defaultOpts.pruneLaunch,
		now:         time.Now,
	}, nil
}

func (s *Source) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	_ = ctx

	entries, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	total := len(entries)
	if total == 0 {
		return nil, internal.ErrNoReposTracked
	}
	filtered := FilterEntries(query, entries)

	s.loadPruneAndMaybeStart()

	candidates := make([]internal.Candidate, 0, len(filtered))
	for _, entry := range filtered {
		candidates = append(candidates, internal.Candidate{
			Meta: map[string]string{
				internal.CandidateMetaSource: internal.CandidateSourceLocal,
				internal.CandidateMetaName:   entry.Name,
				internal.CandidateMetaPath:   entry.Path,
			},
		})
	}

	return candidates, nil
}

func (s *Source) Track(repoPath string) error {
	if repoPath == "" {
		return fmt.Errorf("repo path required")
	}

	entry := internal.Entry{
		Name: filepath.Base(repoPath),
		Path: repoPath,
	}

	return s.store.Upsert(entry)
}

func (s *Source) Resolve(candidate internal.Candidate) (string, error) {
	if candidate.Meta[internal.CandidateMetaSource] != internal.CandidateSourceLocal {
		return "", internal.ErrUnsupportedCandidate
	}

	path := candidate.Meta[internal.CandidateMetaPath]
	if path == "" {
		return "", internal.ErrUnsupportedCandidate
	}

	valid, err := internal.CheckDestination(path)
	if err != nil || !valid {
		if removeErr := s.Remove(path); removeErr != nil {
			return "", removeErr
		}
		return "", fmt.Errorf("repo no longer exists: %s", path)
	}

	return path, nil
}

func (s *Source) Prune() error {
	entries, err := s.store.Load()
	if err != nil {
		return err
	}

	kept := make([]internal.Entry, 0, len(entries))
	for _, entry := range entries {
		valid, err := internal.CheckDestination(entry.Path)
		if err != nil || !valid {
			continue
		}
		kept = append(kept, entry)
	}

	if err := s.store.Save(kept); err != nil {
		return err
	}

	now := s.now().UTC()
	state := PruneState{LastPrunedAt: now, StartedAt: now}
	return s.pruneStore.Save(state)
}

func (s *Source) Remove(path string) error {
	entries, err := s.store.Load()
	if err != nil {
		return err
	}

	updated := removeEntry(entries, path)
	return s.store.Save(updated)
}

func (s *Source) loadPruneAndMaybeStart() {
	s.pruneMu.Lock()
	defer s.pruneMu.Unlock()

	state, exists, err := s.pruneStore.Load()
	if err != nil {
		state = PruneState{}
		exists = false
	}

	if ShouldPrune(state, exists) {
		s.launchPrune(state, exists)
	}
}

func (s *Source) launchPrune(state PruneState, exists bool) {
	if !ShouldLaunchPrune(state, exists, s.now()) {
		return
	}

	startedAt := s.now().UTC()
	if err := s.pruneStore.Save(PruneState{LastPrunedAt: state.LastPrunedAt, StartedAt: startedAt}); err != nil {
		return
	}

	s.pruneLaunch.Launch()
}

func removeEntry(entries []internal.Entry, path string) []internal.Entry {
	filtered := make([]internal.Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Path == path {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

func FilterEntries(query string, entries []internal.Entry) []internal.Entry {
	if strings.TrimSpace(query) == "" {
		return entries
	}

	query = strings.ToLower(query)
	filtered := make([]internal.Entry, 0, len(entries))

	for _, entry := range entries {
		name := strings.ToLower(entry.Name)
		path := strings.ToLower(entry.Path)
		if strings.Contains(name, query) || strings.Contains(path, query) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
