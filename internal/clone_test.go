package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckDestination(t *testing.T) {
	t.Run("return false when destination missing", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "repo")

		got, err := internal.CheckDestination(path)

		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("return true when destination has .git dir", func(t *testing.T) {
		root := t.TempDir()
		dest := filepath.Join(root, "repo")
		require.NoError(t, os.MkdirAll(filepath.Join(dest, ".git"), 0o755))

		got, err := internal.CheckDestination(dest)

		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("return true when destination has .git file", func(t *testing.T) {
		root := t.TempDir()
		dest := filepath.Join(root, "repo")
		require.NoError(t, os.MkdirAll(dest, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dest, ".git"), []byte("gitdir: /tmp"), 0o644))

		got, err := internal.CheckDestination(dest)

		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("return error when destination exists without .git", func(t *testing.T) {
		root := t.TempDir()
		dest := filepath.Join(root, "repo")
		require.NoError(t, os.MkdirAll(dest, 0o755))

		_, err := internal.CheckDestination(dest)

		assert.Error(t, err)
	})

	t.Run("return error when destination is file", func(t *testing.T) {
		root := t.TempDir()
		dest := filepath.Join(root, "repo")
		require.NoError(t, os.WriteFile(dest, []byte("nope"), 0o644))

		_, err := internal.CheckDestination(dest)

		assert.Error(t, err)
	})
}
