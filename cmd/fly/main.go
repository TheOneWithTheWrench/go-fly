package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	fly "github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/config"
	"github.com/TheOneWithTheWrench/go-fly/internal/sources/local"
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
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	appInstance, err := newApp(cfg, isCloneToCWDInvocation(args))
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

func newApp(cfg config.Config, forceCloneToCWD bool) (*fly.App, error) {
	sources, err := newSourcesFromCfg(cfg, forceCloneToCWD)
	if err != nil {
		return nil, err
	}

	picker, err := newPickerFromCfg(cfg)
	if err != nil {
		return nil, err
	}

	appInstance, err := fly.NewApp(
		sources,
		fly.WithPicker(picker),
	)
	if err != nil {
		return nil, err
	}

	return appInstance, nil
}

func newSourcesFromCfg(cfg config.Config, forceCloneToCWD bool) ([]fly.Source, error) {
	cloneBaseDir := strings.TrimSpace(cfg.Clone.DefaultDirectory)
	if forceCloneToCWD {
		cloneBaseDir = ""
	}

	sources := make([]fly.Source, 0, len(cfg.Sources.Enabled))
	for _, sourceName := range cfg.Sources.Enabled {
		source, err := newSource(sourceName, cloneBaseDir)
		if err != nil {
			return nil, err
		}

		sources = append(sources, source)
	}

	return sources, nil
}

func newPickerFromCfg(cfg config.Config) (fly.Picker, error) {
	pickerOptions, err := cfg.PickerOptions(os.Stderr)
	if err != nil {
		return nil, err
	}

	return fly.PickerFunc(func(query string, candidates []fly.Candidate) (int, bool, error) {
		return fly.Pick(query, candidates, pickerOptions...)
	}), nil
}

func newSource(sourceName string, cloneBaseDir string) (fly.Source, error) {
	switch strings.ToLower(strings.TrimSpace(sourceName)) {
	case config.SourceLocal:
		return local.New()
	case config.SourceRemote:
		return remote.New(
			remote.WithCloneBaseDir(cloneBaseDir),
		)
	case config.SourceZoxide:
		return zoxide.New()
	default:
		return nil, fmt.Errorf("unsupported source %q", sourceName)
	}
}

func isCloneToCWDInvocation(args []string) bool {
	if len(args) < 2 {
		return false
	}

	return args[1] == fly.CloneToCWDFlag || args[1] == fly.CloneToCWDLongFlag
}
