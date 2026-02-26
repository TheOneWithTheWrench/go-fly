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
	fetcher       internal.RemoteFetcher
	refreshLaunch internal.Refresher
	now           func() time.Time
}

func New(fetcher internal.RemoteFetcher, refreshLaunch internal.Refresher) (*Source, error) {
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

	return &Source{
		store:         store,
		fetcher:       fetcher,
		refreshLaunch: refreshLaunch,
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
		candidates = append(candidates, internal.Candidate{Kind: internal.KindRemote, Remote: repo})
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
