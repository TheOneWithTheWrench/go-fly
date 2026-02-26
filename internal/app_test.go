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
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refreshableSource struct {
	source      *SourceMock
	refreshable *RefreshableMock
}

func (s refreshableSource) Load(query string) ([]internal.Candidate, error) {
	return s.source.Load(query)
}

func (s refreshableSource) Refresh(ctx context.Context) error {
	return s.refreshable.Refresh(ctx)
}

type prunableSource struct {
	source   *SourceMock
	prunable *PrunableMock
}

func (s prunableSource) Load(query string) ([]internal.Candidate, error) {
	return s.source.Load(query)
}

func (s prunableSource) Prune() error {
	return s.prunable.Prune()
}

type trackableSource struct {
	source    *SourceMock
	trackable *TrackableMock
}

func (s trackableSource) Load(query string) ([]internal.Candidate, error) {
	return s.source.Load(query)
}

func (s trackableSource) Track(path string) error {
	return s.trackable.Track(path)
}

type localCleanSource struct {
	source  *SourceMock
	cleaner *LocalCleanerMock
}

func (s localCleanSource) Load(query string) ([]internal.Candidate, error) {
	return s.source.Load(query)
}

func (s localCleanSource) Remove(path string) error {
	return s.cleaner.Remove(path)
}

var (
	newSut = func(t *testing.T, sources []internal.Source, picker internal.Picker, cloner internal.Cloner) *internal.App {
		app, err := internal.NewApp(sources, picker, cloner)
		require.NoError(t, err)
		return app
	}
)

func localCandidate(entry internal.Entry) internal.Candidate {
	return internal.Candidate{Kind: internal.KindLocal, Local: entry}
}

func remoteCandidate(repo internal.Repo) internal.Candidate {
	return internal.Candidate{Kind: internal.KindRemote, Remote: repo}
}

func TestShouldRefresh(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	t.Run("return true when cache missing", func(t *testing.T) {
		got := remote.ShouldRefresh(remote.Cache{FetchedAt: now}, false)
		assert.True(t, got)
	})

	t.Run("return true when fetched_at is zero", func(t *testing.T) {
		got := remote.ShouldRefresh(remote.Cache{}, true)
		assert.True(t, got)
	})

	t.Run("return true when stale", func(t *testing.T) {
		got := remote.ShouldRefresh(remote.Cache{FetchedAt: now.Add(-remote.RefreshTTL - time.Minute)}, true)
		assert.True(t, got)
	})

	t.Run("return false when fresh", func(t *testing.T) {
		got := remote.ShouldRefresh(remote.Cache{FetchedAt: now.Add(-time.Minute)}, true)
		assert.False(t, got)
	})
}

func TestShouldPrune(t *testing.T) {
	var (
		now = time.Now().UTC()
	)

	t.Run("return true when state missing", func(t *testing.T) {
		got := local.ShouldPrune(local.PruneState{LastPrunedAt: now}, false)
		assert.True(t, got)
	})

	t.Run("return true when last_pruned_at is zero", func(t *testing.T) {
		got := local.ShouldPrune(local.PruneState{}, true)
		assert.True(t, got)
	})

	t.Run("return true when stale", func(t *testing.T) {
		got := local.ShouldPrune(local.PruneState{LastPrunedAt: now.Add(-local.PruneTTL - time.Minute)}, true)
		assert.True(t, got)
	})

	t.Run("return false when fresh", func(t *testing.T) {
		got := local.ShouldPrune(local.PruneState{LastPrunedAt: now.Add(-time.Minute)}, true)
		assert.False(t, got)
	})
}

func TestRefresh(t *testing.T) {
	t.Run("call refreshable sources", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			source = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			refresh = &RefreshableMock{RefreshFunc: func(ctx context.Context) error { return nil }}
		)

		sut := newSut(t, []internal.Source{refreshableSource{source: source, refreshable: refresh}}, picker, cloner)

		err := sut.Refresh(context.Background())

		require.NoError(t, err)
		assert.Len(t, refresh.RefreshCalls(), 1)
	})
}

func TestQuery(t *testing.T) {
	t.Run("remove missing local selection", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			local  = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: "/work/alpha"})}, nil
			}}
			remote = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			cleaner = &LocalCleanerMock{RemoveFunc: func(path string) error { return nil }}
		)

		app, err := internal.NewApp([]internal.Source{
			localCleanSource{source: local, cleaner: cleaner},
			remote,
		}, picker, cloner)
		require.NoError(t, err)

		err = app.Query("alpha", &bytes.Buffer{})

		assert.Error(t, err)
		removeCalls := cleaner.RemoveCalls()
		require.Len(t, removeCalls, 1)
		assert.Equal(t, "/work/alpha", removeCalls[0].S)
	})

	t.Run("return error when no repos", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			local  = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			remote = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
		)

		sut := newSut(t, []internal.Source{local, remote}, picker, cloner)

		err := sut.Query("", &bytes.Buffer{})

		assert.ErrorIs(t, err, internal.ErrNoReposTracked)
	})

	t.Run("return error when no matches", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			local  = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
			remote = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
		)

		sut := newSut(t, []internal.Source{local, remote}, picker, cloner)

		err := sut.Query("zzz", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("return error when clone fails", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			local  = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
			remote = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
			}}
		)

		picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}
		cloner.CloneFunc = func(repo internal.Repo) (string, error) {
			return "", errors.New("boom")
		}

		sut := newSut(t, []internal.Source{local, remote}, picker, cloner)

		err := sut.Query("beta", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("skip picker when single local match", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		local := &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
			return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: valid})}, nil
		}}
		remote := &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
			return nil, internal.ErrNoReposTracked
		}}

		picker := &PickerMock{PickFunc: func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			t.Fatalf("picker should not be called")
			return internal.Candidate{}, false, nil
		}}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{local, remote}, picker, &ClonerMock{})

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

		local := &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
			return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: valid})}, nil
		}}
		remote := &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
			return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
		}}

		picker := &PickerMock{PickFunc: func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{local, remote}, picker, &ClonerMock{})

		err := sut.Query("alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("clone remote selection", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			local  = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			remote = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
			}}
			trackable = &TrackableMock{TrackFunc: func(path string) error { return nil }}
		)

		picker.PickFunc = func(query string, candidates []internal.Candidate) (internal.Candidate, bool, error) {
			return candidates[0], true, nil
		}
		cloner.CloneFunc = func(repo internal.Repo) (string, error) {
			return "/work/beta", nil
		}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{
			local,
			remote,
			trackableSource{source: &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}, trackable: trackable},
		}, picker, cloner)

		err := sut.Query("beta", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), "/work/beta")
		trackCalls := trackable.TrackCalls()
		require.Len(t, trackCalls, 1)
		assert.Equal(t, "/work/beta", trackCalls[0].S)
	})
}

func TestPrune(t *testing.T) {
	t.Run("call prunable sources", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			cloner = &ClonerMock{}
			source = &SourceMock{LoadFunc: func(query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			prunable = &PrunableMock{PruneFunc: func() error { return nil }}
		)

		sut := newSut(t, []internal.Source{prunableSource{source: source, prunable: prunable}}, picker, cloner)

		err := sut.Prune()

		require.NoError(t, err)
		assert.Len(t, prunable.PruneCalls(), 1)
	})
}
