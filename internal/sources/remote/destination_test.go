package remote_test

import (
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDestination(t *testing.T) {
	t.Run("use repo name when present", func(t *testing.T) {
		var (
			cwd = t.TempDir()
		)

		got, err := remote.Destination(internal.Repo{Name: "go-fly", FullName: "TheOneWithTheWrench/go-fly"}, cwd, false)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, "go-fly"), got)
	})

	t.Run("fallback to full name basename", func(t *testing.T) {
		var (
			cwd = t.TempDir()
		)

		got, err := remote.Destination(internal.Repo{FullName: "acme/service"}, cwd, false)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, "service"), got)
	})

	t.Run("group by owner when enabled", func(t *testing.T) {
		var (
			cwd = t.TempDir()
		)

		got, err := remote.Destination(internal.Repo{Name: "service", FullName: "acme/service"}, cwd, true)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, "acme", "service"), got)
	})
}
