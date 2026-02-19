package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	remoteAppName  = "fly"
	remoteFileName = "remote.json"
)

type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	SSHURL   string `json:"ssh_url"`
}

type Cache struct {
	FetchedAt time.Time `json:"fetched_at"`
	Repos     []Repo    `json:"repos"`
}

type RemoteStore struct {
	path string
}

func NewRemoteStore(path string) *RemoteStore {
	return &RemoteStore{path: path}
}

func DefaultRemoteStore() (*RemoteStore, error) {
	baseDir, err := CacheDir(remoteAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	return &RemoteStore{path: filepath.Join(baseDir, remoteFileName)}, nil
}

func (s *RemoteStore) Load() (Cache, bool, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cache{}, false, nil
		}

		return Cache{}, false, fmt.Errorf("read remote cache: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return Cache{}, false, fmt.Errorf("parse remote cache: %w", err)
	}

	return cache, true, nil
}

func (s *RemoteStore) Save(cache Cache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("marshal remote cache: %w", err)
	}

	if err := WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write remote cache: %w", err)
	}

	return nil
}
