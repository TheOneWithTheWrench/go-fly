package remote_test

import (
	"context"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("return error when fetcher is missing", func(t *testing.T) {
		_, err := remote.New(nil, internal.RefresherFunc(func() {}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "remote fetcher required")
	})

	t.Run("return error when refresher is missing", func(t *testing.T) {
		fetcher := &FetcherMock{FetchAllFunc: func(context.Context) ([]internal.Repo, error) {
			return nil, nil
		}}

		_, err := remote.New(fetcher, nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "refresh launcher required")
	})

	t.Run("create source with default cloner", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())
		fetcher := &FetcherMock{FetchAllFunc: func(context.Context) ([]internal.Repo, error) {
			return nil, nil
		}}

		sut, err := remote.New(fetcher, internal.RefresherFunc(func() {}))

		require.NoError(t, err)
		require.NotNil(t, sut)
	})
}

func TestNewOptions(t *testing.T) {
	t.Run("return error when with cloner option has nil cloner", func(t *testing.T) {
		fetcher := &FetcherMock{FetchAllFunc: func(context.Context) ([]internal.Repo, error) {
			return nil, nil
		}}

		_, err := remote.New(fetcher, internal.RefresherFunc(func() {}), remote.WithCloner(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cloner required")
	})

	t.Run("use provided cloner in resolve", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			fetcher = &FetcherMock{FetchAllFunc: func(context.Context) ([]internal.Repo, error) {
				return nil, nil
			}}
			called   = false
			sut, err = remote.New(
				fetcher,
				internal.RefresherFunc(func() {}),
				remote.WithCloner(internal.ClonerFunc(func(repo internal.Repo) (string, error) {
					called = true
					assert.Equal(t, "acme/repo", repo.FullName)
					return "/tmp/repo", nil
				})),
			)
		)
		require.NoError(t, err)

		path, resolveErr := sut.Resolve(internal.Candidate{Meta: map[string]string{
			internal.CandidateMetaSource:   internal.CandidateSourceRemote,
			internal.CandidateMetaFullName: "acme/repo",
		}})

		require.NoError(t, resolveErr)
		assert.True(t, called)
		assert.Equal(t, "/tmp/repo", path)
	})
}
