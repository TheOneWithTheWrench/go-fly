// Package minipick sorts matches by width and start position.
//
// Inspired by mini.pick's default sorting:
// https://github.com/echasnovski/mini.nvim/blob/main/lua/mini/pick.lua
//
// It prefers smaller match width first, then earlier start position, keeping
// stable order for ties.
package minipick

import (
	"cmp"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
)

type Sorter struct{}

func New() sorters.Sorter {
	return Sorter{}
}

func (Sorter) Sort(query string, items []sorters.Item, matcher matchers.Matcher) []sorters.Item {
	if len(items) <= 1 {
		return items
	}
	if strings.TrimSpace(query) == "" {
		return items
	}

	scores := make([]score, len(items))
	for i, item := range items {
		match := matcher.Match(query, item.Value)
		scores[i] = scoreFor(item, match)
	}

	slices.SortStableFunc(scores, func(a, b score) int {
		if a.Width != b.Width {
			return cmp.Compare(a.Width, b.Width)
		}
		if a.Start != b.Start {
			return cmp.Compare(a.Start, b.Start)
		}
		return 0
	})

	result := make([]sorters.Item, len(items))
	for i, entry := range scores {
		result[i] = entry.Item
	}

	return result
}

type score struct {
	Item  sorters.Item
	Width int
	Start int
}

func scoreFor(item sorters.Item, match matchers.Match) score {
	start, end, ok := matchBounds(item.Value, match.Ranges)
	if !match.Matched || !ok {
		return score{Item: item, Width: maxInt, Start: maxInt}
	}

	width := (end - 1) - start
	return score{Item: item, Width: width, Start: start}
}

func matchBounds(value string, ranges []matchers.Range) (int, int, bool) {
	if len(ranges) == 0 {
		return 0, 0, false
	}

	length := len(value)
	minStart := length
	maxEnd := -1
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
		if start < minStart {
			minStart = start
		}
		if end > maxEnd {
			maxEnd = end
		}
	}

	if maxEnd <= minStart {
		return 0, 0, false
	}

	return minStart, maxEnd, true
}

const maxInt = int(^uint(0) >> 1)
