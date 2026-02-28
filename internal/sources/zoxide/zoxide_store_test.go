package zoxide_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal/sources/zoxide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	var (
		newStore = func(path string) *zoxide.Store {
			return zoxide.NewStore(path)
		}
	)

	t.Run("return empty when file missing", func(t *testing.T) {
		var (
			path = filepath.Join(t.TempDir(), "zoxide.json")
			sut  = newStore(path)
		)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.False(t, exists)
		assert.True(t, got.FetchedAt.IsZero())
		assert.Empty(t, got.Backend)
		assert.Len(t, got.Matches, 0)
	})

	t.Run("save and load cache", func(t *testing.T) {
		var (
			path  = filepath.Join(t.TempDir(), "zoxide.json")
			sut   = newStore(path)
			when  = time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
			cache = zoxide.Cache{
				FetchedAt: when,
				Backend:   "zoxide",
				Matches: []zoxide.Match{
					{Path: "/tmp/repo", Score: 99},
				},
			}
		)

		err := sut.Save(cache)

		require.NoError(t, err)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, cache, got)
	})
}
