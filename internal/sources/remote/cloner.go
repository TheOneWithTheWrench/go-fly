package remote

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	runner  commandRunner
	getwd   func() (string, error)
	stdin   io.Reader
	stderr  io.Writer
	baseDir string
}

type GitHubClonerOption func(*gitHubClonerOptions) error

type gitHubClonerOptions struct {
	baseDir string
	runner  commandRunner
	getwd   func() (string, error)
}

func withCloneBaseDir(baseDir string) GitHubClonerOption {
	return func(opts *gitHubClonerOptions) error {
		opts.baseDir = strings.TrimSpace(baseDir)
		return nil
	}
}

func NewGitHubCloner(stdin io.Reader, stderr io.Writer, optionFuncs ...GitHubClonerOption) (*GitHubCloner, error) {
	defaultOpts := gitHubClonerOptions{
		runner: execCommandRunner{},
		getwd:  os.Getwd,
	}
	for _, optionFunc := range optionFuncs {
		if err := optionFunc(&defaultOpts); err != nil {
			return nil, err
		}
	}

	return &GitHubCloner{
		runner:  defaultOpts.runner,
		getwd:   defaultOpts.getwd,
		stdin:   stdin,
		stderr:  stderr,
		baseDir: defaultOpts.baseDir,
	}, nil
}

func withCommandRunner(runner commandRunner) GitHubClonerOption {
	return func(opts *gitHubClonerOptions) error {
		if runner == nil {
			return fmt.Errorf("command runner required")
		}

		opts.runner = runner
		return nil
	}
}

func withGetwd(getwd func() (string, error)) GitHubClonerOption {
	return func(opts *gitHubClonerOptions) error {
		if getwd == nil {
			return fmt.Errorf("getwd required")
		}

		opts.getwd = getwd
		return nil
	}
}

func (c *GitHubCloner) Clone(repo internal.Repo) (string, error) {
	if repo.FullName == "" {
		return "", fmt.Errorf("repo full name required")
	}

	baseDir, err := c.resolveBaseDir()
	if err != nil {
		return "", err
	}

	dest, err := Destination(repo, baseDir)
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

	err = c.runner.Run("gh", []string{"repo", "clone", repo.FullName, dest}, c.stdin, c.stderr, c.stderr)
	if err != nil {
		return "", fmt.Errorf("clone %s: %w", repo.FullName, err)
	}

	return dest, nil
}

func (c *GitHubCloner) resolveBaseDir() (string, error) {
	baseDir := strings.TrimSpace(c.baseDir)
	if baseDir == "" {
		cwd, err := c.getwd()
		if err != nil {
			return "", fmt.Errorf("get working dir: %w", err)
		}

		return cwd, nil
	}

	if baseDir == "~" || strings.HasPrefix(baseDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}

		if baseDir == "~" {
			baseDir = homeDir
		} else {
			baseDir = filepath.Join(homeDir, strings.TrimPrefix(baseDir, "~/"))
		}
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("resolve clone base dir: %w", err)
	}

	return absBaseDir, nil
}
