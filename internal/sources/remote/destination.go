package remote

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func Destination(repo internal.Repo, cwd string) (string, error) {
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
