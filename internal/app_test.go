package internal_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type appDependencies struct {
	localStore      *IndexStorageMock
	remoteStore     *RemoteStorageMock
	remoteFetcher   *RemoteFetcherMock
	picker          *PickerMock
	refreshLauncher *RefreshLauncherMock
	pruneStore      *PruneStateStorageMock
	pruneLauncher   *PruneLauncherMock
	cloner          *ClonerMock
}

var (
	newDefaultDependencies = func() appDependencies {
		deps := appDependencies{
			localStore:      &IndexStorageMock{},
			remoteStore:     &RemoteStorageMock{},
			remoteFetcher:   &RemoteFetcherMock{},
			picker:          &PickerMock{},
			refreshLauncher: &RefreshLauncherMock{},
			pruneStore:      &PruneStateStorageMock{},
			pruneLauncher:   &PruneLauncherMock{},
			cloner:          &ClonerMock{},
		}

		deps.localStore.LoadFunc = func() ([]internal.Entry, error) { return []internal.Entry{}, nil }
		deps.localStore.SaveFunc = func(entries []internal.Entry) error { return nil }
		deps.localStore.UpsertFunc = func(entry internal.Entry) error { return nil }
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) { return internal.Cache{}, true, nil }
		deps.remoteStore.SaveFunc = func(cache internal.Cache) error { return nil }
		deps.remoteFetcher.FetchAllFunc = func(ctx context.Context) ([]internal.Repo, error) { return nil, nil }
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return internal.Candidate{}, false, nil
		}
		deps.refreshLauncher.LaunchFunc = func() {}
		deps.pruneStore.LoadFunc = func() (internal.PruneState, bool, error) { return internal.PruneState{}, true, nil }
		deps.pruneStore.SaveFunc = func(state internal.PruneState) error { return nil }
		deps.pruneLauncher.LaunchFunc = func() {}
		deps.cloner.CloneFunc = func(repo internal.Repo) (string, error) { return "", nil }

		return deps
	}
	newSut = func(t *testing.T, deps appDependencies) *internal.App {
		app, err := internal.NewApp(
			deps.localStore,
			deps.remoteStore,
			deps.remoteFetcher,
			deps.picker,
			deps.refreshLauncher,
			deps.pruneStore,
			deps.pruneLauncher,
			deps.cloner,
		)
		require.NoError(t, err)
		return app
	}
)

func TestShouldRefresh(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	t.Run("return true when cache missing", func(t *testing.T) {
		got := internal.ShouldRefresh(internal.Cache{FetchedAt: now}, false)
		assert.True(t, got)
	})

	t.Run("return true when fetched_at is zero", func(t *testing.T) {
		got := internal.ShouldRefresh(internal.Cache{}, true)
		assert.True(t, got)
	})

	t.Run("return true when stale", func(t *testing.T) {
		got := internal.ShouldRefresh(internal.Cache{FetchedAt: now.Add(-internal.RefreshTTL - time.Minute)}, true)
		assert.True(t, got)
	})

	t.Run("return false when fresh", func(t *testing.T) {
		got := internal.ShouldRefresh(internal.Cache{FetchedAt: now.Add(-time.Minute)}, true)
		assert.False(t, got)
	})
}

func TestShouldPrune(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	t.Run("return true when state missing", func(t *testing.T) {
		got := internal.ShouldPrune(internal.PruneState{LastPrunedAt: now}, false)
		assert.True(t, got)
	})

	t.Run("return true when last_pruned_at is zero", func(t *testing.T) {
		got := internal.ShouldPrune(internal.PruneState{}, true)
		assert.True(t, got)
	})

	t.Run("return true when stale", func(t *testing.T) {
		got := internal.ShouldPrune(internal.PruneState{LastPrunedAt: now.Add(-internal.PruneTTL - time.Minute)}, true)
		assert.True(t, got)
	})

	t.Run("return false when fresh", func(t *testing.T) {
		got := internal.ShouldPrune(internal.PruneState{LastPrunedAt: now.Add(-time.Minute)}, true)
		assert.False(t, got)
	})
}

