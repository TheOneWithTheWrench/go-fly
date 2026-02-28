package local_test

import (
	"context"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create source with default pruner", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		sut, err := local.New()

		require.NoError(t, err)
		require.NotNil(t, sut)
	})
}

func TestNewOptions(t *testing.T) {
	t.Run("return error when with prune launcher option has nil pruner", func(t *testing.T) {
		_, err := local.New(local.WithPruneLauncher(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "prune launcher required")
	})

	t.Run("launch prune only once while stale", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		indexStore, err := local.NewIndexStore()
		require.NoError(t, err)
		require.NoError(t, indexStore.Save([]internal.Entry{{Name: "repo", Path: "/tmp/repo"}}))

		pruneStore, err := local.NewPruneStateStore()
		require.NoError(t, err)
		require.NoError(t, pruneStore.Save(local.PruneState{LastPrunedAt: time.Now().Add(-local.PruneTTL).Add(-time.Minute)}))

		var launches int
		sut, err := local.New(local.WithPruneLauncher(internal.PrunerFunc(func() {
			launches++
		})))
		require.NoError(t, err)

		_, err = sut.Load(context.Background(), "")
		require.NoError(t, err)
		_, err = sut.Load(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, 1, launches)
	})

	t.Run("launch prune once across multiple source instances", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", t.TempDir())
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		indexStore, err := local.NewIndexStore()
		require.NoError(t, err)
		require.NoError(t, indexStore.Save([]internal.Entry{{Name: "repo", Path: "/tmp/repo"}}))

		pruneStore, err := local.NewPruneStateStore()
		require.NoError(t, err)
		require.NoError(t, pruneStore.Save(local.PruneState{LastPrunedAt: time.Now().Add(-local.PruneTTL).Add(-time.Minute)}))

		var launches int
		newSource := func(t *testing.T) *local.Source {
			t.Helper()

			sut, newErr := local.New(local.WithPruneLauncher(internal.PrunerFunc(func() {
				launches++
			})))
			require.NoError(t, newErr)

			return sut
		}

		sut1 := newSource(t)
		sut2 := newSource(t)

		_, err = sut1.Load(context.Background(), "")
		require.NoError(t, err)
		_, err = sut2.Load(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, 1, launches)
	})
}
