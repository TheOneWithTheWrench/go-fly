package internal_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshOutput(t *testing.T) {
	t.Run("write status updates as lines when output is not a terminal", func(t *testing.T) {
		var (
			output = &bytes.Buffer{}
			sut    = internal.NewRefreshOutput(output)
		)

		require.NoError(t, sut.SetStatus("Fetching organizations..."))
		require.NoError(t, sut.SetStatus("Fetching repositories..."))
		require.NoError(t, sut.ClearStatus())
		_, err := fmt.Fprintln(sut, "refresh complete")

		require.NoError(t, err)
		assert.Equal(t, "Fetching organizations...\nFetching repositories...\nrefresh complete\n", output.String())
	})
}
