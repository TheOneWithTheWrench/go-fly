package picker

import (
	"io"
	"os"
)

// Option configures the picker behavior.
type Option func(*Config)

// SortLessFunc defines ordering for items. Return true if a < b.
type SortLessFunc func(string, string) bool

// Config holds the picker configuration.
type Config struct {
	Title          string
	Prompt         string
	SortLess       SortLessFunc
	WindowPosition WindowPosition
	Output         io.Writer
}

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

// WithSorting sets the sorting comparator for items.
func WithSorting(fn SortLessFunc) Option {
	return func(config *Config) {
		config.SortLess = fn
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

var defaultOutput io.Writer = os.Stdout
