package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	fly "github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
)

func main() {
	if err := run(os.Args, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	appInstance, err := newApp()
	if err != nil {
		return err
	}

	return fly.Run(args, stdout, stderr, fly.CliDependencies{
		Init: func(out io.Writer) error {
			_, err := fmt.Fprint(stdout, fly.ZshInitSnippet())
			return err
		},
		Refresh: func() error {
			return appInstance.Refresh(context.Background())
		},
		Prune: func() error {
			return appInstance.Prune()
		},
		Track: func() error {
			currentDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working dir: %w", err)
			}

			root, ok, err := fly.Find(currentDir)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			absRoot, err := filepath.Abs(root)
			if err != nil {
				return fmt.Errorf("resolve repo path: %w", err)
			}

			return appInstance.Track(absRoot)
		},
		Query: func(query string, out io.Writer) error {
			return appInstance.Query(query, out)
		},
	})
}

func newApp() (*fly.App, error) {
	client := fly.NewClient(newRunnerFunc())
	localSource, err := local.New(newPrunerFunc())
	if err != nil {
		return nil, err
	}
	remoteSource, err := remote.New(client, newRefresherFunc())
	if err != nil {
		return nil, err
	}

	appInstance, err := fly.NewApp(
		[]fly.Source{localSource, remoteSource},
		newPickerFunc(),
		newClonerFunc(),
	)
	if err != nil {
		return nil, err
	}

	return appInstance, nil
}

func newPickerFunc() fly.Picker {
	return fly.PickerFunc(fly.Pick)
}

func newRefresherFunc() fly.RefresherFunc {
	return fly.RefresherFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		cmd := exec.Command(exe, "refresh")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			return
		}

		_ = cmd.Process.Release()
	})
}

func newClonerFunc() fly.ClonerFunc {
	return fly.ClonerFunc(func(repo fly.Repo) (string, error) {
		if repo.FullName == "" {
			return "", fmt.Errorf("repo full name required")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working dir: %w", err)
		}

		dest, err := fly.Destination(repo, cwd)
		if err != nil {
			return "", err
		}

		exists, err := fly.CheckDestination(dest)
		if err != nil {
			return "", err
		}
		if exists {
			return dest, nil
		}

		cmd := exec.Command("gh", "repo", "clone", repo.FullName)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("clone %s: %w", repo.FullName, err)
		}

		return dest, nil
	})
}

func newPrunerFunc() fly.PrunerFunc {
	return fly.PrunerFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		cmd := exec.Command(exe, "_prune")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			return
		}

		_ = cmd.Process.Release()
	})
}

func newRunnerFunc() fly.RunnerFunc {
	return fly.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return output, fmt.Errorf("run %s %v: %w: %s", name, args, err, output)
		}
		return output, nil
	})
}
