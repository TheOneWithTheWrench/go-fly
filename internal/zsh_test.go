package internal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
)

func TestZshInitSnippet(t *testing.T) {
	var (
		newSut = func() func() string { return internal.ZshInitSnippet }
	)

	t.Run("guards help flags", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut()

		assert.Contains(t, got, "-h|--help")
	})

	t.Run("only cd when directory exists", func(t *testing.T) {
		var (
			sut = newSut()
		)

		got := sut()

		assert.Contains(t, got, "-d \"$dest\"")
	})
}
