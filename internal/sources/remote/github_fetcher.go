package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

var errForbidden = errors.New("forbidden")

type GitHubFetcher struct {
	runner internal.Runner
}

func NewGitHubFetcher(runner internal.Runner) *GitHubFetcher {
	return &GitHubFetcher{runner: runner}
}

func (f *GitHubFetcher) FetchAll(ctx context.Context, output internal.RefreshOutput) ([]internal.Repo, error) {
	if f.runner == nil {
		return nil, fmt.Errorf("runner required")
	}

	if err := output.SetStatus("Fetching GitHub organizations..."); err != nil {
		return nil, err
	}
	orgs, err := f.listOrgs(ctx)
	if err != nil {
		return nil, err
	}

	if err := output.SetStatus("Fetching personal repositories..."); err != nil {
		return nil, err
	}
	userRepos, err := f.listRepos(ctx, "user/repos")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(userRepos))
	repos := make([]internal.Repo, 0, len(userRepos))

	for _, repo := range userRepos {
		if repo.FullName == "" {
			continue
		}
		if _, ok := seen[repo.FullName]; ok {
			continue
		}
		seen[repo.FullName] = struct{}{}
		repos = append(repos, repo)
	}

	for i, org := range orgs {
		if err := output.SetStatus(fmt.Sprintf("Fetching organization repositories %d/%d: %s...", i+1, len(orgs), org)); err != nil {
			return nil, err
		}
		orgRepos, err := f.listRepos(ctx, fmt.Sprintf("orgs/%s/repos", org))
		if err != nil {
			return nil, err
		}
		for _, repo := range orgRepos {
			if repo.FullName == "" {
				continue
			}
			if _, ok := seen[repo.FullName]; ok {
				continue
			}
			seen[repo.FullName] = struct{}{}
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

func (f *GitHubFetcher) listOrgs(ctx context.Context) ([]string, error) {
	data, err := f.run(ctx, "api", "user/orgs", "--method", "GET", "--paginate", "-F", "per_page=100")
	if err != nil {
		return nil, fmt.Errorf("list orgs: %w", err)
	}

	var orgs []struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(data, &orgs); err != nil {
		return nil, fmt.Errorf("parse orgs: %w", err)
	}

	result := make([]string, 0, len(orgs))
	for _, org := range orgs {
		if org.Login != "" {
			result = append(result, org.Login)
		}
	}

	return result, nil
}

func (f *GitHubFetcher) listRepos(ctx context.Context, endpoint string) ([]internal.Repo, error) {
	data, err := f.run(ctx, "api", endpoint, "--method", "GET", "--paginate", "-F", "per_page=100")
	if err != nil {
		if isOrgEndpoint(endpoint) && errors.Is(err, errForbidden) {
			return nil, nil
		}
		return nil, fmt.Errorf("list repos: %w", err)
	}

	var repos []internal.Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("parse repos: %w", err)
	}

	return repos, nil
}

func (f *GitHubFetcher) run(ctx context.Context, args ...string) ([]byte, error) {
	data, err := f.runner.Run(ctx, "gh", args...)
	if err != nil {
		if isForbidden(data, err) {
			return nil, fmt.Errorf("%w: %s", errForbidden, err)
		}
		return nil, err
	}

	return data, nil
}

func isOrgEndpoint(endpoint string) bool {
	return strings.HasPrefix(endpoint, "orgs/")
}

func isForbidden(output []byte, err error) bool {
	if err == nil {
		return false
	}

	text := string(output) + " " + err.Error()
	return strings.Contains(text, "HTTP 403") ||
		strings.Contains(text, "\"status\":\"403\"") ||
		strings.Contains(text, "\"status\":403") ||
		strings.Contains(text, "status: 403")
}
