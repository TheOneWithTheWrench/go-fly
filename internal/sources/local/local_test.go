package local_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/stretchr/testify/assert"
)

func TestFilterEntries(t *testing.T) {
	entries := []internal.Entry{
		{Name: "alpha", Path: "/work/alpha"},
		{Name: "bravo", Path: "/work/bravo"},
		{Name: "charlie", Path: "/work/charlie"},
	}

	t.Run("return all when query empty", func(t *testing.T) {
		got := local.FilterEntries("", entries)

		assert.Equal(t, entries, got)
	})

	t.Run("match name or path case-insensitive", func(t *testing.T) {
		got := local.FilterEntries("ALP", entries)

		assert.Len(t, got, 1)
		assert.Equal(t, "alpha", got[0].Name)
	})

	t.Run("return empty when no matches", func(t *testing.T) {
		got := local.FilterEntries("delta", entries)

		assert.Empty(t, got)
	})
}
