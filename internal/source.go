package internal

import (
	"context"
	"errors"
)

type Source interface {
	Load(query string) ([]Candidate, error)
}

type Refreshable interface {
	Refresh(context.Context) error
}

type Trackable interface {
	Track(string) error
}

type Prunable interface {
	Prune() error
}

type LocalCleaner interface {
	Remove(string) error
}

var ErrNoReposTracked = errors.New("no repos tracked yet")
