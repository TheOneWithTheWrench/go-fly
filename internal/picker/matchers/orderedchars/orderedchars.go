// Package orderedchars matches when all query characters appear in order.
//
// Examples:
//
//	Query: "ap"
//	Candidate: "alpha" -> match (ranges: [0,1), [2,3))
//	Candidate: "paper" -> match (ranges: [1,2), [2,3))
//	Candidate: "pear"  -> no match ("a" appears before "p")
//
// Matching is case-insensitive. Ranges are byte offsets in the candidate.
package orderedchars

import (
	"cmp"
	"slices"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
)

type Matcher struct{}

func New() matchers.Matcher {
	return Matcher{}
}

func (Matcher) Match(query, candidate string) matchers.Match {
	query = strings.TrimSpace(query)
	if query == "" {
		return matchers.Match{Matched: true}
	}

	needle := strings.ToLower(query)
	haystack := strings.ToLower(candidate)
	if len(needle) == 0 {
		return matchers.Match{Matched: true}
	}

	indices := make([]int, 0, len(needle))
	pos := 0
	for i := 0; i < len(needle); i++ {
		ch := needle[i]
		found := -1
		for j := pos; j < len(haystack); j++ {
			if haystack[j] == ch {
				found = j
				break
			}
		}
		if found < 0 {
			return matchers.Match{}
		}
		indices = append(indices, found)
		pos = found + 1
	}

	ranges := make([]matchers.Range, 0, len(indices))
	for _, idx := range indices {
		ranges = append(ranges, matchers.Range{Start: idx, End: idx + 1})
	}

	return matchers.Match{Matched: true, Ranges: mergeRanges(ranges)}
}

func mergeRanges(ranges []matchers.Range) []matchers.Range {
	if len(ranges) == 0 {
		return nil
	}

	sorted := append([]matchers.Range(nil), ranges...)
	slices.SortFunc(sorted, func(a, b matchers.Range) int {
		if a.Start != b.Start {
			return cmp.Compare(a.Start, b.Start)
		}
		return cmp.Compare(a.End, b.End)
	})

	merged := []matchers.Range{sorted[0]}
	for _, r := range sorted[1:] {
		last := &merged[len(merged)-1]
		if r.Start <= last.End {
			if r.End > last.End {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}

	return merged
}
