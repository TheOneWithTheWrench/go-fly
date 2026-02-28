package zoxide

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Cache struct {
	FetchedAt time.Time `json:"fetched_at"`
	Backend   string    `json:"backend"`
	Matches   []Match   `json:"matches"`
}

type Store struct {
	path string
}

const (
	zoxideStoreAppName  = "fly"
	zoxideStoreFileName = "zoxide.json"
)

func NewStore(path string) *Store {
	return &Store{path: path}
}

func DefaultStore() (*Store, error) {
	baseDir, err := internal.CacheDir(zoxideStoreAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	return &Store{path: filepath.Join(baseDir, zoxideStoreFileName)}, nil
}

func (s *Store) Load() (Cache, bool, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Cache{}, false, nil
		}

		return Cache{}, false, fmt.Errorf("read zoxide cache: %w", err)
	}

	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return Cache{}, false, fmt.Errorf("parse zoxide cache: %w", err)
	}

	return cache, true, nil
}

func (s *Store) Save(cache Cache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("marshal zoxide cache: %w", err)
	}

	if err := internal.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write zoxide cache: %w", err)
	}

	return nil
}
