package remote_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestSourceRefresh(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	var (
		outputBuffer = &bytes.Buffer{}
		output       = internal.NewRefreshOutput(outputBuffer)
		fetcher      = &FetcherMock{FetchAllFunc: func(_ context.Context, gotOutput internal.RefreshOutput) ([]internal.Repo, error) {
			return []internal.Repo{{Name: "repo", FullName: "acme/repo"}}, nil
		}}
		sut, err = remote.New(remote.WithFetcher(fetcher))
	)
	require.NoError(t, err)

	err = sut.Refresh(context.Background(), output)

	require.NoError(t, err)
	assert.Same(t, output, fetcher.FetchAllCalls()[0].RefreshOutput)
	assert.Contains(t, outputBuffer.String(), "Refreshing remote repositories...")
	assert.Contains(t, outputBuffer.String(), "Saving remote repository cache...")
	assert.Contains(t, outputBuffer.String(), "Refreshed 1 remote repositories")
}
