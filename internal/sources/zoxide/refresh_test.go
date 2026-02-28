package zoxide_test

import (
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal/sources/zoxide"
	"github.com/stretchr/testify/assert"
)

func TestShouldRefresh(t *testing.T) {
	t.Run("refresh when cache missing", func(t *testing.T) {
		got := zoxide.ShouldRefresh(zoxide.Cache{}, false)

		assert.True(t, got)
	})

	t.Run("refresh when fetched time is zero", func(t *testing.T) {
		got := zoxide.ShouldRefresh(zoxide.Cache{}, true)

		assert.True(t, got)
	})

	t.Run("do not refresh when cache is fresh", func(t *testing.T) {
		cache := zoxide.Cache{FetchedAt: time.Now().Add(-time.Minute)}

		got := zoxide.ShouldRefresh(cache, true)

		assert.False(t, got)
	})

	t.Run("refresh when cache is stale", func(t *testing.T) {
		cache := zoxide.Cache{FetchedAt: time.Now().Add(-zoxide.RefreshTTL).Add(-time.Second)}

		got := zoxide.ShouldRefresh(cache, true)

		assert.True(t, got)
	})
}