func TestRefresh(t *testing.T) {
	t.Run("fetch and save cache", func(t *testing.T) {
		var (
			deps   = newDefaultDependencies()
			appErr error
		)

		deps.remoteFetcher.FetchAllFunc = func(ctx context.Context) ([]internal.Repo, error) {
			return []internal.Repo{{Name: "repo", FullName: "acme/repo"}}, nil
		}
		deps.remoteStore.SaveFunc = func(cache internal.Cache) error {
			assert.Len(t, cache.Repos, 1)
			assert.False(t, cache.FetchedAt.IsZero())
			return nil
		}

		sut := newSut(t, deps)

		appErr = sut.Refresh(context.Background())

		require.NoError(t, appErr)
	})
}

func TestQuery(t *testing.T) {
	t.Run("remove missing local selection", func(t *testing.T) {
		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) {
			return []internal.Entry{{Name: "alpha", Path: "/work/alpha"}}, nil
		}
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) {
			return internal.Cache{}, true, nil
		}
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}
		deps.localStore.SaveFunc = func(entries []internal.Entry) error {
			require.Len(t, entries, 0)
			return nil
		}

		sut := newSut(t, deps)

		err := sut.Query("alpha", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("return error when no repos", func(t *testing.T) {
		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) { return []internal.Entry{}, nil }
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) { return internal.Cache{}, false, nil }

		sut := newSut(t, deps)

		err := sut.Query("", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("return error when no matches", func(t *testing.T) {
		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) { return []internal.Entry{{Name: "alpha", Path: "/work/alpha"}}, nil }
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) { return internal.Cache{}, true, nil }

		sut := newSut(t, deps)

		err := sut.Query("zzz", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("return error when clone fails", func(t *testing.T) {
		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) { return []internal.Entry{}, nil }
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) {
			return internal.Cache{Repos: []internal.Repo{{Name: "beta", FullName: "acme/beta"}}}, true, nil
		}
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}
		deps.cloner.CloneFunc = func(repo internal.Repo) (string, error) {
			return "", errors.New("boom")
		}

		sut := newSut(t, deps)

		err := sut.Query("beta", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("skip picker when single local match", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) {
			return []internal.Entry{{Name: "alpha", Path: valid}}, nil
		}
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) { return internal.Cache{}, true, nil }
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			t.Fatalf("picker should not be called")
			return internal.Candidate{}, false, nil
		}

		out := &bytes.Buffer{}
		sut := newSut(t, deps)

		err := sut.Query("alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("print local selection", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) {
			return []internal.Entry{{Name: "alpha", Path: valid}}, nil
		}
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) { return internal.Cache{}, true, nil }
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}

		out := &bytes.Buffer{}
		sut := newSut(t, deps)

		err := sut.Query("alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("clone remote selection", func(t *testing.T) {
		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) { return []internal.Entry{}, nil }
		deps.remoteStore.LoadFunc = func() (internal.Cache, bool, error) {
			return internal.Cache{Repos: []internal.Repo{{Name: "beta", FullName: "acme/beta"}}}, true, nil
		}
		deps.picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}
		deps.cloner.CloneFunc = func(repo internal.Repo) (string, error) {
			return "/work/beta", nil
		}
		deps.localStore.UpsertFunc = func(entry internal.Entry) error {
			assert.Equal(t, "/work/beta", entry.Path)
			return nil
		}

		out := &bytes.Buffer{}
		sut := newSut(t, deps)

		err := sut.Query("beta", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), "/work/beta")
	})
}

func TestPrune(t *testing.T) {
	t.Run("remove missing and non-repo entries", func(t *testing.T) {
		var (
			root    = t.TempDir()
			valid   = filepath.Join(root, "valid")
			invalid = filepath.Join(root, "invalid")
			missing = filepath.Join(root, "missing")
		)

		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))
		require.NoError(t, os.MkdirAll(invalid, 0o755))

		deps := newDefaultDependencies()
		deps.localStore.LoadFunc = func() ([]internal.Entry, error) {
			return []internal.Entry{
				{Name: "valid", Path: valid},
				{Name: "invalid", Path: invalid},
				{Name: "missing", Path: missing},
			}, nil
		}
		deps.localStore.SaveFunc = func(entries []internal.Entry) error {
			require.Len(t, entries, 1)
			assert.Equal(t, valid, entries[0].Path)
			return nil
		}
		deps.pruneStore.SaveFunc = func(state internal.PruneState) error {
			assert.False(t, state.LastPrunedAt.IsZero())
			return nil
		}

		sut := newSut(t, deps)

		err := sut.Prune()

		require.NoError(t, err)
	})
}
