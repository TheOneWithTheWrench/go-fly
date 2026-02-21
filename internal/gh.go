package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Client struct {
	runner Runner
}

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

var (
	ErrForbidden = errors.New("forbidden")
)

func NewClient(runner Runner) *Client {
	return &Client{runner: runner}
}

func (c *Client) FetchAll(ctx context.Context) ([]Repo, error) {
	if c.runner == nil {
		return nil, fmt.Errorf("runner required")
	}

	orgs, err := c.listOrgs(ctx)
	if err != nil {
		return nil, err
	}

	userRepos, err := c.listRepos(ctx, "user/repos")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(userRepos))
	repos := make([]Repo, 0, len(userRepos))

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

	for _, org := range orgs {
		orgRepos, err := c.listRepos(ctx, fmt.Sprintf("orgs/%s/repos", org))
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

func (c *Client) listOrgs(ctx context.Context) ([]string, error) {
	data, err := c.run(ctx, "api", "user/orgs", "--paginate")
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

func (c *Client) listRepos(ctx context.Context, endpoint string) ([]Repo, error) {
	data, err := c.run(ctx, "api", endpoint, "--paginate")
	if err != nil {
		if isOrgEndpoint(endpoint) {
			if errors.Is(err, ErrForbidden) {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("list repos: %w", err)
	}

	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("parse repos: %w", err)
	}

	return repos, nil
}

func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	data, err := c.runner.Run(ctx, "gh", args...)
	if err != nil {
		if isForbidden(data, err) {
			return nil, fmt.Errorf("%w: %s", ErrForbidden, err)
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
