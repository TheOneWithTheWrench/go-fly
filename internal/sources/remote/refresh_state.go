package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

const RefreshLaunchCooldown = 30 * time.Second

type RefreshState struct {
	StartedAt time.Time `json:"started_at"`
}

type RefreshStateStore struct {
	path string
}

func NewRefreshStateStore() (*RefreshStateStore, error) {
	baseDir, err := internal.CacheDir(remoteAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	return &RefreshStateStore{path: filepath.Join(baseDir, "remote_refresh_state.json")}, nil
}

func (s *RefreshStateStore) Load() (RefreshState, bool, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return RefreshState{}, false, nil
		}

		return RefreshState{}, false, fmt.Errorf("read refresh state: %w", err)
	}

	var state RefreshState
	if err := json.Unmarshal(data, &state); err != nil {
		return RefreshState{}, false, fmt.Errorf("parse refresh state: %w", err)
	}

	return state, true, nil
}

func (s *RefreshStateStore) Save(state RefreshState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal refresh state: %w", err)
	}

	if err := internal.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write refresh state: %w", err)
	}

	return nil
}

func ShouldLaunchRefresh(state RefreshState, exists bool, now time.Time) bool {
	if !exists {
		return true
	}
	if state.StartedAt.IsZero() {
		return true
	}
	if now.Sub(state.StartedAt) >= RefreshLaunchCooldown {
		return true
	}

	return false
}
