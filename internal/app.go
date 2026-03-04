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

type Option func(*options) error

type options struct {
	picker Picker
}

func WithPicker(picker Picker) Option {
	return func(opts *options) error {
		if picker == nil {
			return fmt.Errorf("picker required")
		}

		opts.picker = picker
		return nil
	}
}

func NewApp(sources []Source, sourceOptions ...Option) (*App, error) {
	opts := options{
		picker: PickerFunc(func(query string, candidates []Candidate) (int, bool, error) {
			return Pick(query, candidates)
		}),
	}
	for _, option := range sourceOptions {
		if err := option(&opts); err != nil {
			return nil, err
		}
	}

	return &App{
		sources: sources,
		picker:  opts.picker,
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

func (a *App) Query(ctx context.Context, query string, stdout io.Writer) error {
	candidates, err := a.loadSources(ctx, query)
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

func (a *App) loadSources(ctx context.Context, query string) ([]Candidate, error) {
	var (
		candidates []Candidate
		noRepos    int
		firstErr   error
	)

	for _, source := range a.sources {
		result, err := source.Load(ctx, query)
		if err != nil {
			if errors.Is(err, ErrNoReposTracked) {
				noRepos++
				continue
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
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
	if len(candidates) == 0 && firstErr != nil {
		return nil, firstErr
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
