package internal

import (
	"errors"

	"github.com/ktr0731/go-fuzzyfinder"
)

func Pick(query string, candidates []Candidate) (Candidate, bool, error) {
	idx, err := fuzzyfinder.Find(candidates, func(i int) string {
		return CandidateLabel(candidates[i])
	}, fuzzyfinder.WithQuery(query), fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionTop))
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return Candidate{}, false, nil
		}
		return Candidate{}, false, err
	}

	return candidates[idx], true, nil
}
