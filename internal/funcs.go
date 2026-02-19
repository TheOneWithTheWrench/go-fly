package internal

import "context"

type PickerFunc func(string, []Candidate) (Candidate, bool, error)

func (f PickerFunc) Pick(query string, candidates []Candidate) (Candidate, bool, error) {
	return f(query, candidates)
}

type RefreshFunc func()

func (f RefreshFunc) Launch() {
	f()
}

type PruneFunc func()

func (f PruneFunc) Launch() {
	f()
}

type ClonerFunc func(Repo) (string, error)

func (f ClonerFunc) Clone(repo Repo) (string, error) {
	return f(repo)
}

type RunnerFunc func(context.Context, string, ...string) ([]byte, error)

func (f RunnerFunc) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f(ctx, name, args...)
}
