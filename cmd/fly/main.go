package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	fly "github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/remote"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/zoxide"
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
			return appInstance.Query(context.Background(), query, out)
		},
	})
}

func newApp() (*fly.App, error) {
	remoteSource, err := remote.New()
	if err != nil {
		return nil, err
	}

	zoxideSource, err := zoxide.New()
	if err != nil {
		return nil, err
	}

	appInstance, err := fly.NewApp(
		[]fly.Source{zoxideSource, remoteSource},
	)
	if err != nil {
		return nil, err
	}

	return appInstance, nil
}
