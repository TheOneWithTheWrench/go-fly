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
	FetchAll(context.Context, internal.RefreshOutput) ([]internal.Repo, error)
}

type Option func(*options) error

type options struct {
	fetcher       Fetcher
	cloner        internal.Cloner
	cloneBaseDir  string
	groupByOwner  bool
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

func WithCloneBaseDir(baseDir string) Option {
	return func(opts *options) error {
		opts.cloneBaseDir = strings.TrimSpace(baseDir)
		return nil
	}
}

func WithCloneGroupByOwner() Option {
	return func(opts *options) error {
		opts.groupByOwner = true
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
		runner:        defaultRunner(),
		refreshLaunch: newDetachedRefresher(),
	}

	for _, option := range sourceOptions {
		if err := option(&defaultOpts); err != nil {
			return nil, err
		}
	}

	if defaultOpts.cloner != nil && (defaultOpts.cloneBaseDir != "" || defaultOpts.groupByOwner) {
		return nil, fmt.Errorf("with cloner and clone config options are mutually exclusive")
	}

	if defaultOpts.cloner == nil {
		clonerOptionFuncs := make([]GitHubClonerOption, 0, 2)
		if defaultOpts.cloneBaseDir != "" {
			clonerOptionFuncs = append(clonerOptionFuncs, withCloneBaseDir(defaultOpts.cloneBaseDir))
		}
		if defaultOpts.groupByOwner {
			clonerOptionFuncs = append(clonerOptionFuncs, withGroupByOwner(true))
		}

		defaultCloner, err := NewGitHubCloner(os.Stdin, os.Stderr, clonerOptionFuncs...)
		if err != nil {
			return nil, err
		}
		defaultOpts.cloner = defaultCloner
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
		return nil, fmt.Errorf("load cache: %w", err)
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
	if err != nil {
		return
	}
	if !ShouldLaunchRefresh(state, exists, s.now()) {
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

func (s *Source) Refresh(ctx context.Context, output internal.RefreshOutput) error {
	if err := output.SetStatus("Refreshing remote repositories..."); err != nil {
		return err
	}

	repos, err := s.fetcher.FetchAll(ctx, output)
	if err != nil {
		return fmt.Errorf("fetch repos: %w", err)
	}

	if err := output.SetStatus("Saving remote repository cache..."); err != nil {
		return err
	}

	cache := Cache{
		FetchedAt: s.now().UTC(),
		Repos:     repos,
	}

	if err := s.store.Save(cache); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(output, "Refreshed %d remote repositories\n", len(repos)); err != nil {
		return err
	}

	return nil
}
