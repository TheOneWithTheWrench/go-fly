package picker

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/layout"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
)

type WindowPosition = layout.WindowPosition

const (
	WindowTop    = layout.WindowTop
	WindowBottom = layout.WindowBottom
)

const helpLine = "enter select • esc cancel • ↑↓/ctrl+n/ctrl+p navigate"

type Model struct {
	config        Config
	input         textinput.Model
	items         []sorters.Item
	filtered      []sorters.Item
	cursor        int
	offset        int
	width         int
	height        int
	selected      sorters.Item
	selectedIndex int
	ok            bool
	layout        layout.Strategy
}

func newModelWithConfig(items []Item, config Config, input textinput.Model) Model {
	baseItems := make([]sorters.Item, len(items))
	for i, value := range items {
		baseItems[i] = sorters.Item{
			Index:   i,
			Value:   value.Value,
			Signals: maps.Clone(value.Signals),
			Meta:    maps.Clone(value.Meta),
		}
	}
	if config.Matcher == nil {
		config.Matcher = orderedchars.New()
	}

	m := Model{
		config: config,
		input:  input,
		items:  baseItems,
		layout: layout.NewStrategy(config.WindowPosition),
	}
	m.applyFilter(true)
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyPressMsg:
		key := msg.String()
		if isKeyCancel(key) {
			m.ok = false
			return m, tea.Quit
		}
		if isKeySelect(key) {
			if len(m.filtered) == 0 {
				return m, nil
			}
			m.selected = m.filtered[m.cursor]
			m.selectedIndex = m.selected.Index
			m.ok = true
			return m, tea.Quit
		}
		if isKeyUp(key) {
			m.layout.MoveUp(&m)
			m.ensureOffset()
			return m, nil
		}
		if isKeyDown(key) {
			m.layout.MoveDown(&m)
			m.ensureOffset()
			return m, nil
		}
		if isKeyPageDown(key) {
			m.layout.PageDown(&m)
			return m, nil
		}
		if isKeyPageUp(key) {
			m.layout.PageUp(&m)
			return m, nil
		}

		before := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if m.input.Value() != before {
			m.applyFilter(false)
		}
		return m, cmd
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m Model) View() tea.View {
	var lines []string
	if m.config.Title != "" {
		lines = append(lines, titleStyle.Render(m.config.Title))
	}
	if m.config.WindowPosition == WindowTop {
		lines = append(lines, inputStyle.Render(m.input.View()))
	}

	lines = append(lines, m.layout.RenderList(&m)...)

	if m.config.WindowPosition == WindowBottom {
		lines = append(lines, inputStyle.Render(m.input.View()))
	}

	lines = append(lines, helpStyle.Render(helpLine))

	v := tea.NewView(strings.Join(lines, "\n"))
	v.AltScreen = true

	return v
}

func (m *Model) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *Model) MoveCursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
	}
}

func (m *Model) PageDown() {
	if len(m.filtered) == 0 {
		return
	}
	step := max(1, m.visibleCount()/2)
	m.cursor = min(len(m.filtered)-1, m.cursor+step)
	m.ensureOffset()
}

func (m *Model) PageUp() {
	if len(m.filtered) == 0 {
		return
	}
	step := max(1, m.visibleCount()/2)
	m.cursor = max(0, m.cursor-step)
	m.ensureOffset()
}

func (m *Model) RenderList(reverse bool) []string {
	return renderList(m, reverse)
}

func (m Model) Result() Result {
	if !m.ok {
		return Result{OK: false}
	}

	selected := Item{
		Index:   m.selected.Index,
		Value:   m.selected.Value,
		Signals: maps.Clone(m.selected.Signals),
		Meta:    maps.Clone(m.selected.Meta),
	}

	return Result{
		Index: m.selectedIndex,
		Value: m.selected.Value,
		Item:  selected,
		OK:    true,
	}
}

func (m Model) FilteredItems() []string {
	items := make([]string, len(m.filtered))
	for i, entry := range m.filtered {
		items[i] = entry.Value
	}
	return items
}

func (m Model) CursorIndex() int {
	return m.cursor
}

func (m Model) Query() string {
	return m.input.Value()
}

func (m *Model) applyFilter(initial bool) {
	query := m.input.Value()
	m.filtered = filterItems(query, m.items, m.config.Matcher)
	if m.config.Sorter != nil {
		m.filtered = sortItems(query, m.filtered, m.config.Matcher, m.config.Sorter)
	}

	if len(m.filtered) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}

	if !initial {
		m.cursor = 0
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.offset = 0
}

