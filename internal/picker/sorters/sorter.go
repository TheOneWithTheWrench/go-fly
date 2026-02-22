package sorters

import "github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"

type Sorter interface {
	Sort(query string, items []Item, matcher matchers.Matcher) []Item
}

type Item struct {
	Index int
	Value string
}
