package internal_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPruneStateStore(t *testing.T) {
	var (
		newStore = func(path string) *internal.PruneStateStore {
			return internal.NewPruneStateStore(path)
		}
	)

	t.Run("return empty when file missing", func(t *testing.T) {
		var (
			path = filepath.Join(t.TempDir(), "prune.json")
			sut  = newStore(path)
		)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.False(t, exists)
		assert.True(t, got.LastPrunedAt.IsZero())
	})

	t.Run("save and load state", func(t *testing.T) {
		var (
			path  = filepath.Join(t.TempDir(), "prune.json")
			sut   = newStore(path)
			when  = time.Date(2026, 2, 19, 13, 0, 0, 0, time.UTC)
			state = internal.PruneState{LastPrunedAt: when}
		)

		err := sut.Save(state)

		require.NoError(t, err)

		got, exists, err := sut.Load()

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, state, got)
	})
}
