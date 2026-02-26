package by_signal

import (
	"cmp"
	"slices"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
)

type Sorter struct {
	key      string
	fallback sorters.Sorter
}

func New(key string, fallback sorters.Sorter) sorters.Sorter {
	return Sorter{key: key, fallback: fallback}
}

func (s Sorter) Sort(query string, items []sorters.Item, matcher matchers.Matcher) []sorters.Item {
	if len(items) <= 1 {
		return items
	}

	ordered := slices.Clone(items)
	if s.fallback != nil {
		ordered = s.fallback.Sort(query, ordered, matcher)
	}

	slices.SortStableFunc(ordered, func(a, b sorters.Item) int {
		aScore, aHas := scoreFor(a, s.key)
		bScore, bHas := scoreFor(b, s.key)

		if aHas != bHas {
			if aHas {
				return -1
			}
			return 1
		}
		if aScore != bScore {
			return cmp.Compare(bScore, aScore)
		}

		return 0
	})

	return ordered
}

func scoreFor(item sorters.Item, key string) (float64, bool) {
	if item.Signals == nil {
		return 0, false
	}

	score, ok := item.Signals[key]
	return score, ok
}
