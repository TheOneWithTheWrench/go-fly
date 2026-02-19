package internal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	var (
		newSut = func() func(string, []internal.Entry) []internal.Entry {
			return internal.Filter
		}
		entries = []internal.Entry{
			{Name: "alpha", Path: "/work/alpha"},
			{Name: "beta", Path: "/work/beta"},
			{Name: "gamma", Path: "/src/gamma"},
		}
	)

	t.Run("match by name", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("AlP", entries)

		assert.Len(t, got, 1)
		assert.Equal(t, "alpha", got[0].Name)
	})

	t.Run("match by path", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("/src", entries)

		assert.Len(t, got, 1)
		assert.Equal(t, "gamma", got[0].Name)
	})

	t.Run("return all when query empty", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("", entries)

		assert.Equal(t, entries, got)
	})
}

func TestFilterRemote(t *testing.T) {
	var (
		newSut = func() func(string, []internal.Repo) []internal.Repo {
			return internal.FilterRemote
		}
		repos = []internal.Repo{
			{Name: "alpha", FullName: "acme/alpha"},
			{Name: "beta", FullName: "acme/beta"},
			{Name: "gamma", FullName: "other/gamma"},
		}
	)

	t.Run("match by full name", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("other/", repos)

		assert.Len(t, got, 1)
		assert.Equal(t, "gamma", got[0].Name)
	})

	t.Run("match by name", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("AlP", repos)

		assert.Len(t, got, 1)
		assert.Equal(t, "alpha", got[0].Name)
	})

	t.Run("return all when query empty", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut("", repos)

		assert.Equal(t, repos, got)
	})
}
