package orderedchars_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/item"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/stretchr/testify/assert"
)

func TestOrderedCharsMatcher(t *testing.T) {
	var (
		newSut = func() matchers.Matcher {
			return orderedchars.New()
		}
	)

	t.Run("match ordered characters", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("ap", item.Item{Value: "alpha"})
		)

		assert.True(t, match.Matched)
		assert.Equal(t, []matchers.Range{{Start: 0, End: 1}, {Start: 2, End: 3}}, match.Ranges)
	})

	t.Run("match case insensitive", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("AP", item.Item{Value: "alpha"})
		)

		assert.True(t, match.Matched)
		assert.Equal(t, []matchers.Range{{Start: 0, End: 1}, {Start: 2, End: 3}}, match.Ranges)
	})

	t.Run("merge adjacent ranges", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("al", item.Item{Value: "alpha"})
		)

		assert.True(t, match.Matched)
		assert.Equal(t, []matchers.Range{{Start: 0, End: 2}}, match.Ranges)
	})

	t.Run("no match", func(t *testing.T) {
		var (
			sut   = newSut()
			match = sut.Match("pa", item.Item{Value: "map"})
		)

		assert.False(t, match.Matched)
		assert.Len(t, match.Ranges, 0)
	})
}
