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

func (s refreshableSource) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	return s.source.Load(ctx, query)
}

func (s refreshableSource) Resolve(candidate internal.Candidate) (string, error) {
	return "", internal.ErrUnsupportedCandidate
}

func (s refreshableSource) Refresh(ctx context.Context) error {
	return s.refreshable.Refresh(ctx)
}

type prunableSource struct {
	source   *SourceMock
	prunable *PrunableMock
}

func (s prunableSource) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	return s.source.Load(ctx, query)
}

func (s prunableSource) Resolve(candidate internal.Candidate) (string, error) {
	return "", internal.ErrUnsupportedCandidate
}

func (s prunableSource) Prune() error {
	return s.prunable.Prune()
}

type trackableSource struct {
	source    *SourceMock
	trackable *TrackableMock
}

func (s trackableSource) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	return s.source.Load(ctx, query)
}

func (s trackableSource) Resolve(candidate internal.Candidate) (string, error) {
	return "", internal.ErrUnsupportedCandidate
}

func (s trackableSource) Track(path string) error {
	return s.trackable.Track(path)
}

type resolverSource struct {
	source      *SourceMock
	resolveFunc func(internal.Candidate) (string, error)
}

func (s resolverSource) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	return s.source.Load(ctx, query)
}

func (s resolverSource) Resolve(candidate internal.Candidate) (string, error) {
	if s.resolveFunc == nil {
		return "", internal.ErrUnsupportedCandidate
	}

	return s.resolveFunc(candidate)
}

var (
	newSut = func(t *testing.T, sources []internal.Source, picker internal.Picker) *internal.App {
		app, err := internal.NewApp(sources, internal.WithPicker(picker))
		require.NoError(t, err)
		return app
	}
)

func TestNewApp(t *testing.T) {
	t.Run("return error when with picker option has nil picker", func(t *testing.T) {
		_, err := internal.NewApp([]internal.Source{&SourceMock{}}, internal.WithPicker(nil))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "picker required")
	})

	t.Run("create app with default picker", func(t *testing.T) {
		sut, err := internal.NewApp([]internal.Source{&SourceMock{}})

		require.NoError(t, err)
		require.NotNil(t, sut)
	})
}

func localCandidate(entry internal.Entry) internal.Candidate {
	return internal.Candidate{
		Meta: map[string]string{
			internal.CandidateMetaSource: internal.CandidateSourceLocal,
			internal.CandidateMetaName:   entry.Name,
			internal.CandidateMetaPath:   entry.Path,
		},
	}
}

func remoteCandidate(repo internal.Repo) internal.Candidate {
	return internal.Candidate{
		Meta: map[string]string{
			internal.CandidateMetaSource:   internal.CandidateSourceRemote,
			internal.CandidateMetaName:     repo.Name,
			internal.CandidateMetaFullName: repo.FullName,
			internal.CandidateMetaSSHURL:   repo.SSHURL,
		},
	}
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
			source = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			refresh = &RefreshableMock{RefreshFunc: func(ctx context.Context) error { return nil }}
		)

		sut := newSut(t, []internal.Source{refreshableSource{source: source, refreshable: refresh}}, picker)

		err := sut.Refresh(context.Background())

		require.NoError(t, err)
		assert.Len(t, refresh.RefreshCalls(), 1)
	})
}

