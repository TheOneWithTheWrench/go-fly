package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type PruneStateStore struct {
	path string
}

func NewPruneStateStore(path string) *PruneStateStore {
	return &PruneStateStore{path: path}
}

const pruneAppName = "fly"

type PruneState struct {
	LastPrunedAt time.Time `json:"last_pruned_at"`
}

func DefaultPruneStateStore() (*PruneStateStore, error) {
	baseDir, err := internal.CacheDir(pruneAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	return &PruneStateStore{path: filepath.Join(baseDir, "prune.json")}, nil
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
