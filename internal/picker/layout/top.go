package layout

type Top struct{}

func (Top) MoveUp(m Model) {
	m.MoveCursorUp()
}

func (Top) MoveDown(m Model) {
	m.MoveCursorDown()
}

func (Top) PageUp(m Model) {
	m.PageUp()
}

func (Top) PageDown(m Model) {
	m.PageDown()
}

func (Top) RenderList(m Model) []string {
	return m.RenderList(false)
}
