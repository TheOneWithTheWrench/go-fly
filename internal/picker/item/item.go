package item

type Item struct {
	Index   int
	Value   string
	Signals map[string]float64
	Meta    map[string]string
}
