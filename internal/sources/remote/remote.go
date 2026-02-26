package remote

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Source struct {
	store         *RemoteStore
	fetcher       Fetcher
	refreshLaunch internal.Refresher
	cloner        internal.Cloner
	now           func() time.Time
}

type Fetcher interface {
	FetchAll(context.Context) ([]internal.Repo, error)
}

func New(fetcher Fetcher, refreshLaunch internal.Refresher, cloner internal.Cloner) (*Source, error) {
	store, err := DefaultRemoteStore()
	if err != nil {
		return nil, err
	}
	if fetcher == nil {
		return nil, fmt.Errorf("remote fetcher required")
	}
	if refreshLaunch == nil {
		return nil, fmt.Errorf("refresh launcher required")
	}
	if cloner == nil {
		return nil, fmt.Errorf("cloner required")
	}

	return &Source{
		store:         store,
		fetcher:       fetcher,
		refreshLaunch: refreshLaunch,
		cloner:        cloner,
		now:           time.Now,
	}, nil
}

func (s *Source) Load(query string) ([]internal.Candidate, error) {
	cache, exists, err := s.store.Load()
	if err != nil {
		cache = Cache{}
		exists = false
	}

	if ShouldRefresh(cache, exists) {
		s.refreshLaunch.Launch()
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
