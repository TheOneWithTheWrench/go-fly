package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickerItem(t *testing.T) {
	t.Run("build local picker item", func(t *testing.T) {
		candidate := Candidate{
			Signals: map[string]float64{"source.zoxide": 12.4},
			Meta: map[string]string{
				CandidateMetaSource: CandidateSourceLocal,
				CandidateMetaName:   "repo",
				CandidateMetaPath:   "/tmp/repo",
			},
		}

		got := pickerItem(candidate)

		assert.Equal(t, "repo (/tmp/repo)", got.Value)
		assert.Equal(t, "local", got.Meta[CandidateMetaKind])
		assert.Equal(t, "repo", got.Meta[CandidateMetaName])
		assert.Equal(t, "/tmp/repo", got.Meta[CandidateMetaPath])
		assert.Equal(t, 12.4, got.Signals["source.zoxide"])
	})

	t.Run("prefer label from candidate meta", func(t *testing.T) {
		candidate := Candidate{
			Meta: map[string]string{
				CandidateMetaLabel:    "acme/repo (zoxide #1)",
				CandidateMetaSource:   CandidateSourceRemote,
				CandidateMetaFullName: "acme/repo",
			},
		}

		got := pickerItem(candidate)

		assert.Equal(t, "acme/repo (zoxide #1)", got.Value)
		assert.Equal(t, "remote", got.Meta[CandidateMetaKind])
		assert.Equal(t, "acme/repo", got.Meta[CandidateMetaFullName])
	})
}