func TestQuery(t *testing.T) {
	t.Run("continue loading candidates when one source fails", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		failing := &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return nil, errors.New("source unavailable")
		}}
		working := resolverSource{source: &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: valid})}, nil
		}}, resolveFunc: func(candidate internal.Candidate) (string, error) {
			return candidate.Meta[internal.CandidateMetaPath], nil
		}}

		picker := &PickerMock{PickFunc: func(query string, candidates []internal.Candidate) (int, bool, error) {
			t.Fatalf("picker should not be called")
			return -1, false, nil
		}}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{failing, working}, picker)

		err := sut.Query(context.Background(), "alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("return source error when no candidates available", func(t *testing.T) {
		var (
			expectedErr = errors.New("source unavailable")
			failing     = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, expectedErr
			}}
			empty = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
		)

		sut := newSut(t, []internal.Source{failing, empty}, &PickerMock{})

		err := sut.Query(context.Background(), "alpha", &bytes.Buffer{})

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("remove missing local selection", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			local  = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: "/work/alpha"})}, nil
			}}
			remote = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			resolverCalled bool
		)

		app, err := internal.NewApp([]internal.Source{
			resolverSource{source: local, resolveFunc: func(candidate internal.Candidate) (string, error) {
				resolverCalled = true
				return "", errors.New("repo no longer exists: /work/alpha")
			}},
			remote,
		}, internal.WithPicker(picker))
		require.NoError(t, err)

		err = app.Query(context.Background(), "alpha", &bytes.Buffer{})

		assert.Error(t, err)
		assert.True(t, resolverCalled)
	})

	t.Run("return error when no repos", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			local  = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			remote = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
		)

		sut := newSut(t, []internal.Source{local, remote}, picker)

		err := sut.Query(context.Background(), "", &bytes.Buffer{})

		assert.ErrorIs(t, err, internal.ErrNoReposTracked)
	})

	t.Run("return error when no matches", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			local  = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
			remote = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
		)

		sut := newSut(t, []internal.Source{local, remote}, picker)

		err := sut.Query(context.Background(), "zzz", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("return error when resolve fails", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			local  = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, nil
			}}
			remote = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
			}}
			remoteResolver = resolverSource{source: remote, resolveFunc: func(candidate internal.Candidate) (string, error) {
				if candidate.Meta[internal.CandidateMetaSource] != internal.CandidateSourceRemote {
					return "", internal.ErrUnsupportedCandidate
				}

				return "", errors.New("boom")
			}}
		)

		picker.PickFunc = func(query string, candidates []internal.Candidate) (int, bool, error) {
			return 0, true, nil
		}
		sut := newSut(t, []internal.Source{local, remoteResolver}, picker)

		err := sut.Query(context.Background(), "beta", &bytes.Buffer{})

		assert.Error(t, err)
	})

	t.Run("skip picker when single local match", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		local := resolverSource{source: &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: valid})}, nil
		}}, resolveFunc: func(candidate internal.Candidate) (string, error) {
			return candidate.Meta[internal.CandidateMetaPath], nil
		}}
		remote := &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return nil, internal.ErrNoReposTracked
		}}

		picker := &PickerMock{PickFunc: func(query string, candidates []internal.Candidate) (int, bool, error) {
			t.Fatalf("picker should not be called")
			return -1, false, nil
		}}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{local, remote}, picker)

		err := sut.Query(context.Background(), "alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("print local selection", func(t *testing.T) {
		var (
			root  = t.TempDir()
			valid = filepath.Join(root, "alpha")
		)
		require.NoError(t, os.MkdirAll(filepath.Join(valid, ".git"), 0o755))

		local := resolverSource{source: &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return []internal.Candidate{localCandidate(internal.Entry{Name: "alpha", Path: valid})}, nil
		}}, resolveFunc: func(candidate internal.Candidate) (string, error) {
			return candidate.Meta[internal.CandidateMetaPath], nil
		}}
		remote := &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
			return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
		}}

		picker := &PickerMock{PickFunc: func(query string, candidates []internal.Candidate) (int, bool, error) {
			return 0, true, nil
		}}

		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{local, remote}, picker)

		err := sut.Query(context.Background(), "alpha", out)

		require.NoError(t, err)
		assert.Contains(t, out.String(), valid)
	})

	t.Run("resolve remote selection", func(t *testing.T) {
		var (
			picker = &PickerMock{}
			local  = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			remote = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return []internal.Candidate{remoteCandidate(internal.Repo{Name: "beta", FullName: "acme/beta"})}, nil
			}}
			remoteResolver = resolverSource{source: remote, resolveFunc: func(candidate internal.Candidate) (string, error) {
				if candidate.Meta[internal.CandidateMetaSource] != internal.CandidateSourceRemote {
					return "", internal.ErrUnsupportedCandidate
				}

				return "/work/beta", nil
			}}
			trackable = &TrackableMock{TrackFunc: func(path string) error { return nil }}
		)

		picker.PickFunc = func(query string, candidates []internal.Candidate) (int, bool, error) {
			return 0, true, nil
		}
		out := &bytes.Buffer{}
		sut := newSut(t, []internal.Source{
			local,
			remoteResolver,
			trackableSource{source: &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}, trackable: trackable},
		}, picker)

		err := sut.Query(context.Background(), "beta", out)

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
			source = &SourceMock{LoadFunc: func(_ context.Context, query string) ([]internal.Candidate, error) {
				return nil, internal.ErrNoReposTracked
			}}
			prunable = &PrunableMock{PruneFunc: func() error { return nil }}
		)

		sut := newSut(t, []internal.Source{prunableSource{source: source, prunable: prunable}}, picker)

		err := sut.Prune()

		require.NoError(t, err)
		assert.Len(t, prunable.PruneCalls(), 1)
	})
}
