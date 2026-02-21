package internal

import (
	"os"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker"
)

func Pick(query string, candidates []Candidate) (Candidate, bool, error) {
	items := make([]string, len(candidates))
	for i := range candidates {
		items[i] = CandidateLabel(candidates[i])
	}

	result, err := picker.Run(items, query,
		picker.WithOutput(os.Stderr),
		picker.WithWindowPosition(picker.WindowTop),
	)
	if err != nil {
		return Candidate{}, false, err
	}
	if !result.OK {
		return Candidate{}, false, nil
	}

	if result.Index < 0 || result.Index >= len(candidates) {
		return Candidate{}, false, nil
	}

	return candidates[result.Index], true, nil
}
