// Package substring matches when the query exists as a contiguous substring.
//
// Examples:
//
//	Query: "alp"
//	Candidate: "alpha"  -> match (range: [0,3))
//	Candidate: "caliper" -> match (range: [1,4))
//
// Matching is case-insensitive. Ranges are byte offsets in the candidate.
package substring

import (
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
	idx := strings.Index(haystack, needle)
	if idx < 0 {
		return matchers.Match{}
	}

	return matchers.Match{
		Matched: true,
		Ranges:  []matchers.Range{{Start: idx, End: idx + len(query)}},
	}
}
