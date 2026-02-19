package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Find(path string) (string, bool, error) {
	if path == "" {
		return "", false, fmt.Errorf("path required")
	}

	current, err := filepath.Abs(path)
	if err != nil {
		return "", false, fmt.Errorf("resolve path: %w", err)
	}

	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, true, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", false, fmt.Errorf("stat %s: %w", gitPath, err)
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}

		current = parent
	}
}
