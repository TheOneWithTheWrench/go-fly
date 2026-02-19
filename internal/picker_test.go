package internal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
)

func TestLabel(t *testing.T) {
	t.Run("return candidate label", func(t *testing.T) {
		got := internal.CandidateLabel(internal.Candidate{
			Kind:  internal.KindLocal,
			Local: internal.Entry{Name: "repo", Path: "/tmp/repo"},
		})

		assert.NotEmpty(t, got)
	})
}
