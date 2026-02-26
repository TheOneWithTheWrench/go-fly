package internal

import "context"

type PickerFunc func(string, []Candidate) (int, bool, error)

func (f PickerFunc) Pick(query string, candidates []Candidate) (int, bool, error) {
	return f(query, candidates)
}

type RefresherFunc func()

func (f RefresherFunc) Launch() {
	f()
}

type PrunerFunc func()

func (f PrunerFunc) Launch() {
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
