package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type IndexStore struct {
	path string
}

type IndexStoreOption func(*indexStoreOptions) error

type indexStoreOptions struct {
	path string
}

const (
	indexAppName  = "fly"
	indexFileName = "index.json"
)

func WithIndexStorePath(path string) IndexStoreOption {
	return func(opts *indexStoreOptions) error {
		if strings.TrimSpace(path) == "" {
			return fmt.Errorf("index path required")
		}

		opts.path = path
		return nil
	}
}

func NewIndexStore(options ...IndexStoreOption) (*IndexStore, error) {
	baseDir, err := internal.DataDir(indexAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}

	opts := indexStoreOptions{path: filepath.Join(baseDir, indexFileName)}
	for _, option := range options {
		if option == nil {
			continue
		}

		if err := option(&opts); err != nil {
			return nil, err
		}
	}

	return &IndexStore{path: opts.path}, nil
}

func (s *IndexStore) Load() ([]internal.Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []internal.Entry{}, nil
		}

		return nil, fmt.Errorf("read index: %w", err)
	}

	var entries []internal.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	return entries, nil
}

func (s *IndexStore) Save(entries []internal.Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := internal.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	return nil
}

func (s *IndexStore) Upsert(entry internal.Entry) error {
	if entry.Path == "" {
		return fmt.Errorf("entry path required")
	}

	entries, err := s.Load()
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].Path == entry.Path {
			entries[i] = entry
			return s.Save(entries)
		}
	}

	entries = append(entries, entry)
	return s.Save(entries)
}
