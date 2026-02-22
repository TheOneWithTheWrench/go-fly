package minipick_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/substring"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/minipick"
	"github.com/stretchr/testify/assert"
)

func TestMiniPickSorter(t *testing.T) {
	var (
		newSut = func() sorters.Sorter {
			return minipick.New()
		}
	)

	t.Run("sort by width then start", func(t *testing.T) {
		var (
			sut     = newSut()
			matcher = orderedchars.New()
			query   = "ap"
			items   = []sorters.Item{
				{Index: 0, Value: "alpha"},
				{Index: 1, Value: "zap"},
				{Index: 2, Value: "apple"},
			}
		)

		result := sut.Sort(query, items, matcher)

		assert.Equal(t, []int{2, 1, 0}, indexes(result))
	})

	t.Run("sort by earliest start when widths tie", func(t *testing.T) {
		var (
			sut     = newSut()
			matcher = substring.New()
			query   = "alp"
			items   = []sorters.Item{
				{Index: 0, Value: "xalp"},
				{Index: 1, Value: "alpha"},
			}
		)

		result := sut.Sort(query, items, matcher)

		assert.Equal(t, []int{1, 0}, indexes(result))
	})

	t.Run("stable for equal scores", func(t *testing.T) {
		var (
			sut     = newSut()
			matcher = substring.New()
			query   = "alp"
			items   = []sorters.Item{
				{Index: 0, Value: "alpha"},
				{Index: 1, Value: "alpha"},
			}
		)

		result := sut.Sort(query, items, matcher)

		assert.Equal(t, []int{0, 1}, indexes(result))
	})
}

func indexes(items []sorters.Item) []int {
	result := make([]int, len(items))
	for i, item := range items {
		result[i] = item.Index
	}
	return result
}
