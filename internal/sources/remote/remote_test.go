package remote_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
)

func TestFilterRepos(t *testing.T) {
	repos := []internal.Repo{
		{Name: "alpha", FullName: "acme/alpha"},
		{Name: "bravo", FullName: "acme/bravo"},
		{Name: "charlie", FullName: "core/charlie"},
	}

	t.Run("return all when query empty", func(t *testing.T) {
		got := remote.FilterRepos("", repos)

		assert.Equal(t, repos, got)
	})

	t.Run("match name or full name case-insensitive", func(t *testing.T) {
		got := remote.FilterRepos("CORE", repos)

		assert.Len(t, got, 1)
		assert.Equal(t, "charlie", got[0].Name)
	})

	t.Run("return empty when no matches", func(t *testing.T) {
		got := remote.FilterRepos("delta", repos)

		assert.Empty(t, got)
	})
}
