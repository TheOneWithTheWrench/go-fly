package remote

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func Destination(repo internal.Repo, cwd string, groupByOwner bool) (string, error) {
	name := repo.Name
	if name == "" {
		name = path.Base(repo.FullName)
	}

	dest := filepath.Join(cwd, name)
	if groupByOwner {
		owner, _, ok := strings.Cut(strings.TrimSpace(repo.FullName), "/")
		if ok && owner != "" {
			dest = filepath.Join(cwd, owner, name)
		}
	}

	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", fmt.Errorf("resolve clone path: %w", err)
	}

	return absDest, nil
}
