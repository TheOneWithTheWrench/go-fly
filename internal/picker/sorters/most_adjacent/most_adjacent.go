// Package most_adjacent sorts matches by their most contiguous match span.
//
// It prefers candidates with the largest contiguous matched range and breaks
// ties by shorter candidate length.
package most_adjacent

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

// Sort orders by the longest contiguous match range.
//
// Example:
//
//	Query: "ap" (ordered-chars matcher)
//	Candidates:
//	  - "zap"   -> max range 2 ("ap" adjacent)
//	  - "apple" -> max range 2 ("ap" adjacent)
//	  - "alpha" -> max range 1 ("a" then "p" separated)
//	Ordering rule: larger max range first, then shorter candidate length.
//	Sorted order: "zap", "apple", "alpha"
func (s Sorter) Sort(query string, items []sorters.Item, matcher matchers.Matcher) []sorters.Item {
	if len(items) <= 1 {
		return items
	}
	if strings.TrimSpace(query) == "" {
		return items
	}

	scores := make([]score, len(items))
	for i, item := range items {
		match := matcher.Match(query, item)
		scores[i] = scoreFor(item, match)
	}

	slices.SortStableFunc(scores, func(a, b score) int {
		if a.Score != b.Score {
			return cmp.Compare(b.Score, a.Score) // a and b swapped because we want highest score at the top, not lowest score.
		}

		return cmp.Compare(len(a.Item.Value), len(b.Item.Value))
	})

	result := make([]sorters.Item, len(items))
	for i, entry := range scores {
		result[i] = entry.Item
	}

	return result
}

type score struct {
	Item  sorters.Item
	Score int
}

func scoreFor(item sorters.Item, match matchers.Match) score {
	var (
		scoreValue = -1
		matched    = match.Matched
		ranges     = match.Ranges
	)

	if !matched {
		return score{Item: item, Score: math.MinInt}
	}

	for _, matchRange := range ranges {
		scoreValue = max(scoreValue, matchRange.End-matchRange.Start)
	}

	return score{Item: item, Score: scoreValue}
}
