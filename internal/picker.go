package internal

import (
	"fmt"
	"maps"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker"
)

func Pick(query string, candidates []Candidate, options ...picker.Option) (int, bool, error) {
	items := make([]picker.Item, len(candidates))
	for i := range candidates {
		items[i] = pickerItem(candidates[i])
	}

	result, err := picker.Run(items, query, options...)
	if err != nil {
		return -1, false, err
	}
	if !result.OK {
		return -1, false, nil
	}

	if result.Index < 0 || result.Index >= len(candidates) {
		return -1, false, nil
	}

	return result.Index, true, nil
}

func pickerItem(candidate Candidate) picker.Item {
	meta := maps.Clone(candidate.Meta)
	if meta == nil {
		meta = map[string]string{}
	}
	if meta[CandidateMetaSource] == CandidateSourceRemote {
		setMetaIfMissing(meta, CandidateMetaKind, CandidateSourceRemote)
	} else {
		setMetaIfMissing(meta, CandidateMetaKind, CandidateSourceLocal)
	}

	value := meta[CandidateMetaLabel]
	if value == "" {
		value = fallbackCandidateValue(candidate)
	}

	return picker.Item{
		Value:   value,
		Signals: maps.Clone(candidate.Signals),
		Meta:    meta,
	}
}

func fallbackCandidateValue(candidate Candidate) string {
	meta := candidate.Meta
	if meta[CandidateMetaSource] == CandidateSourceRemote {
		label := meta[CandidateMetaFullName]
		if label == "" {
			label = meta[CandidateMetaName]
		}
		if label == "" {
			label = "remote"
		}
		return fmt.Sprintf("%s (remote)", label)
	}

	name := meta[CandidateMetaName]
	path := meta[CandidateMetaPath]
	if name == "" && path == "" {
		return "candidate"
	}
	if name == "" {
		return path
	}
	if path == "" {
		return name
	}

	return fmt.Sprintf("%s (%s)", name, path)
}

func setMetaIfMissing(meta map[string]string, key, value string) {
	if value == "" {
		return
	}
	if meta[key] != "" {
		return
	}

	meta[key] = value
}
