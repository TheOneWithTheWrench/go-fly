package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

func CheckDestination(dest string) (bool, error) {
	info, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat destination: %w", err)
	}

	if !info.IsDir() {
		return false, fmt.Errorf("destination path %q exists and is not a directory", dest)
	}

	gitPath := filepath.Join(dest, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat %s: %w", gitPath, err)
	}

	return false, fmt.Errorf("destination path %q already exists and is not a git repo", dest)
}
