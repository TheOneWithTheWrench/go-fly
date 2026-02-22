package most_adjacent_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/most_adjacent"
	"github.com/stretchr/testify/assert"
)

func TestMostAdjacentSorter(t *testing.T) {
	var (
		newSut = func() sorters.Sorter {
			return most_adjacent.New()
		}
	)

	t.Run("sorts by max contiguous range then length", func(t *testing.T) {
		var (
			sut     = newSut()
			matcher = orderedchars.New()
			query   = "ap"
			items   = []sorters.Item{
				{Index: 0, Value: "alpha"},
				{Index: 1, Value: "zap"},
				{Index: 2, Value: "apple"},
				{Index: 3, Value: "stone"},
			}
		)

		result := sut.Sort(query, items, matcher)

		assert.Equal(t, []int{1, 2, 0, 3}, indexes(result))
	})
}

func indexes(items []sorters.Item) []int {
	result := make([]int, len(items))
	for i, item := range items {
		result[i] = item.Index
	}
	return result
}
