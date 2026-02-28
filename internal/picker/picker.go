package picker

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/item"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/minipick"
)

type Result struct {
	Index int
	Value string
	Item  Item
	OK    bool
}

type Item = item.Item

func defaultConfig() Config {
	return Config{
		Prompt:         "> ",
		Matcher:        orderedchars.New(),
		Sorter:         minipick.New(),
		WindowPosition: WindowTop,
		Output:         defaultOutput,
	}
}

func Run(items []Item, query string, opts ...Option) (Result, error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	pickerModel := newModel(items, query, config)
	program := tea.NewProgram(
		pickerModel,
		tea.WithOutput(config.Output),
	)

	final, err := program.Run()
	if err != nil {
		return Result{}, err
	}

	finalModel, ok := final.(Model)
	if !ok {
		return Result{}, fmt.Errorf("unexpected model type")
	}

	result := finalModel.Result()
	return Result{Index: result.Index, Value: result.Value, Item: result.Item, OK: result.OK}, nil
}

func newModel(items []Item, query string, config Config) Model {
	input := textinput.New()
	input.Prompt = config.Prompt
	input.SetValue(query)
	_ = input.Focus()

	styles := input.Styles()
	styles.Focused.Prompt = cursorStyle
	styles.Focused.Text = inputStyle
	styles.Blurred.Prompt = cursorStyle
	styles.Blurred.Text = inputStyle
	styles.Cursor.Color = paletteAccentCyan
	input.SetStyles(styles)

	return newModelWithConfig(items, config, input)
}
