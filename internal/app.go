package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type App struct {
	sources []Source
	picker  Picker
}

type Picker interface {
	Pick(string, []Candidate) (int, bool, error)
}

type Refresher interface {
	Launch()
}

type Pruner interface {
	Launch()
}

type Cloner interface {
	Clone(Repo) (string, error)
}

func NewApp(sources []Source, picker Picker) (*App, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("sources required")
	}
	if picker == nil {
		return nil, fmt.Errorf("picker required")
	}

	return &App{
		sources: sources,
		picker:  picker,
	}, nil
}

func (a *App) Refresh(ctx context.Context) error {
	for _, source := range a.sources {
		refreshable, ok := source.(Refreshable)
		if !ok {
			continue
		}
		if err := refreshable.Refresh(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Track(repoPath string) error {
	for _, source := range a.sources {
		trackable, ok := source.(Trackable)
		if !ok {
			continue
		}
		if err := trackable.Track(repoPath); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) Query(query string, stdout io.Writer) error {
	candidates, err := a.loadSources(query)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no matches for %q", query)
	}
	selected, err := a.selectCandidate(query, candidates)
	if err != nil {
		return err
	}
	if selected == nil {
		return nil
	}

	path, err := a.resolveCandidate(*selected)
	if err != nil {
		return err
	}

	if err := a.Track(path); err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, path)
	return err
}

func (a *App) Prune() error {
	for _, source := range a.sources {
		prunable, ok := source.(Prunable)
		if !ok {
			continue
		}
		if err := prunable.Prune(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) selectCandidate(query string, candidates []Candidate) (*Candidate, error) {
	if len(candidates) == 1 {
		selected := candidates[0]
		return &selected, nil
	}

	selectedIndex, ok, err := a.picker.Pick(query, candidates)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	if selectedIndex < 0 || selectedIndex >= len(candidates) {
		return nil, nil
	}

	selected := candidates[selectedIndex]

	return &selected, nil
}

func (a *App) loadSources(query string) ([]Candidate, error) {
	var (
		candidates []Candidate
		noRepos    int
	)

	for _, source := range a.sources {
		result, err := source.Load(query)
		if err != nil {
			if errors.Is(err, ErrNoReposTracked) {
				noRepos++
				continue
			}
			return nil, err
		}

		for _, candidate := range result {
			sourceName := candidate.Meta[CandidateMetaSource]
			if sourceName == "" {
				return nil, fmt.Errorf("candidate source missing")
			}

			candidate.resolver = source

			candidates = append(candidates, candidate)
		}
	}

	if noRepos == len(a.sources) {
		return nil, ErrNoReposTracked
	}

	return candidates, nil
}

func (a *App) resolveCandidate(candidate Candidate) (string, error) {
	if candidate.resolver == nil {
		return "", fmt.Errorf("no source to resolve candidate")
	}

	path, err := candidate.resolver.Resolve(candidate)
	if err != nil {
		if errors.Is(err, ErrUnsupportedCandidate) {
			return "", fmt.Errorf("selected source cannot resolve candidate")
		}

		return "", err
	}
	if path == "" {
		return "", fmt.Errorf("selected source returned empty path")
	}

	return path, nil
}
