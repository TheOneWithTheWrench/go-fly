package internal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {
	t.Run("skip remote when local exists", func(t *testing.T) {
		locals := []internal.Entry{{Name: "advanced-modeling-pocs", Path: "/work/advanced-modeling-pocs"}}
		remotes := []internal.Repo{{Name: "advanced-modeling-pocs", FullName: "lunarway/advanced-modeling-pocs"}}

		got := internal.Build(locals, remotes)

		require.Len(t, got, 1)
		assert.Equal(t, internal.KindLocal, got[0].Kind)
		assert.Equal(t, "advanced-modeling-pocs", got[0].Local.Name)
	})

	t.Run("include remote when no local match", func(t *testing.T) {
		locals := []internal.Entry{{Name: "alpha", Path: "/work/alpha"}}
		remotes := []internal.Repo{{Name: "beta", FullName: "acme/beta"}}

		got := internal.Build(locals, remotes)

		require.Len(t, got, 2)
		assert.Equal(t, internal.KindLocal, got[0].Kind)
		assert.Equal(t, internal.KindRemote, got[1].Kind)
	})
}
