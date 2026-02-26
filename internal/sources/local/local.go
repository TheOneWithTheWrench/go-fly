package local

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Source struct {
	store       *IndexStore
	pruneStore  *PruneStateStore
	pruneLaunch internal.Pruner
	now         func() time.Time
}

func New(pruneLaunch internal.Pruner) (*Source, error) {
	store, err := DefaultIndexStore()
	if err != nil {
		return nil, err
	}
	pruneStore, err := DefaultPruneStateStore()
	if err != nil {
		return nil, err
	}
	if pruneLaunch == nil {
		return nil, fmt.Errorf("prune launcher required")
	}

	return &Source{
		store:       store,
		pruneStore:  pruneStore,
		pruneLaunch: pruneLaunch,
		now:         time.Now,
	}, nil
}

func (s *Source) Load(query string) ([]internal.Candidate, error) {
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
		candidates = append(candidates, internal.Candidate{Kind: internal.KindLocal, Local: entry})
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

	state := PruneState{LastPrunedAt: s.now().UTC()}
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
	state, exists, err := s.pruneStore.Load()
	if err != nil {
		state = PruneState{}
		exists = false
	}

	if ShouldPrune(state, exists) {
		s.pruneLaunch.Launch()
	}
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
