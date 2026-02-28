package local_test

import (
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/stretchr/testify/assert"
)

func TestShouldLaunchPrune(t *testing.T) {
	now := time.Now().UTC()

	t.Run("launch when state missing", func(t *testing.T) {
		got := local.ShouldLaunchPrune(local.PruneState{}, false, now)

		assert.True(t, got)
	})

	t.Run("launch when started_at is zero", func(t *testing.T) {
		got := local.ShouldLaunchPrune(local.PruneState{}, true, now)

		assert.True(t, got)
	})

	t.Run("do not launch during cooldown", func(t *testing.T) {
		state := local.PruneState{StartedAt: now.Add(-time.Second)}

		got := local.ShouldLaunchPrune(state, true, now)

		assert.False(t, got)
	})

	t.Run("launch after cooldown", func(t *testing.T) {
		state := local.PruneState{StartedAt: now.Add(-local.PruneLaunchCooldown).Add(-time.Second)}

		got := local.ShouldLaunchPrune(state, true, now)

		assert.True(t, got)
	})
}
