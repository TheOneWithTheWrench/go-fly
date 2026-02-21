package picker

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestFilterWithInitialQuery(t *testing.T) {
	model := newModel([]string{"alpha", "beta", "gamma"}, "et")

	assert.Equal(t, []string{"beta"}, model.FilteredItems())
}

func TestSortLessOrdersItems(t *testing.T) {
	model := newModel(
		[]string{"beta", "alpha", "gamma"},
		"",
		WithSorting(func(a, b string) bool {
			return a > b
		}),
	)

	assert.Equal(t, []string{"gamma", "beta", "alpha"}, model.FilteredItems())
}

func TestCursorNavigation(t *testing.T) {
	model := newModel([]string{"alpha", "beta", "gamma"}, "")

	model = sendKey(model, tea.KeyDown)
	assert.Equal(t, 1, model.CursorIndex())

	model = sendKey(model, tea.KeyCtrlP)
	assert.Equal(t, 0, model.CursorIndex())
}

func TestTypingJKUpdatesQuery(t *testing.T) {
	model := newModel([]string{"alpha", "beta", "gamma"}, "")

	model = sendKeyRune(model, 'j')
	assert.Equal(t, "j", model.Query())

	model = sendKeyRune(model, 'k')
	assert.Equal(t, "jk", model.Query())
}

func TestPageNavigation(t *testing.T) {
	model := newModel([]string{"a", "b", "c", "d", "e"}, "")
	model.height = 5

	model = sendKey(model, tea.KeyCtrlD)
	assert.Equal(t, 1, model.CursorIndex())

	model = sendKey(model, tea.KeyCtrlU)
	assert.Equal(t, 0, model.CursorIndex())
}

func TestSelectReturnsResult(t *testing.T) {
	model := newModel([]string{"alpha", "beta"}, "")

	model = sendKey(model, tea.KeyDown)
	model = sendKey(model, tea.KeyEnter)

	result := model.Result()
	assert.True(t, result.OK)
	assert.Equal(t, "beta", result.Value)
}

func TestCancelReturnsNotOk(t *testing.T) {
	model := newModel([]string{"alpha", "beta"}, "")

	model = sendKey(model, tea.KeyEsc)

	result := model.Result()
	assert.False(t, result.OK)
	assert.Empty(t, result.Value)
}

func TestSelectReturnsOriginalIndexAfterSorting(t *testing.T) {
	model := newModel(
		[]string{"beta", "alpha"},
		"",
		WithSorting(func(a, b string) bool {
			return a < b
		}),
	)

	model = sendKey(model, tea.KeyEnter)

	result := model.Result()
	assert.True(t, result.OK)
	assert.Equal(t, 1, result.Index)
	assert.Equal(t, "alpha", result.Value)
}

func TestWindowPositionBottomPlacesInputAfterList(t *testing.T) {
	model := newModel(
		[]string{"alpha", "beta"},
		"",
		WithWindowPosition(WindowBottom),
	)

	lines := strings.Split(model.View(), "\n")
	requireLineCount(t, lines, 4)
	assert.Contains(t, lines[0], "beta")
	assert.Contains(t, lines[1], "alpha")
	assert.Contains(t, lines[2], model.input.View())
	assert.Contains(t, lines[3], helpLine)
}

func TestWindowBottomNavigationKeepsDirection(t *testing.T) {
	model := newModel(
		[]string{"alpha", "beta", "gamma"},
		"",
		WithWindowPosition(WindowBottom),
	)

	model = sendKey(model, tea.KeyDown)
	assert.Equal(t, 0, model.CursorIndex())

	model = sendKey(model, tea.KeyUp)
	assert.Equal(t, 1, model.CursorIndex())
}

func TestViewIncludesTitleAndHelp(t *testing.T) {
	model := newModel(
		[]string{"alpha"},
		"",
		WithTitle("Pick a repo"),
	)

	lines := strings.Split(model.View(), "\n")
	requireLineCount(t, lines, 4)
	assert.Contains(t, lines[0], "Pick a repo")
	assert.Contains(t, lines[1], model.input.View())
	assert.Contains(t, lines[2], "alpha")
	assert.Contains(t, lines[3], helpLine)
}

func TestViewHighlightsMatch(t *testing.T) {
	model := newModel([]string{"alpha"}, "alp")

	view := model.View()
	assert.Contains(t, view, highlightStyle.Render("alp"))
}

func sendKey(model Model, key tea.KeyType) Model {
	msg := tea.KeyMsg{Type: key}
	updated, _ := model.Update(msg)
	return updated.(Model)
}

func sendKeyRune(model Model, key rune) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{key}}
	updated, _ := model.Update(msg)
	return updated.(Model)
}

func requireLineCount(t *testing.T, lines []string, count int) {
	t.Helper()
	if len(lines) != count {
		t.Fatalf("expected %d lines, got %d", count, len(lines))
	}
}
