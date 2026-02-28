package remote

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubCloner(t *testing.T) {
	t.Run("return error when repo full name is missing", func(t *testing.T) {
		var (
			runner = &commandRunnerMock{RunFunc: func(string, []string, io.Reader, io.Writer, io.Writer) error {
				return nil
			}}
			sut = newGitHubCloner(runner, func() (string, error) { return t.TempDir(), nil }, bytes.NewBuffer(nil), bytes.NewBuffer(nil))
		)

		_, err := sut.Clone(internal.Repo{Name: "repo"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "repo full name required")
		assert.Len(t, runner.RunCalls(), 0)
	})

	t.Run("return destination when repo already exists", func(t *testing.T) {
		var (
			cwd    = t.TempDir()
			repo   = internal.Repo{Name: "go-fly", FullName: "acme/go-fly"}
			dest   = filepath.Join(cwd, "go-fly")
			runner = &commandRunnerMock{RunFunc: func(string, []string, io.Reader, io.Writer, io.Writer) error {
				return nil
			}}
			stderrBuf = bytes.NewBuffer(nil)
			sut       = newGitHubCloner(runner, func() (string, error) { return cwd, nil }, bytes.NewBuffer(nil), stderrBuf)
		)
		require.NoError(t, os.MkdirAll(filepath.Join(dest, ".git"), 0o755))

		got, err := sut.Clone(repo)

		require.NoError(t, err)
		assert.Equal(t, dest, got)
		assert.Len(t, runner.RunCalls(), 0)
	})

	t.Run("clone with gh when repo does not exist", func(t *testing.T) {
		var (
			cwd       = t.TempDir()
			repo      = internal.Repo{Name: "go-fly", FullName: "acme/go-fly"}
			stdinBuf  = bytes.NewBufferString("input")
			stderrBuf = bytes.NewBuffer(nil)
			runner    = &commandRunnerMock{RunFunc: func(string, []string, io.Reader, io.Writer, io.Writer) error {
				return nil
			}}
			sut = newGitHubCloner(runner, func() (string, error) { return cwd, nil }, stdinBuf, stderrBuf)
		)

		got, err := sut.Clone(repo)

		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, "go-fly"), got)
		require.Len(t, runner.RunCalls(), 1)
		assert.Equal(t, "gh", runner.RunCalls()[0].Name)
		assert.Equal(t, []string{"repo", "clone", "acme/go-fly"}, runner.RunCalls()[0].Args)
		assert.Same(t, stdinBuf, runner.RunCalls()[0].Stdin)
		assert.Same(t, stderrBuf, runner.RunCalls()[0].Stdout)
		assert.Same(t, stderrBuf, runner.RunCalls()[0].Stderr)
	})

	t.Run("return clone error", func(t *testing.T) {
		var (
			cwd         = t.TempDir()
			expectedErr = errors.New("clone failed")
			runner      = &commandRunnerMock{RunFunc: func(string, []string, io.Reader, io.Writer, io.Writer) error {
				return expectedErr
			}}
			sut = newGitHubCloner(runner, func() (string, error) { return cwd, nil }, bytes.NewBuffer(nil), bytes.NewBuffer(nil))
		)

		_, err := sut.Clone(internal.Repo{Name: "go-fly", FullName: "acme/go-fly"})

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Contains(t, err.Error(), "clone acme/go-fly")
		require.Len(t, runner.RunCalls(), 1)
	})
}
