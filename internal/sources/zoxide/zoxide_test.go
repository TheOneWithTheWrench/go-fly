package zoxide_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/zoxide"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceLoad(t *testing.T) {
	t.Run("return no repos tracked when zoxide has no entries", func(t *testing.T) {
		var (
			sut, err = zoxide.New(zoxide.WithLister(&ListerMock{ListFunc: func(context.Context) ([]zoxide.Match, error) {
				return nil, nil
			}}))
		)
		require.NoError(t, err)

		_, loadErr := sut.Load(context.Background(), "")

		assert.ErrorIs(t, loadErr, internal.ErrNoReposTracked)
	})

	t.Run("filter and map zoxide matches to local candidates", func(t *testing.T) {
		var (
			root      = t.TempDir()
			alphaPath = filepath.Join(root, "alpha")
			betaPath  = filepath.Join(root, "beta")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(alphaPath, ".git"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(betaPath, ".git"), 0o755))

		var (
			sut, err = zoxide.New(zoxide.WithLister(&ListerMock{ListFunc: func(context.Context) ([]zoxide.Match, error) {
				return []zoxide.Match{{Path: alphaPath, Score: 20}, {Path: betaPath, Score: 10}}, nil
			}}))
		)
		require.NoError(t, err)

		result, loadErr := sut.Load(context.Background(), "alp")

		require.NoError(t, loadErr)
		require.Len(t, result, 1)
		assert.Equal(t, alphaPath, result[0].Meta[internal.CandidateMetaPath])
		assert.Equal(t, 20.0, result[0].Signals[internal.CandidateSignalZoxideScore])
		assert.Equal(t, internal.CandidateSourceZoxide, result[0].Meta[internal.CandidateMetaSource])
		assert.Contains(t, result[0].Meta[internal.CandidateMetaLabel], "[zoxide]")
	})

	t.Run("skip non git directories", func(t *testing.T) {
		var (
			root    = t.TempDir()
			invalid = filepath.Join(root, "plain-dir")
		)
		require.NoError(t, os.MkdirAll(invalid, 0o755))

		var (
			sut, err = zoxide.New(zoxide.WithLister(&ListerMock{ListFunc: func(context.Context) ([]zoxide.Match, error) {
				return []zoxide.Match{{Path: invalid, Score: 42}}, nil
			}}))
		)
		require.NoError(t, err)

		result, loadErr := sut.Load(context.Background(), "")

		require.NoError(t, loadErr)
		assert.Empty(t, result)
	})

	t.Run("propagate lister errors", func(t *testing.T) {
		var (
			expectedErr = errors.New("boom")
			sut, err    = zoxide.New(zoxide.WithLister(&ListerMock{ListFunc: func(context.Context) ([]zoxide.Match, error) {
				return nil, expectedErr
			}}))
		)
		require.NoError(t, err)

		_, loadErr := sut.Load(context.Background(), "")

		assert.ErrorIs(t, loadErr, expectedErr)
	})
}

