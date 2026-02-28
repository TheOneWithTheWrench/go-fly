package remote

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Source struct {
	store         *RemoteStore
	refreshState  *RefreshStateStore
	fetcher       Fetcher
	refreshLaunch internal.Refresher
	cloner        internal.Cloner
	now           func() time.Time
	refreshMu     sync.Mutex
}

type Fetcher interface {
	FetchAll(context.Context) ([]internal.Repo, error)
}

type Option func(*options) error

type options struct {
	fetcher       Fetcher
	cloner        internal.Cloner
	runner        internal.Runner
	refreshLaunch internal.Refresher
}

func WithFetcher(fetcher Fetcher) Option {
	return func(opts *options) error {
		if fetcher == nil {
			return fmt.Errorf("remote fetcher required")
		}

		opts.fetcher = fetcher
		return nil
	}
}

func WithCloner(cloner internal.Cloner) Option {
	return func(opts *options) error {
		if cloner == nil {
			return fmt.Errorf("cloner required")
		}

		opts.cloner = cloner
		return nil
	}
}

func WithRunner(runner internal.Runner) Option {
	return func(opts *options) error {
		if runner == nil {
			return fmt.Errorf("runner required")
		}

		opts.runner = runner
		return nil
	}
}

func WithRefreshLauncher(refreshLaunch internal.Refresher) Option {
	return func(opts *options) error {
		if refreshLaunch == nil {
			return fmt.Errorf("refresh launcher required")
		}

		opts.refreshLaunch = refreshLaunch
		return nil
	}
}

func New(sourceOptions ...Option) (*Source, error) {
	defaultOpts := options{
		cloner:        NewGitHubCloner(os.Stdin, os.Stderr),
		runner:        defaultRunner(),
		refreshLaunch: newDetachedRefresher(),
	}

	for _, option := range sourceOptions {
		if err := option(&defaultOpts); err != nil {
			return nil, err
		}
	}

	if defaultOpts.fetcher == nil {
		defaultOpts.fetcher = NewGitHubFetcher(defaultOpts.runner)
	}

	store, err := NewRemoteStore()
	if err != nil {
		return nil, err
	}
	refreshState, err := NewRefreshStateStore()
	if err != nil {
		return nil, err
	}

	return &Source{
		store:         store,
		refreshState:  refreshState,
		fetcher:       defaultOpts.fetcher,
		refreshLaunch: defaultOpts.refreshLaunch,
		cloner:        defaultOpts.cloner,
		now:           time.Now,
	}, nil
}

func defaultRunner() internal.Runner {
	return internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return output, fmt.Errorf("run %s %v: %w: %s", name, args, err, output)
		}

		return output, nil
	})
}

func (s *Source) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	_ = ctx

	cache, exists, err := s.store.Load()
	if err != nil {
		cache = Cache{}
		exists = false
	}

	if ShouldRefresh(cache, exists) {
		s.launchRefresh()
	}

	total := len(cache.Repos)
	if total == 0 {
		return nil, internal.ErrNoReposTracked
	}
	filtered := FilterRepos(query, cache.Repos)

	candidates := make([]internal.Candidate, 0, len(filtered))
	for _, repo := range filtered {
		candidates = append(candidates, internal.Candidate{
			Meta: map[string]string{
				internal.CandidateMetaSource:   internal.CandidateSourceRemote,
				internal.CandidateMetaName:     repo.Name,
				internal.CandidateMetaFullName: repo.FullName,
				internal.CandidateMetaSSHURL:   repo.SSHURL,
			},
		})
	}

	return candidates, nil
}

func (s *Source) launchRefresh() {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()

	state, exists, err := s.refreshState.Load()
	if err == nil && !ShouldLaunchRefresh(state, exists, s.now()) {
		return
	}

	now := s.now().UTC()
	if err := s.refreshState.Save(RefreshState{StartedAt: now}); err != nil {
		return
	}

	s.refreshLaunch.Launch()
}

func FilterRepos(query string, repos []internal.Repo) []internal.Repo {
	if strings.TrimSpace(query) == "" {
		return repos
	}

	query = strings.ToLower(query)
	filtered := make([]internal.Repo, 0, len(repos))

	for _, repo := range repos {
		name := strings.ToLower(repo.Name)
		fullName := strings.ToLower(repo.FullName)
		if strings.Contains(name, query) || strings.Contains(fullName, query) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}

func (s *Source) Resolve(candidate internal.Candidate) (string, error) {
	source := candidate.Meta[internal.CandidateMetaSource]
	if source != internal.CandidateSourceRemote {
		return "", internal.ErrUnsupportedCandidate
	}

	repo := internal.Repo{
		Name:     candidate.Meta[internal.CandidateMetaName],
		FullName: candidate.Meta[internal.CandidateMetaFullName],
		SSHURL:   candidate.Meta[internal.CandidateMetaSSHURL],
	}
	if repo.FullName == "" {
		return "", internal.ErrUnsupportedCandidate
	}

	path, err := s.cloner.Clone(repo)
	if err != nil {
		return "", err
	}

	return path, nil
}

func (s *Source) Refresh(ctx context.Context) error {
	repos, err := s.fetcher.FetchAll(ctx)
	if err != nil {
		return fmt.Errorf("fetch repos: %w", err)
	}

	cache := Cache{
		FetchedAt: s.now().UTC(),
		Repos:     repos,
	}

	return s.store.Save(cache)
}
