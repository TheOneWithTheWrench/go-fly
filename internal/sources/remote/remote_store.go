package remote

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

type Cache struct {
	FetchedAt time.Time       `json:"fetched_at"`
	Repos     []internal.Repo `json:"repos"`
}

type RemoteStore struct {
	path string
}

type StoreOption func(*storeOptions) error

type storeOptions struct {
	path string
}

const (
	remoteAppName  = "fly"
	remoteFileName = "remote.json"
)

func WithStorePath(path string) StoreOption {
	return func(opts *storeOptions) error {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("store path required")
		}

		opts.path = path
		return nil
	}
}

func NewRemoteStore(options ...StoreOption) (*RemoteStore, error) {
	defaultOpts := storeOptions{}

	baseDir, err := internal.CacheDir(remoteAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}
	defaultOpts.path = filepath.Join(baseDir, remoteFileName)

	for _, option := range options {
		if err := option(&defaultOpts); err != nil {
			return nil, err
		}
	}

	return &RemoteStore{path: defaultOpts.path}, nil
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

	if err := internal.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write remote cache: %w", err)
	}

	return nil
}
