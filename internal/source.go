package internal

import (
	"context"
	"errors"
	"io"
)

type Source interface {
	Load(context.Context, string) ([]Candidate, error)
	Resolve(Candidate) (string, error)
}

type Refreshable interface {
	Refresh(context.Context, RefreshOutput) error
}

type RefreshOutput interface {
	io.Writer
	SetStatus(string) error
	ClearStatus() error
}

type Trackable interface {
	Track(string) error
}

type Prunable interface {
	Prune() error
}

var ErrNoReposTracked = errors.New("no repos tracked yet")

var ErrUnsupportedCandidate = errors.New("unsupported candidate")