func TestSourceRefresh(t *testing.T) {
	t.Run("use lister refresh when available", func(t *testing.T) {
		var (
			listCalled    = false
			refreshCalled = false
			lister        = &refreshingLister{
				ListFunc: func(context.Context) ([]zoxide.Match, error) {
					listCalled = true
					return nil, nil
				},
				RefreshFunc: func(context.Context) error {
					refreshCalled = true
					return nil
				},
			}
		)

		sut, err := zoxide.New(zoxide.WithLister(lister))
		require.NoError(t, err)

		err = sut.Refresh(context.Background())

		require.NoError(t, err)
		assert.True(t, refreshCalled)
		assert.False(t, listCalled)
	})

	t.Run("fallback to list when lister does not implement refresh", func(t *testing.T) {
		var (
			listCalled = false
			lister     = &ListerMock{ListFunc: func(context.Context) ([]zoxide.Match, error) {
				listCalled = true
				return nil, nil
			}}
		)

		sut, err := zoxide.New(zoxide.WithLister(lister))
		require.NoError(t, err)

		err = sut.Refresh(context.Background())

		require.NoError(t, err)
		assert.True(t, listCalled)
	})

	t.Run("propagate refresh errors", func(t *testing.T) {
		var (
			expectedErr = errors.New("refresh failed")
			lister      = &refreshingLister{RefreshFunc: func(context.Context) error {
				return expectedErr
			}}
		)

		sut, err := zoxide.New(zoxide.WithLister(lister))
		require.NoError(t, err)

		err = sut.Refresh(context.Background())

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCommandLister(t *testing.T) {
	t.Run("parse zoxide command output", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				assert.Equal(t, "zoxide", name)
				assert.Equal(t, []string{"query", "--list", "--score"}, args)
				return []byte("99 /tmp/repo\n1.5 /tmp/with space\ninvalid line\n"), nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		result, listErr := sut.List(context.Background())

		require.NoError(t, listErr)
		require.Len(t, result, 2)
		assert.Equal(t, 99.0, result[0].Score)
		assert.Equal(t, "/tmp/repo", result[0].Path)
		assert.Equal(t, 1.5, result[1].Score)
		assert.Equal(t, "/tmp/with space", result[1].Path)
	})

	t.Run("fallback to z shell function when zoxide binary missing", func(t *testing.T) {
		t.Setenv("SHELL", "zsh")
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			calls  = 0
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				calls++
				if calls == 1 {
					assert.Equal(t, "zoxide", name)
					assert.Equal(t, []string{"query", "--list", "--score"}, args)
					return nil, &exec.Error{Name: "zoxide", Err: exec.ErrNotFound}
				}
				if calls == 2 {
					assert.Equal(t, "z", name)
					assert.Equal(t, []string{"-l"}, args)
					return nil, &exec.Error{Name: "z", Err: exec.ErrNotFound}
				}

				assert.Equal(t, "zsh", name)
				assert.Equal(t, []string{"-ic", "z -l"}, args)
				return []byte("10 /tmp/repo\n"), nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		result, listErr := sut.List(context.Background())

		require.NoError(t, listErr)
		require.Len(t, result, 1)
		assert.Equal(t, "/tmp/repo", result[0].Path)
		assert.Equal(t, 10.0, result[0].Score)
		assert.Equal(t, 3, calls)
	})

	t.Run("fallback to direct z command when zoxide binary missing", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				if name == "zoxide" {
					assert.Equal(t, []string{"query", "--list", "--score"}, args)
					return nil, &exec.Error{Name: "zoxide", Err: exec.ErrNotFound}
				}

				assert.Equal(t, "z", name)
				assert.Equal(t, []string{"-l"}, args)
				return []byte("/tmp/repo\n/tmp/second\n"), nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		result, listErr := sut.List(context.Background())

		require.NoError(t, listErr)
		require.Len(t, result, 2)
		assert.Equal(t, "/tmp/repo", result[0].Path)
		assert.Equal(t, 2.0, result[0].Score)
		assert.Equal(t, "/tmp/second", result[1].Path)
		assert.Equal(t, 1.0, result[1].Score)
	})

	t.Run("do not fallback on non not-found errors", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				return nil, errors.New("permission denied")
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		_, listErr := sut.List(context.Background())

		require.Error(t, listErr)
		assert.Contains(t, listErr.Error(), "permission denied")
	})

	t.Run("use fresh cache without executing commands", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			store, err = zoxide.NewStore()
		)
		require.NoError(t, err)

		err = store.Save(zoxide.Cache{
			FetchedAt: time.Now().UTC(),
			Backend:   "shell",
			Matches:   []zoxide.Match{{Path: "/tmp/repo", Score: 10}},
		})
		require.NoError(t, err)

		var (
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				t.Fatalf("unexpected command execution: %s %v", name, args)
				return nil, nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		result, listErr := sut.List(context.Background())

		require.NoError(t, listErr)
		require.Len(t, result, 1)
		assert.Equal(t, "/tmp/repo", result[0].Path)
		assert.Equal(t, 10.0, result[0].Score)
	})

	t.Run("cache only git repos when refreshing stale entries", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			root       = t.TempDir()
			gitPath    = filepath.Join(root, "repo")
			plain      = filepath.Join(root, "plain")
			store, err = zoxide.NewStore()
		)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Join(gitPath, ".git"), 0o755))
		require.NoError(t, os.MkdirAll(plain, 0o755))
		require.NoError(t, store.Save(zoxide.Cache{
			FetchedAt: time.Now().Add(-zoxide.RefreshTTL).Add(-time.Minute).UTC(),
			Backend:   "zoxide",
			Matches:   []zoxide.Match{{Path: "/tmp/old", Score: 1}},
		}))

		var (
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				assert.Equal(t, "zoxide", name)
				assert.Equal(t, []string{"query", "--list", "--score"}, args)
				return []byte("10 " + gitPath + "\n9 " + plain + "\n"), nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		result, listErr := sut.List(context.Background())

		require.NoError(t, listErr)
		require.Len(t, result, 2)

		cache, exists, err := store.Load()
		require.NoError(t, err)
		assert.True(t, exists)
		require.Len(t, cache.Matches, 1)
		assert.Equal(t, gitPath, cache.Matches[0].Path)
		assert.Equal(t, 10.0, cache.Matches[0].Score)
		assert.Equal(t, "zoxide", cache.Backend)
		assert.False(t, cache.FetchedAt.IsZero())
	})

	t.Run("refresh updates cache even when cache is fresh", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", t.TempDir())

		var (
			root       = t.TempDir()
			gitPath    = filepath.Join(root, "repo")
			plain      = filepath.Join(root, "plain")
			store, err = zoxide.NewStore()
		)
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(filepath.Join(gitPath, ".git"), 0o755))
		require.NoError(t, os.MkdirAll(plain, 0o755))
		require.NoError(t, store.Save(zoxide.Cache{
			FetchedAt: time.Now().UTC(),
			Backend:   "zoxide",
			Matches:   []zoxide.Match{{Path: "/tmp/old", Score: 1}},
		}))

		var (
			calls  = 0
			runner = internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				calls++
				assert.Equal(t, "zoxide", name)
				assert.Equal(t, []string{"query", "--list", "--score"}, args)
				return []byte("10 " + gitPath + "\n9 " + plain + "\n"), nil
			})
		)

		sut, err := zoxide.NewCommandLister(runner)
		require.NoError(t, err)

		err = sut.Refresh(context.Background())

		require.NoError(t, err)
		assert.Equal(t, 1, calls)

		cache, exists, err := store.Load()
		require.NoError(t, err)
		assert.True(t, exists)
		require.Len(t, cache.Matches, 1)
		assert.Equal(t, gitPath, cache.Matches[0].Path)
		assert.Equal(t, 10.0, cache.Matches[0].Score)
	})
}

type refreshingLister struct {
	ListFunc    func(context.Context) ([]zoxide.Match, error)
	RefreshFunc func(context.Context) error
}

func (l *refreshingLister) List(ctx context.Context) ([]zoxide.Match, error) {
	if l.ListFunc == nil {
		return nil, nil
	}

	return l.ListFunc(ctx)
}

func (l *refreshingLister) Refresh(ctx context.Context) error {
	if l.RefreshFunc == nil {
		return nil
	}

	return l.RefreshFunc(ctx)
}
