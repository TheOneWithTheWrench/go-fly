package matchers

import "github.com/TheOneWithTheWrench/go-fly/internal/picker/item"

type Matcher interface {
	Match(query string, item item.Item) Match
}

type Match struct {
	Matched bool
	Ranges  []Range
}

type Range struct {
	Start int
	End   int
}
