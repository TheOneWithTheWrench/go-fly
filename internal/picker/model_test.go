package picker

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModel(t *testing.T) {
	var (
		newTestModelWithItems = func(items []Item, query string, opts ...Option) Model {
			config := defaultConfig()
			for _, opt := range opts {
				opt(&config)
			}

			return newModel(items, query, config)
		}
		newTestModel = func(items []string, query string, opts ...Option) Model {
			pickerItems := make([]Item, len(items))
			for i, value := range items {
				pickerItems[i] = Item{Value: value}
			}

			return newTestModelWithItems(pickerItems, query, opts...)
		}
		sendKey = func(model Model, key tea.KeyPressMsg) Model {
			msg := key
			updated, _ := model.Update(msg)
			return updated.(Model)
		}
		sendKeyRune = func(model Model, key rune) Model {
			msg := tea.KeyPressMsg{Code: key, Text: string(key)}
			updated, _ := model.Update(msg)
			return updated.(Model)
		}
	)

	t.Run("filter with initial query", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta", "gamma"}
			model = newTestModel(items, "et")
		)

		assert.Equal(t, []string{"beta"}, model.FilteredItems())
	})

	t.Run("cursor navigation", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta", "gamma"}
			model = newTestModel(items, "")
		)

		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyDown})
		assert.Equal(t, 1, model.CursorIndex())

		model = sendKey(model, tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
		assert.Equal(t, 0, model.CursorIndex())
	})

	t.Run("typing j and k updates query", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta", "gamma"}
			model = newTestModel(items, "")
		)

		model = sendKeyRune(model, 'j')
		assert.Equal(t, "j", model.Query())

		model = sendKeyRune(model, 'k')
		assert.Equal(t, "jk", model.Query())
	})

	t.Run("page navigation", func(t *testing.T) {
		var (
			items = []string{"a", "b", "c", "d", "e"}
			model = newTestModel(items, "")
		)
		model.height = 5

		model = sendKey(model, tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
		assert.Equal(t, 1, model.CursorIndex())

		model = sendKey(model, tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
		assert.Equal(t, 0, model.CursorIndex())
	})

	t.Run("select returns result", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta"}
			model = newTestModel(items, "")
		)

		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyDown})
		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyEnter})

		result := model.Result()
		assert.True(t, result.OK)
		assert.Equal(t, "beta", result.Value)
	})

	t.Run("cancel returns not ok", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta"}
			model = newTestModel(items, "")
		)

		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyEsc})

		result := model.Result()
		assert.False(t, result.OK)
		assert.Empty(t, result.Value)
	})

	t.Run("select keeps item metadata and signals", func(t *testing.T) {
		items := []Item{{
			Value:   "alpha",
			Signals: map[string]float64{"source.zoxide": 91.3},
			Meta: map[string]string{
				"kind": "local",
				"path": "/tmp/alpha",
			},
		}}

		model := newTestModelWithItems(items, "")
		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyEnter})

		result := model.Result()
		require.True(t, result.OK)
		assert.Equal(t, "alpha", result.Value)
		assert.Equal(t, "local", result.Item.Meta["kind"])
		assert.Equal(t, 91.3, result.Item.Signals["source.zoxide"])
	})

	t.Run("window bottom places input after list", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta"}
			model = newTestModel(
				items,
				"",
				WithWindowPosition(WindowBottom),
			)
		)

		lines := strings.Split(model.View().Content, "\n")
		require.Len(t, lines, 4)
		assert.Contains(t, lines[0], "beta")
		assert.Contains(t, lines[1], "alpha")
		assert.Contains(t, lines[2], model.input.View())
		assert.Contains(t, lines[3], helpLine)
	})

	t.Run("window bottom navigation keeps direction", func(t *testing.T) {
		var (
			items = []string{"alpha", "beta", "gamma"}
			model = newTestModel(
				items,
				"",
				WithWindowPosition(WindowBottom),
			)
		)

		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyDown})
		assert.Equal(t, 0, model.CursorIndex())

		model = sendKey(model, tea.KeyPressMsg{Code: tea.KeyUp})
		assert.Equal(t, 1, model.CursorIndex())
	})

	t.Run("view includes title and help", func(t *testing.T) {
		var (
			items = []string{"alpha"}
			model = newTestModel(
				items,
				"",
				WithTitle("Pick a repo"),
			)
		)

		lines := strings.Split(model.View().Content, "\n")
		require.Len(t, lines, 4)
		assert.Contains(t, lines[0], "Pick a repo")
		assert.Contains(t, lines[1], model.input.View())
		assert.Contains(t, lines[2], "alpha")
		assert.Contains(t, lines[3], helpLine)
	})

	t.Run("view highlights match", func(t *testing.T) {
		var (
			items = []string{"alpha"}
			model = newTestModel(items, "alp")
		)

		view := model.View().Content
		assert.Contains(t, view, highlightStyle.Render("alp"))
	})

	t.Run("ordered matcher highlights matches", func(t *testing.T) {
		var (
			items = []string{"alpha"}
			model = newTestModel(
				items,
				"ap",
				WithMatcher(orderedchars.New()),
			)
		)

		view := model.View().Content
		assert.Contains(t, view, highlightStyle.Render("a"))
		assert.Contains(t, view, highlightStyle.Render("p"))
	})
}
