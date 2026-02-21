package layout

type Bottom struct{}

func (Bottom) MoveUp(m Model) {
	m.MoveCursorDown()
}

func (Bottom) MoveDown(m Model) {
	m.MoveCursorUp()
}

func (Bottom) PageUp(m Model) {
	m.PageDown()
}

func (Bottom) PageDown(m Model) {
	m.PageUp()
}

func (Bottom) RenderList(m Model) []string {
	return m.RenderList(true)
}
