package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	fly "github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/substring"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/by_signal"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/minipick"
)

const (
	EnvConfigPath = "FLY_CONFIG"
	appName       = "fly"
	fileName      = "config.toml"

	SourceLocal  = "local"
	SourceRemote = "remote"
	SourceZoxide = "zoxide"

	WindowTop    = "top"
	WindowBottom = "bottom"

	MatcherOrderedChars = "orderedchars"
	MatcherSubstring    = "substring"

	SorterSignalMinipick = "signal_minipick"
	SorterMinipick       = "minipick"
)

type Config struct {
	Version int           `toml:"version"`
	Sources SourcesConfig `toml:"sources"`
	Picker  PickerConfig  `toml:"picker"`
}

type SourcesConfig struct {
	Enabled []string `toml:"enabled"`
}

type PickerConfig struct {
	Title          string `toml:"title"`
	PromptMarker   string `toml:"prompt_marker"`
	WindowPosition string `toml:"window_position"`
	Matcher        string `toml:"matcher"`
	Sorter         string `toml:"sorter"`
}

func Default() Config {
	return Config{
		Version: 1,
		Sources: SourcesConfig{Enabled: []string{SourceZoxide, SourceRemote}},
		Picker: PickerConfig{
			PromptMarker:   "> ",
			WindowPosition: WindowTop,
			Matcher:        MatcherOrderedChars,
			Sorter:         SorterSignalMinipick,
		},
	}
}

func Load() (Config, error) {
	path, err := resolvePath()
	if err != nil {
		return Config{}, err
	}

	return loadFromPath(path)
}

func (c Config) PickerOptions(output io.Writer) ([]picker.Option, error) {
	if err := validate(c); err != nil {
		return nil, err
	}

	options := []picker.Option{
		picker.WithOutput(output),
		picker.WithTitle(c.Picker.Title),
		picker.WithPrompt(c.Picker.PromptMarker),
	}

	windowPosition, err := toWindowPosition(c.Picker.WindowPosition)
	if err != nil {
		return nil, err
	}
	options = append(options, picker.WithWindowPosition(windowPosition))

	matcher, err := toMatcher(c.Picker.Matcher)
	if err != nil {
		return nil, err
	}
	options = append(options, picker.WithMatcher(matcher))

	sorter, err := toSorter(c.Picker.Sorter)
	if err != nil {
		return nil, err
	}
	options = append(options, picker.WithSorter(sorter))

	return options, nil
}

func resolvePath() (string, error) {
	override := strings.TrimSpace(os.Getenv(EnvConfigPath))
	if override != "" {
		return override, nil
	}

	configDir, err := fly.ConfigDir(appName)
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, fileName), nil
}

func loadFromPath(path string) (Config, error) {
	config := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}

		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	metadata, err := toml.Decode(string(data), &config)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		unknown := make([]string, len(undecoded))
		for i, key := range undecoded {
			unknown[i] = key.String()
		}

		return Config{}, fmt.Errorf("unknown config keys: %s", strings.Join(unknown, ", "))
	}

	if err := validate(config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func validate(config Config) error {
	if config.Version != 1 {
		return fmt.Errorf("unsupported config version %d", config.Version)
	}

	if len(config.Sources.Enabled) == 0 {
		return fmt.Errorf("sources.enabled must contain at least one source")
	}

	seen := map[string]struct{}{}
	for _, source := range config.Sources.Enabled {
		normalized := strings.ToLower(strings.TrimSpace(source))
		if normalized == "" {
			return fmt.Errorf("sources.enabled contains an empty source")
		}

		switch normalized {
		case SourceLocal, SourceRemote, SourceZoxide:
		default:
			return fmt.Errorf("unsupported source %q", source)
		}

		if _, exists := seen[normalized]; exists {
			return fmt.Errorf("duplicate source %q in sources.enabled", normalized)
		}
		seen[normalized] = struct{}{}
	}

	if _, err := toWindowPosition(config.Picker.WindowPosition); err != nil {
		return err
	}
	if _, err := toMatcher(config.Picker.Matcher); err != nil {
		return err
	}
	if _, err := toSorter(config.Picker.Sorter); err != nil {
		return err
	}

	return nil
}

func toWindowPosition(value string) (picker.WindowPosition, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case WindowTop:
		return picker.WindowTop, nil
	case WindowBottom:
		return picker.WindowBottom, nil
	default:
		return picker.WindowTop, fmt.Errorf("unsupported picker.window_position %q", value)
	}
}

func toMatcher(value string) (matchers.Matcher, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case MatcherOrderedChars:
		return orderedchars.New(), nil
	case MatcherSubstring:
		return substring.New(), nil
	default:
		return nil, fmt.Errorf("unsupported picker.matcher %q", value)
	}
}

func toSorter(value string) (sorters.Sorter, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case SorterSignalMinipick:
		return by_signal.New(fly.CandidateSignalZoxideScore, minipick.New()), nil
	case SorterMinipick:
		return minipick.New(), nil
	default:
		return nil, fmt.Errorf("unsupported picker.sorter %q", value)
	}
}
