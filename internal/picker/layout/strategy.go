package layout

type WindowPosition string

const (
	WindowTop    WindowPosition = "top"
	WindowBottom WindowPosition = "bottom"
)

type Model interface {
	MoveCursorUp()
	MoveCursorDown()
	PageUp()
	PageDown()
	RenderList(reverse bool) []string
}

type Strategy interface {
	MoveUp(Model)
	MoveDown(Model)
	PageUp(Model)
	PageDown(Model)
	RenderList(Model) []string
}

func NewStrategy(position WindowPosition) Strategy {
	if position == WindowBottom {
		return Bottom{}
	}
	return Top{}
}
