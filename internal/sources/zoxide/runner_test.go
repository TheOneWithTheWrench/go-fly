package zoxide

import (
	"context"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultRunner(t *testing.T) {
	newSut := func() internal.Runner { return defaultRunner() }

	t.Run("keep stderr out of the parsed output", func(t *testing.T) {
		sut := newSut()

		got, err := sut.Run(context.Background(), "/bin/sh", "-c", "echo warning >&2; echo 42 /tmp")

		require.NoError(t, err)
		assert.Equal(t, "42 /tmp\n", string(got))
	})

	t.Run("report stderr when the command fails", func(t *testing.T) {
		sut := newSut()

		_, err := sut.Run(context.Background(), "/bin/sh", "-c", "echo boom >&2; exit 3")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})

	t.Run("run the shell backend detached from the terminal", func(t *testing.T) {
		sut := newSut()

		got, err := sut.Run(context.Background(), "/bin/sh", "-c", "printf %s \"$"+internal.ChildEnvVar+"\"")

		require.NoError(t, err)
		assert.Equal(t, "1", string(got), "backend must run through internal.NewIsolatedCommand")
	})

	t.Run("report a missing executable as command not found", func(t *testing.T) {
		sut := newSut()

		_, err := sut.Run(context.Background(), "fly-no-such-backend-binary")

		require.Error(t, err)
		assert.True(t, isCommandNotFound(err), "fallback to the next backend relies on this: %v", err)
	})
}
