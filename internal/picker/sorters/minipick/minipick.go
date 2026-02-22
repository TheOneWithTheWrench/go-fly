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
	"math"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
)

type Sorter struct{}

func New() sorters.Sorter {
	return Sorter{}
}

// Sort orders matched items by score derived from the match ranges.
//
// Width is the distance between the first and last matched character:
// width = (end - 1) - start.
//
// Example:
//
//	Query: "ap" (ordered-chars matcher)
//	Candidates:
//	  - "alpha" -> start 0, end 3, width 2
//	  - "apple" -> start 0, end 2, width 1
//	  - "zap"   -> start 1, end 3, width 1
//	Ordering rule: smaller width first, then earlier start.
//	Sorted order: "apple", "zap", "alpha"
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

// scoreFor converts match ranges into a width/start score.
func scoreFor(item sorters.Item, match matchers.Match) score {
	start, end, ok := matchBounds(item.Value, match.Ranges)
	if !match.Matched || !ok {
		return score{Item: item, Width: math.MaxInt, Start: math.MaxInt}
	}

	width := (end - 1) - start
	return score{Item: item, Width: width, Start: start}
}

// matchBounds returns the smallest start and largest end over match ranges.
//
// Example:
//
//	Query: "ap" (ordered-chars matcher)
//	Candidate: "alpha"
//	Match ranges: [0,1), [2,3) -> start 0, end 3
func matchBounds(value string, ranges []matchers.Range) (int, int, bool) {
	if len(ranges) == 0 {
		return 0, 0, false
	}

	var (
		length   = len(value)
		minStart = length
		maxEnd   = -1
	)

	for _, r := range ranges {
		start := max(0, r.Start)
		end := min(length, r.End)

		// Only process valid, non-empty ranges
		if start < end {
			minStart = min(minStart, start)
			maxEnd = max(maxEnd, end)
		}
	}

	// If no valid ranges were found, maxEnd will still be -1
	// (or at least less than or equal to minStart)
	if minStart >= maxEnd {
		return 0, 0, false
	}

	return minStart, maxEnd, true
}
