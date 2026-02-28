package remote_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	var (
		newStore = func(t *testing.T, path string) *remote.RemoteStore {
			t.Helper()

			sut, err := remote.NewRemoteStore(remote.WithStorePath(path))
			require.NoError(t, err)

			return sut
		}
	)

	t.Run("return empty when file missing", func(t *testing.T) {
		var (
			path = filepath.Join(t.TempDir(), "remote.json")
			sut  = newStore(t, path)
		)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.False(t, exists)
		assert.True(t, got.FetchedAt.IsZero())
		assert.Len(t, got.Repos, 0)
	})

	t.Run("save and load cache", func(t *testing.T) {
		var (
			path  = filepath.Join(t.TempDir(), "remote.json")
			sut   = newStore(t, path)
			when  = time.Date(2026, 2, 19, 10, 30, 0, 0, time.UTC)
			cache = remote.Cache{
				FetchedAt: when,
				Repos: []internal.Repo{
					{Name: "repo", FullName: "acme/repo", SSHURL: "git@github.com:acme/repo.git"},
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
