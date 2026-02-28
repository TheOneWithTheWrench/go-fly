package local_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneStateStore(t *testing.T) {
	var (
		newStore = func(t *testing.T, path string) *local.PruneStateStore {
			t.Helper()

			sut, err := local.NewPruneStateStore(local.WithPruneStateStorePath(path))
			require.NoError(t, err)

			return sut
		}
	)

	t.Run("return empty when file missing", func(t *testing.T) {
		var (
			path = filepath.Join(t.TempDir(), "prune.json")
			sut  = newStore(t, path)
		)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.False(t, exists)
		assert.True(t, got.LastPrunedAt.IsZero())
	})

	t.Run("save and load state", func(t *testing.T) {
		var (
			path  = filepath.Join(t.TempDir(), "prune.json")
			sut   = newStore(t, path)
			when  = time.Date(2026, 2, 19, 13, 0, 0, 0, time.UTC)
			state = local.PruneState{LastPrunedAt: when}
		)

		err := sut.Save(state)

		require.NoError(t, err)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, state, got)
	})
}
