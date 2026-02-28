package remote

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type commandRunner interface {
	Run(name string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error
}

type execCommandRunner struct{}

func (r execCommandRunner) Run(name string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

type GitHubCloner struct {
	runner commandRunner
	getwd  func() (string, error)
	stdin  io.Reader
	stderr io.Writer
}

func NewGitHubCloner(stdin io.Reader, stderr io.Writer) *GitHubCloner {
	return &GitHubCloner{
		runner: execCommandRunner{},
		getwd:  os.Getwd,
		stdin:  stdin,
		stderr: stderr,
	}
}

func (c *GitHubCloner) Clone(repo internal.Repo) (string, error) {
	if repo.FullName == "" {
		return "", fmt.Errorf("repo full name required")
	}

	cwd, err := c.getwd()
	if err != nil {
		return "", fmt.Errorf("get working dir: %w", err)
	}

	dest, err := Destination(repo, cwd)
	if err != nil {
		return "", err
	}

	exists, err := internal.CheckDestination(dest)
	if err != nil {
		return "", err
	}
	if exists {
		return dest, nil
	}

	err = c.runner.Run("gh", []string{"repo", "clone", repo.FullName}, c.stdin, c.stderr, c.stderr)
	if err != nil {
		return "", fmt.Errorf("clone %s: %w", repo.FullName, err)
	}

	return dest, nil
}

func newGitHubCloner(runner commandRunner, getwd func() (string, error), stdin io.Reader, stderr io.Writer) *GitHubCloner {
	return &GitHubCloner{
		runner: runner,
		getwd:  getwd,
		stdin:  stdin,
		stderr: stderr,
	}
}
