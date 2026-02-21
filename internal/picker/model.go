package picker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/layout"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type WindowPosition = layout.WindowPosition

const (
	WindowTop    = layout.WindowTop
	WindowBottom = layout.WindowBottom
)

const helpLine = "enter select • esc cancel • ↑↓/ctrl+n/ctrl+p navigate"

type item struct {
	index int
	value string
}

type Model struct {
	config        Config
	input         textinput.Model
	items         []item
	filtered      []item
	cursor        int
	offset        int
	width         int
	height        int
	selected      string
	selectedIndex int
	ok            bool
	layout        layout.Strategy
}

func newModelWithConfig(items []string, config Config, input textinput.Model) Model {
	baseItems := make([]item, len(items))
	for i, value := range items {
		baseItems[i] = item{index: i, value: value}
	}
	if config.SortLess != nil {
		sort.SliceStable(baseItems, func(i, j int) bool {
			return config.SortLess(baseItems[i].value, baseItems[j].value)
		})
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
	case tea.KeyMsg:
		key := msg.String()
		if isKeyCancel(key) {
			m.ok = false
			return m, tea.Quit
		}
		if isKeySelect(key) {
			if len(m.filtered) == 0 {
				return m, nil
			}
			m.selected = m.filtered[m.cursor].value
			m.selectedIndex = m.filtered[m.cursor].index
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

func (m Model) View() string {
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
	return strings.Join(lines, "\n")
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
	return Result{Index: m.selectedIndex, Value: m.selected, OK: m.ok}
}

func (m Model) FilteredItems() []string {
	items := make([]string, len(m.filtered))
	for i, entry := range m.filtered {
		items[i] = entry.value
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
	m.filtered = defaultFilter(query, m.items)

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

func defaultFilter(query string, items []item) []item {
	if strings.TrimSpace(query) == "" {
		return items
	}

	needle := strings.ToLower(query)
	filtered := make([]item, 0, len(items))
	for _, entry := range items {
		if strings.Contains(strings.ToLower(entry.value), needle) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func renderValue(value string, query string, selected bool) string {
	baseStyle := itemStyle
	if selected {
		baseStyle = selectedStyle
	}
	if strings.TrimSpace(query) == "" {
		return baseStyle.Render(value)
	}

	needle := strings.ToLower(query)
	label := strings.ToLower(value)
	idx := strings.Index(label, needle)
	if idx < 0 {
		return baseStyle.Render(value)
	}

	start := value[:idx]
	match := value[idx : idx+len(query)]
	end := value[idx+len(query):]

	return baseStyle.Render(start) + highlightStyle.Render(match) + baseStyle.Render(end)
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
	line := renderValue(m.filtered[index].value, m.input.Value(), index == m.cursor)
	return fmt.Sprintf("%s %s", prefix, line)
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
