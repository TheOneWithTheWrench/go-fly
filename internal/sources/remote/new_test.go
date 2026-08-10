package remote_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("create source with defaults", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		sut, err := remote.New()

		require.NoError(t, err)
		require.NotNil(t, sut)
	})
}

func TestNewOptions(t *testing.T) {
	t.Run("return error when with fetcher option has nil fetcher", func(t *testing.T) {
		_, err := remote.New(remote.WithFetcher(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "remote fetcher required")
	})

	t.Run("return error when with cloner option has nil cloner", func(t *testing.T) {
		_, err := remote.New(remote.WithCloner(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cloner required")
	})

	t.Run("return error when custom cloner and clone settings are combined", func(t *testing.T) {
		_, err := remote.New(
			remote.WithCloner(internal.ClonerFunc(func(repo internal.Repo) (string, error) { return "", nil })),
			remote.WithCloneBaseDir("/tmp"),
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("return error when custom cloner and owner grouping are combined", func(t *testing.T) {
		_, err := remote.New(
			remote.WithCloner(internal.ClonerFunc(func(repo internal.Repo) (string, error) { return "", nil })),
			remote.WithCloneGroupByOwner(),
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("return error when with runner option has nil runner", func(t *testing.T) {
		_, err := remote.New(remote.WithRunner(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "runner required")
	})

	t.Run("return error when with refresh launcher option has nil refresher", func(t *testing.T) {
		_, err := remote.New(remote.WithRefreshLauncher(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "refresh launcher required")
	})

	t.Run("use provided cloner in resolve", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			fetcher = &FetcherMock{FetchAllFunc: func(context.Context, internal.RefreshOutput) ([]internal.Repo, error) {
				return nil, nil
			}}
			called   = false
			sut, err = remote.New(
				remote.WithFetcher(fetcher),
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

	t.Run("launch refresh only once while stale", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			store, err = remote.NewRemoteStore()
		)
		require.NoError(t, err)
		require.NoError(t, store.Save(remote.Cache{
			FetchedAt: time.Now().Add(-remote.RefreshTTL).Add(-time.Minute),
			Repos: []internal.Repo{
				{Name: "repo", FullName: "acme/repo", SSHURL: "git@github.com:acme/repo.git"},
			},
		}))

		var launches int
		sut, err := remote.New(
			remote.WithFetcher(&FetcherMock{FetchAllFunc: func(context.Context, internal.RefreshOutput) ([]internal.Repo, error) {
				return nil, nil
			}}),
			remote.WithRefreshLauncher(internal.RefresherFunc(func() {
				launches++
			})),
		)
		require.NoError(t, err)

		_, err = sut.Load(context.Background(), "")
		require.NoError(t, err)
		_, err = sut.Load(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, 1, launches)
	})

	t.Run("launch refresh once across multiple source instances", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		store, err := remote.NewRemoteStore()
		require.NoError(t, err)
		require.NoError(t, store.Save(remote.Cache{
			FetchedAt: time.Now().Add(-remote.RefreshTTL).Add(-time.Minute),
			Repos: []internal.Repo{
				{Name: "repo", FullName: "acme/repo", SSHURL: "git@github.com:acme/repo.git"},
			},
		}))

		var launches int
		newSource := func(t *testing.T) *remote.Source {
			t.Helper()

			sut, newErr := remote.New(
				remote.WithFetcher(&FetcherMock{FetchAllFunc: func(context.Context, internal.RefreshOutput) ([]internal.Repo, error) {
					return nil, nil
				}}),
				remote.WithRefreshLauncher(internal.RefresherFunc(func() {
					launches++
				})),
			)
			require.NoError(t, newErr)

			return sut
		}

		sut1 := newSource(t)
		sut2 := newSource(t)

		_, err = sut1.Load(context.Background(), "")
		require.NoError(t, err)
		_, err = sut2.Load(context.Background(), "")
		require.NoError(t, err)

		assert.Equal(t, 1, launches)
	})

	t.Run("launch refresh when cache missing", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var launches int
		sut, err := remote.New(
			remote.WithFetcher(&FetcherMock{FetchAllFunc: func(context.Context, internal.RefreshOutput) ([]internal.Repo, error) {
				return nil, nil
			}}),
			remote.WithRefreshLauncher(internal.RefresherFunc(func() {
				launches++
			})),
		)
		require.NoError(t, err)

		_, err = sut.Load(context.Background(), "repo")
		assert.ErrorIs(t, err, internal.ErrNoReposTracked)
		_, err = sut.Load(context.Background(), "repo")
		assert.ErrorIs(t, err, internal.ErrNoReposTracked)

		assert.Equal(t, 1, launches)
	})

	t.Run("propagate cache load errors", func(t *testing.T) {
		cacheHome := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", cacheHome)

		cacheDir := filepath.Join(cacheHome, "fly")
		require.NoError(t, os.MkdirAll(cacheDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(cacheDir, "remote.json"), []byte("{"), 0o644))

		sut, err := remote.New(
			remote.WithFetcher(&FetcherMock{FetchAllFunc: func(context.Context, internal.RefreshOutput) ([]internal.Repo, error) {
				return nil, nil
			}}),
		)
		require.NoError(t, err)

		_, err = sut.Load(context.Background(), "repo")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "load cache")
	})
}
