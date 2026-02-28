package zoxide

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
	FetchedAt time.Time `json:"fetched_at"`
	Backend   string    `json:"backend"`
	Matches   []Match   `json:"matches"`
}

type Store struct {
	path string
}

type StoreOption interface {
	apply(*storeOptions) error
}

type storeOptionFunc func(*storeOptions) error

func (f storeOptionFunc) apply(opts *storeOptions) error {
	return f(opts)
}

type storeOptions struct {
	path string
}

const (
	zoxideStoreAppName  = "fly"
	zoxideStoreFileName = "zoxide.json"
)

func WithStorePath(path string) StoreOption {
	return storeOptionFunc(func(opts *storeOptions) error {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("store path required")
		}

		opts.path = path
		return nil
	})
}

func NewStore(options ...StoreOption) (*Store, error) {
	baseDir, err := internal.CacheDir(zoxideStoreAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}

	opts := storeOptions{path: filepath.Join(baseDir, zoxideStoreFileName)}
	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option.apply(&opts); err != nil {
			return nil, err
		}
	}

	return &Store{path: opts.path}, nil
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
