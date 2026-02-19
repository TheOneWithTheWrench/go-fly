package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	indexAppName  = "fly"
	indexFileName = "index.json"
)

type Entry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type IndexStore struct {
	path string
}

func NewIndexStore(path string) *IndexStore {
	return &IndexStore{path: path}
}

func DefaultIndexStore() (*IndexStore, error) {
	baseDir, err := DataDir(indexAppName)
	if err != nil {
		return nil, fmt.Errorf("resolve data dir: %w", err)
	}

	return &IndexStore{path: filepath.Join(baseDir, indexFileName)}, nil
}

func (s *IndexStore) Load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, nil
		}

		return nil, fmt.Errorf("read index: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	return entries, nil
}

func (s *IndexStore) Save(entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal index: %w", err)
	}

	if err := WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	return nil
}

func (s *IndexStore) Upsert(entry Entry) error {
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
