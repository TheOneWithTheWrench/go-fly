package internal_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchAll(t *testing.T) {
	newSut := func(runner internal.Runner) *internal.Client { return internal.NewClient(runner) }
	countMatches := func(calls []string, needle string) int {
		count := 0
		for _, call := range calls {
			if strings.Contains(call, needle) {
				count++
			}
		}
		return count
	}
	assertUsesGet := func(t *testing.T, calls []string) {
		t.Helper()
		for _, call := range calls {
			assert.Contains(t, call, "--method GET")
		}
	}

	t.Run("return error when gh fails", func(t *testing.T) {
		var (
			expected = errors.New("boom")
			runner   = &RunnerMock{
				RunFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
					return nil, expected
				},
			}
			sut = newSut(runner)
		)

		got, err := sut.FetchAll(context.Background())

		assert.ErrorIs(t, err, expected)
		assert.Len(t, got, 0)
	})

	t.Run("fetch all user and org repos", func(t *testing.T) {
		var (
			runner = &RunnerMock{}
			sut    = newSut(runner)
			calls  []string
		)

		responses := map[string][]byte{
			"user/orgs": []byte(`[{"login":"acme"}]`),
			"user/repos": []byte(`[
  {"name":"repo","full_name":"user/repo","ssh_url":"git@github.com:user/repo.git"},
  {"name":"shared","full_name":"acme/shared","ssh_url":"git@github.com:acme/shared.git"}
]`),
			"orgs/acme/repos": []byte(`[
  {"name":"shared","full_name":"acme/shared","ssh_url":"git@github.com:acme/shared.git"},
  {"name":"tools","full_name":"acme/tools","ssh_url":"git@github.com:acme/tools.git"}
]`),
		}

		runner.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, errors.New("unexpected command")
			}

			joined := strings.Join(args, " ")
			calls = append(calls, joined)
			endpoint := ""
			switch {
			case strings.Contains(joined, "user/orgs"):
				endpoint = "user/orgs"
			case strings.Contains(joined, "user/repos"):
				endpoint = "user/repos"
			case strings.Contains(joined, "orgs/acme/repos"):
				endpoint = "orgs/acme/repos"
			}
			resp, ok := responses[endpoint]
			if !ok {
				return nil, errors.New("unexpected args: " + joined)
			}
			return resp, nil
		}

		got, err := sut.FetchAll(context.Background())

		require.NoError(t, err)
		assertUsesGet(t, calls)
		assert.Equal(t, 1, countMatches(calls, "user/orgs"))
		assert.Equal(t, 1, countMatches(calls, "user/repos"))
		assert.Equal(t, 1, countMatches(calls, "orgs/acme/repos"))
		assert.Equal(t, []internal.Repo{
			{Name: "repo", FullName: "user/repo", SSHURL: "git@github.com:user/repo.git"},
			{Name: "shared", FullName: "acme/shared", SSHURL: "git@github.com:acme/shared.git"},
			{Name: "tools", FullName: "acme/tools", SSHURL: "git@github.com:acme/tools.git"},
		}, got)
	})

	t.Run("skip org repos when forbidden", func(t *testing.T) {
		var (
			runner    = &RunnerMock{}
			sut       = newSut(runner)
			calls     []string
			responses = map[string]struct {
				output []byte
				err    error
			}{
				"user/orgs": {
					output: []byte(`[{"login":"acme"}]`),
				},
				"user/repos": {
					output: []byte(`[{"name":"repo","full_name":"user/repo","ssh_url":"git@github.com:user/repo.git"}]`),
				},
				"orgs/acme/repos": {
					output: []byte("HTTP 403"),
					err:    errors.New("exit status 1"),
				},
			}
		)

		runner.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, errors.New("unexpected command")
			}

			joined := strings.Join(args, " ")
			calls = append(calls, joined)
			endpoint := ""
			switch {
			case strings.Contains(joined, "user/orgs"):
				endpoint = "user/orgs"
			case strings.Contains(joined, "user/repos"):
				endpoint = "user/repos"
			case strings.Contains(joined, "orgs/acme/repos"):
				endpoint = "orgs/acme/repos"
			}
			resp, ok := responses[endpoint]
			if !ok {
				return nil, errors.New("unexpected args: " + joined)
			}
			return resp.output, resp.err
		}

		got, err := sut.FetchAll(context.Background())

		require.NoError(t, err)
		assertUsesGet(t, calls)
		assert.Equal(t, 1, countMatches(calls, "user/orgs"))
		assert.Equal(t, 1, countMatches(calls, "user/repos"))
		assert.Equal(t, 1, countMatches(calls, "orgs/acme/repos"))
		assert.Equal(t, []internal.Repo{
			{Name: "repo", FullName: "user/repo", SSHURL: "git@github.com:user/repo.git"},
		}, got)
	})
}
