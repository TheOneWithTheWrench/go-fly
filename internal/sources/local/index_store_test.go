package local_test

import (
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreLoad(t *testing.T) {
	var (
		newStore = func(path string) *local.IndexStore {
			return local.NewIndexStore(path)
		}
	)

	t.Run("return empty when file missing", func(t *testing.T) {
		var (
			path = filepath.Join(t.TempDir(), "index.json")
			sut  = newStore(path)
		)

		entries, err := sut.Load()

		require.NoError(t, err)
		assert.Len(t, entries, 0)
	})
}

func TestStoreUpsert(t *testing.T) {
	var (
		newStore = func(path string) *local.IndexStore {
			return local.NewIndexStore(path)
		}
	)

	t.Run("replace by matching path", func(t *testing.T) {
		var (
			path    = filepath.Join(t.TempDir(), "index.json")
			sut     = newStore(path)
			initial = []internal.Entry{{Name: "old", Path: "/tmp/repo"}}
		)

		err := sut.Save(initial)

		require.NoError(t, err)

		err = sut.Upsert(internal.Entry{Name: "new", Path: "/tmp/repo"})

		require.NoError(t, err)

		got, err := sut.Load()

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "new", got[0].Name)
		assert.Equal(t, "/tmp/repo", got[0].Path)
	})
}

func TestStoreSaveLoad(t *testing.T) {
	var (
		newStore = func(path string) *local.IndexStore {
			return local.NewIndexStore(path)
		}
	)

	t.Run("save and load entries", func(t *testing.T) {
		var (
			path    = filepath.Join(t.TempDir(), "index.json")
			sut     = newStore(path)
			entries = []internal.Entry{{Name: "repo", Path: "/tmp/repo"}}
		)

		err := sut.Save(entries)

		require.NoError(t, err)

		got, err := sut.Load()

		require.NoError(t, err)
		assert.Equal(t, entries, got)
	})
}
