package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	var (
		newSut = func() func(string) (string, bool, error) {
			return internal.Find
		}
	)

	t.Run("return false when not inside repo", func(t *testing.T) {
		var (
			sut    = newSut()
			work   = filepath.Join(t.TempDir(), "no-repo")
			create = func() { require.NoError(t, os.MkdirAll(work, 0o755)) }
		)

		create()

		got, ok, err := sut(work)

		require.NoError(t, err)
		assert.False(t, ok)
		assert.Empty(t, got)
	})

	t.Run("detect .git file", func(t *testing.T) {
		var (
			sut     = newSut()
			root    = t.TempDir()
			nested  = filepath.Join(root, "subdir")
			gitFile = filepath.Join(root, ".git")
		)

		require.NoError(t, os.MkdirAll(nested, 0o755))
		require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /tmp/worktree"), 0o644))

		got, ok, err := sut(nested)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, root, got)
	})

	t.Run("find root from nested directory", func(t *testing.T) {
		var (
			sut      = newSut()
			rootDir  = t.TempDir()
			nested   = filepath.Join(rootDir, "one", "two")
			gitDir   = filepath.Join(rootDir, ".git")
			makeDirs = func() {
				require.NoError(t, os.MkdirAll(nested, 0o755))
				require.NoError(t, os.MkdirAll(gitDir, 0o755))
			}
		)

		makeDirs()

		got, ok, err := sut(nested)

		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, rootDir, got)
	})
}
