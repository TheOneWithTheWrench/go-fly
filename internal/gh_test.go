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
		)

		responses := map[string][]byte{
			"api user/orgs --paginate": []byte(`[{"login":"acme"}]`),
			"api user/repos --paginate": []byte(`[
  {"name":"repo","full_name":"user/repo","ssh_url":"git@github.com:user/repo.git"},
  {"name":"shared","full_name":"acme/shared","ssh_url":"git@github.com:acme/shared.git"}
]`),
			"api orgs/acme/repos --paginate": []byte(`[
  {"name":"shared","full_name":"acme/shared","ssh_url":"git@github.com:acme/shared.git"},
  {"name":"tools","full_name":"acme/tools","ssh_url":"git@github.com:acme/tools.git"}
]`),
		}

		runner.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, errors.New("unexpected command")
			}

			key := strings.Join(args, " ")
			resp, ok := responses[key]
			if !ok {
				return nil, errors.New("unexpected args: " + key)
			}
			return resp, nil
		}

		got, err := sut.FetchAll(context.Background())

		require.NoError(t, err)
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
			responses = map[string]struct {
				output []byte
				err    error
			}{
				"api user/orgs --paginate": {
					output: []byte(`[{"login":"acme"}]`),
				},
				"api user/repos --paginate": {
					output: []byte(`[{"name":"repo","full_name":"user/repo","ssh_url":"git@github.com:user/repo.git"}]`),
				},
				"api orgs/acme/repos --paginate": {
					output: []byte("HTTP 403"),
					err:    errors.New("exit status 1"),
				},
			}
		)

		runner.RunFunc = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			if name != "gh" {
				return nil, errors.New("unexpected command")
			}

			key := strings.Join(args, " ")
			resp, ok := responses[key]
			if !ok {
				return nil, errors.New("unexpected args: " + key)
			}
			return resp.output, resp.err
		}

		got, err := sut.FetchAll(context.Background())

		require.NoError(t, err)
		assert.Equal(t, []internal.Repo{
			{Name: "repo", FullName: "user/repo", SSHURL: "git@github.com:user/repo.git"},
		}, got)
	})
}