func (m *Model) ensureOffset() {
	visible := m.visibleCount()
	if visible <= 0 {
		m.offset = 0
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
		return
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
}

func (m Model) visibleRange() (int, int) {
	visible := m.visibleCount()
	if visible <= 0 || visible >= len(m.filtered) {
		return 0, len(m.filtered)
	}
	start := m.offset
	end := m.offset + visible
	if start < 0 {
		start = 0
	}
	if end > len(m.filtered) {
		end = len(m.filtered)
	}
	return start, end
}

func (m Model) visibleCount() int {
	if m.height <= 0 {
		return len(m.filtered)
	}
	lines := 2
	if m.config.Title != "" {
		lines++
	}
	visible := m.height - lines
	if visible < 1 {
		return 1
	}
	return visible
}

func filterItems(query string, items []sorters.Item, matcher matchers.Matcher) []sorters.Item {
	if strings.TrimSpace(query) == "" {
		return items
	}

	filtered := make([]sorters.Item, 0, len(items))
	for _, entry := range items {
		if matcher.Match(query, entry).Matched {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func sortItems(query string, items []sorters.Item, matcher matchers.Matcher, sorter sorters.Sorter) []sorters.Item {
	entries := make([]sorters.Item, len(items))
	for i, entry := range items {
		entries[i] = sorters.Item{
			Index:   entry.Index,
			Value:   entry.Value,
			Signals: maps.Clone(entry.Signals),
			Meta:    maps.Clone(entry.Meta),
		}
	}

	sorted := sorter.Sort(query, entries, matcher)
	if len(sorted) == 0 {
		return nil
	}

	result := make([]sorters.Item, len(sorted))
	for i, entry := range sorted {
		result[i] = sorters.Item{
			Index:   entry.Index,
			Value:   entry.Value,
			Signals: maps.Clone(entry.Signals),
			Meta:    maps.Clone(entry.Meta),
		}
	}

	return result
}

func renderValue(value string, selected bool, match matchers.Match) string {
	baseStyle := itemStyle
	if selected {
		baseStyle = selectedStyle
	}
	if !match.Matched || len(match.Ranges) == 0 {
		return baseStyle.Render(value)
	}

	length := len(value)
	ranges := normalizeRanges(match.Ranges, length)
	if len(ranges) == 0 {
		return baseStyle.Render(value)
	}

	var rendered strings.Builder
	pos := 0
	for _, r := range ranges {
		if r.Start > pos {
			rendered.WriteString(baseStyle.Render(value[pos:r.Start]))
		}
		rendered.WriteString(highlightStyle.Render(value[r.Start:r.End]))
		pos = r.End
	}
	if pos < length {
		rendered.WriteString(baseStyle.Render(value[pos:]))
	}

	return rendered.String()
}

func renderList(m *Model, reverse bool) []string {
	if len(m.filtered) == 0 {
		return []string{emptyStyle.Render("no matches")}
	}

	start, end := m.visibleRange()
	lines := make([]string, 0, end-start)
	if reverse {
		for i := end - 1; i >= start; i-- {
			lines = append(lines, renderListLine(m, i))
		}
		return lines
	}

	for i := start; i < end; i++ {
		lines = append(lines, renderListLine(m, i))
	}

	return lines
}

func renderListLine(m *Model, index int) string {
	prefix := " "
	if index == m.cursor {
		prefix = cursorStyle.Render(">")
	}
	entry := m.filtered[index]
	value := entry.Value
	match := m.config.Matcher.Match(m.input.Value(), entry)
	line := renderValue(value, index == m.cursor, match)
	return fmt.Sprintf("%s %s", prefix, line)
}

func normalizeRanges(ranges []matchers.Range, length int) []matchers.Range {
	if length <= 0 || len(ranges) == 0 {
		return nil
	}

	clamped := make([]matchers.Range, 0, len(ranges))
	for _, r := range ranges {
		start := r.Start
		end := r.End
		if start < 0 {
			start = 0
		}
		if end > length {
			end = length
		}
		if end <= start {
			continue
		}
		clamped = append(clamped, matchers.Range{Start: start, End: end})
	}

	if len(clamped) == 0 {
		return nil
	}

	return mergeRanges(clamped)
}

func mergeRanges(ranges []matchers.Range) []matchers.Range {
	if len(ranges) == 0 {
		return nil
	}

	sorted := append([]matchers.Range(nil), ranges...)
	slices.SortFunc(sorted, func(a, b matchers.Range) int {
		if a.Start != b.Start {
			return cmp.Compare(a.Start, b.Start)
		}
		return cmp.Compare(a.End, b.End)
	})

	merged := []matchers.Range{sorted[0]}
	for _, r := range sorted[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End {
			if r.End > last.End {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}

	return merged
}

func isKeyUp(value string) bool {
	return value == "up" || value == "ctrl+p"
}

func isKeyDown(value string) bool {
	return value == "down" || value == "ctrl+n"
}

func isKeySelect(value string) bool {
	return value == "enter"
}

func isKeyCancel(value string) bool {
	return value == "esc" || value == "ctrl+c"
}

func isKeyPageDown(value string) bool {
	return value == "ctrl+d"
}

func isKeyPageUp(value string) bool {
	return value == "ctrl+u"
}
