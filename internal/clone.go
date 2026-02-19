package internal

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func Destination(repo Repo, cwd string) (string, error) {
	name := repo.Name
	if name == "" {
		name = path.Base(repo.FullName)
	}

	dest := filepath.Join(cwd, name)
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", fmt.Errorf("resolve clone path: %w", err)
	}

	return absDest, nil
}

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
