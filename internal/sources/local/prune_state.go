package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type PruneStateStore struct {
	path string
}

type PruneStateStoreOption func(*pruneStateStoreOptions) error

type pruneStateStoreOptions struct {
	path string
}

func WithPruneStateStorePath(path string) PruneStateStoreOption {
	return func(opts *pruneStateStoreOptions) error {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("prune state path required")
		}

		opts.path = path
		return nil
	}
}

const pruneAppName = "fly"

type PruneState struct {
	LastPrunedAt time.Time `json:"last_pruned_at"`
	StartedAt    time.Time `json:"started_at"`
}

func NewPruneStateStore(options ...PruneStateStoreOption) (*PruneStateStore, error) {
	baseDir, err := internal.CacheDir(pruneAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	opts := pruneStateStoreOptions{path: filepath.Join(baseDir, "prune.json")}
	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option(&opts); err != nil {
			return nil, err
		}
	}

	return &PruneStateStore{path: opts.path}, nil
}

func (s *PruneStateStore) Load() (PruneState, bool, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return PruneState{}, false, nil
		}
		return PruneState{}, false, fmt.Errorf("read prune state: %w", err)
	}

	var state PruneState
	if err := json.Unmarshal(data, &state); err != nil {
		return PruneState{}, false, fmt.Errorf("parse prune state: %w", err)
	}

	return state, true, nil
}

func (s *PruneStateStore) Save(state PruneState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal prune state: %w", err)
	}

	if err := internal.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write prune state: %w", err)
	}

	return nil
}
