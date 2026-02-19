package internal_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dependencies struct {
	initCalled    int
	refreshCalled int
	pruneCalled   int
	trackCalled   int
	queryCalled   int
	lastQuery     string
}

func TestRun(t *testing.T) {
	var (
		newDeps = func() *dependencies { return &dependencies{} }
		newSut  = func(args []string, stdout *bytes.Buffer, stderr *bytes.Buffer, deps *dependencies) func() error {
			return func() error {
				return internal.Run(args, stdout, stderr, internal.CliDependencies{
					Init: func(out io.Writer) error {
						deps.initCalled++
						return nil
					},
					Refresh: func() error {
						deps.refreshCalled++
						return nil
					},
					Prune: func() error {
						deps.pruneCalled++
						return nil
					},
					Track: func() error {
						deps.trackCalled++
						return nil
					},
					Query: func(query string, out io.Writer) error {
						deps.queryCalled++
						deps.lastQuery = query
						return nil
					},
				})
			}
		}
	)

	t.Run("routes query when no args", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 1, deps.queryCalled)
		assert.Equal(t, "", deps.lastQuery)
		assert.Empty(t, stderr.String())
	})

	t.Run("prints help", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "--help"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), internal.Usage)
		assert.Empty(t, stderr.String())
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 0, deps.queryCalled)
	})

	t.Run("routes init", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "init"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 1, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 0, deps.queryCalled)
	})

	t.Run("routes refresh", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "refresh"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 1, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 0, deps.queryCalled)
	})

	t.Run("routes prune", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "_prune"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 1, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 0, deps.queryCalled)
	})

	t.Run("routes track", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "track"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 1, deps.trackCalled)
		assert.Equal(t, 0, deps.queryCalled)
	})

	t.Run("routes query", func(t *testing.T) {
		var (
			stdout = &bytes.Buffer{}
			stderr = &bytes.Buffer{}
			deps   = newDeps()
			sut    = newSut([]string{"fly", "my", "repo"}, stdout, stderr, deps)
		)

		err := sut()

		require.NoError(t, err)
		assert.Equal(t, 0, deps.initCalled)
		assert.Equal(t, 0, deps.refreshCalled)
		assert.Equal(t, 0, deps.pruneCalled)
		assert.Equal(t, 0, deps.trackCalled)
		assert.Equal(t, 1, deps.queryCalled)
		assert.Equal(t, "my repo", deps.lastQuery)
	})
}
