package substring_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/substring"
	"github.com/stretchr/testify/assert"
)

func TestSubstringMatcher(t *testing.T) {
	var (
		newSut = func() matchers.Matcher {
			return substring.New()
		}
	)

	t.Run("match empty query", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("", "alpha")
		)

		assert.True(t, match.Matched)
		assert.Len(t, match.Ranges, 0)
	})

	t.Run("match substring", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("alp", "alpha")
		)

		assert.True(t, match.Matched)
		assert.Equal(t, []matchers.Range{{Start: 0, End: 3}}, match.Ranges)
	})

	t.Run("match case insensitive", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("ALP", "alpha")
		)

		assert.True(t, match.Matched)
		assert.Equal(t, []matchers.Range{{Start: 0, End: 3}}, match.Ranges)
	})

	t.Run("no match", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("zz", "alpha")
		)

		assert.False(t, match.Matched)
		assert.Len(t, match.Ranges, 0)
	})
}
