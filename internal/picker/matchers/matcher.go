package matchers

type Matcher interface {
	Match(query, candidate string) Match
}

type Match struct {
	Matched bool
	Ranges  []Range
}

type Range struct {
	Start int
	End   int
}
