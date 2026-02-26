package by_signal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal/picker/matchers/orderedchars"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/by_signal"
	"github.com/TheOneWithTheWrench/go-fly/internal/picker/sorters/minipick"
	"github.com/stretchr/testify/assert"
)

func TestSorter(t *testing.T) {
	t.Run("sort by signal descending", func(t *testing.T) {
		var (
			sut   = by_signal.New("source.zoxide.score", nil)
			items = []sorters.Item{
				{Index: 0, Value: "alpha", Signals: map[string]float64{"source.zoxide.score": 10}},
				{Index: 1, Value: "beta", Signals: map[string]float64{"source.zoxide.score": 30}},
				{Index: 2, Value: "gamma", Signals: map[string]float64{"source.zoxide.score": 20}},
			}
		)

		result := sut.Sort("", items, orderedchars.New())

		assert.Equal(t, []int{1, 2, 0}, indexes(result))
	})

	t.Run("prefer entries with signal over missing", func(t *testing.T) {
		var (
			sut   = by_signal.New("source.zoxide.score", nil)
			items = []sorters.Item{
				{Index: 0, Value: "alpha"},
				{Index: 1, Value: "beta", Signals: map[string]float64{"source.zoxide.score": 1}},
			}
		)

		result := sut.Sort("", items, orderedchars.New())

		assert.Equal(t, []int{1, 0}, indexes(result))
	})

	t.Run("use fallback order for equal signals", func(t *testing.T) {
		var (
			sut     = by_signal.New("source.zoxide.score", minipick.New())
			matcher = orderedchars.New()
			items   = []sorters.Item{
				{Index: 0, Value: "alpha", Signals: map[string]float64{"source.zoxide.score": 1}},
				{Index: 1, Value: "zap", Signals: map[string]float64{"source.zoxide.score": 1}},
				{Index: 2, Value: "apple", Signals: map[string]float64{"source.zoxide.score": 1}},
			}
		)

		result := sut.Sort("ap", items, matcher)

		assert.Equal(t, []int{2, 1, 0}, indexes(result))
	})
}

func indexes(items []sorters.Item) []int {
	result := make([]int, len(items))
	for i, item := range items {
		result[i] = item.Index
	}

	return result
}
