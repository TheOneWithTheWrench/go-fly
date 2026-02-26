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
	cloner  Cloner
}

type RemoteFetcher interface {
	FetchAll(context.Context) ([]Repo, error)
}

type Picker interface {
	Pick(string, []Candidate) (Candidate, bool, error)
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

func NewApp(sources []Source, picker Picker, cloner Cloner) (*App, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("sources required")
	}
	if picker == nil {
		return nil, fmt.Errorf("picker required")
	}
	if cloner == nil {
		return nil, fmt.Errorf("cloner required")
	}

	return &App{
		sources: sources,
		picker:  picker,
		cloner:  cloner,
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

	if selected.Kind == KindLocal {
		valid, err := CheckDestination(selected.Local.Path)
		if err != nil || !valid {
			if saveErr := a.removeLocalEntry(selected.Local.Path); saveErr != nil {
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

	if err := a.Track(dest); err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, dest)
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
		candidates = append(candidates, result...)
	}

	if noRepos == len(a.sources) {
		return nil, ErrNoReposTracked
	}

	return candidates, nil
}

func (a *App) removeLocalEntry(path string) error {
	for _, source := range a.sources {
		cleaner, ok := source.(LocalCleaner)
		if !ok {
			continue
		}
		if err := cleaner.Remove(path); err != nil {
			return err
		}
	}

	return nil
}
