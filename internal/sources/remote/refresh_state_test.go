package remote_test

import (
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
)

func TestShouldLaunchRefresh(t *testing.T) {
	now := time.Now().UTC()

	t.Run("launch when state missing", func(t *testing.T) {
		got := remote.ShouldLaunchRefresh(remote.RefreshState{}, false, now)

		assert.True(t, got)
	})

	t.Run("launch when started_at is zero", func(t *testing.T) {
		got := remote.ShouldLaunchRefresh(remote.RefreshState{}, true, now)

		assert.True(t, got)
	})

	t.Run("do not launch during cooldown", func(t *testing.T) {
		state := remote.RefreshState{StartedAt: now.Add(-time.Second)}

		got := remote.ShouldLaunchRefresh(state, true, now)

		assert.False(t, got)
	})

	t.Run("launch after cooldown", func(t *testing.T) {
		state := remote.RefreshState{StartedAt: now.Add(-remote.RefreshLaunchCooldown).Add(-time.Second)}

		got := remote.ShouldLaunchRefresh(state, true, now)

		assert.True(t, got)
	})
}
