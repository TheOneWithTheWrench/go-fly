package picker

import (
	"io"
	"os"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
)

// Config holds the picker configuration.
type Config struct {
	Title          string
	Prompt         string
	Matcher        matchers.Matcher
	Sorter         sorters.Sorter
	WindowPosition WindowPosition
	Output         io.Writer
}

// Option configures the picker behavior.
type Option func(*Config)

var (
	defaultOutput io.Writer = os.Stdout
)

// WithTitle sets the header line shown above the list.
func WithTitle(title string) Option {
	return func(config *Config) {
		config.Title = title
	}
}

// WithPrompt sets the input prompt shown before the query.
func WithPrompt(prompt string) Option {
	return func(config *Config) {
		config.Prompt = prompt
	}
}

func WithMatcher(matcher matchers.Matcher) Option {
	return func(config *Config) {
		config.Matcher = matcher
	}
}

func WithSorter(sorter sorters.Sorter) Option {
	return func(config *Config) {
		config.Sorter = sorter
	}
}

// WithWindowPosition sets whether the input is above or below the list.
func WithWindowPosition(position WindowPosition) Option {
	return func(config *Config) {
		config.WindowPosition = position
	}
}

// WithOutput sets the output writer used by Bubbletea.
func WithOutput(output io.Writer) Option {
	return func(config *Config) {
		config.Output = output
	}
}
