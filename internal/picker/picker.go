package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type Result struct {
	Index int
	Value string
	OK    bool
}

func newModel(items []string, query string, opts ...Option) Model {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	input := textinput.New()
	input.Prompt = config.Prompt
	input.SetValue(query)
	input.Focus()
	input.PromptStyle = cursorStyle
	input.TextStyle = inputStyle
	input.Cursor.Style = cursorStyle

	return newModelWithConfig(items, config, input)
}

func Run(items []string, query string, opts ...Option) (Result, error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}
	lipgloss.SetColorProfile(termenv.NewOutput(config.Output).Profile)

	pickerModel := newModel(items, query, opts...)
	program := tea.NewProgram(
		pickerModel,
		tea.WithOutput(config.Output),
		tea.WithAltScreen(),
	)
	final, err := program.Run()
	if err != nil {
		return Result{}, err
	}

	finalModel, ok := final.(Model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected model type")
	}

	return Result{Index: finalModel.selectedIndex, Value: finalModel.selected, OK: finalModel.ok}, nil
}

func defaultConfig() Config {
	return Config{
		Prompt:         "> ",
		WindowPosition: WindowTop,
		Output:         defaultOutput,
	}
}
