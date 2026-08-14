package internal

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTerminalWriter(t *testing.T) {
	newSut := func() func(writer io.Writer) bool { return isTerminalWriter }

	t.Run("reject os.DevNull even though it is a character device", func(t *testing.T) {
		var (
			sut          = newSut()
			devNull, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		)
		require.NoError(t, err)
		t.Cleanup(func() { _ = devNull.Close() })

		got := sut(devNull)

		assert.False(t, got)
	})

	t.Run("reject a regular file", func(t *testing.T) {
		var (
			sut       = newSut()
			file, err = os.CreateTemp(t.TempDir(), "out")
		)
		require.NoError(t, err)
		t.Cleanup(func() { _ = file.Close() })

		got := sut(file)

		assert.False(t, got)
	})

	t.Run("reject a writer that is not a file", func(t *testing.T) {
		sut := newSut()

		got := sut(&bytes.Buffer{})

		assert.False(t, got)
	})
}
