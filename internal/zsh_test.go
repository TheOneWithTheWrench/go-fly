package internal_test

import (
	"testing"

	"github.com/TheOneWithTheWrench/go-fly/internal"
	"github.com/stretchr/testify/assert"
)

func TestZshInitSnippet(t *testing.T) {
	newSut := func() func() string { return internal.ZshInitSnippet }

	t.Run("guards help flags", func(t *testing.T) {
		sut := newSut()

		got := sut()

		assert.Contains(t, got, "-h|--help")
	})

	t.Run("only cd when directory exists", func(t *testing.T) {
		sut := newSut()

		got := sut()

		assert.Contains(t, got, "-d \"$dest\"")
	})
}

func TestZshInitSnippetChildGuard(t *testing.T) {
	newSut := func() func() string { return internal.ZshInitSnippet }

	t.Run("skip startup tracking in fly child processes", func(t *testing.T) {
		sut := newSut()

		got := sut()

		assert.Contains(t, got, `if [[ -z "$`+internal.ChildEnvVar+`" ]]; then`)
	})

	t.Run("never return from the sourced rc file", func(t *testing.T) {
		sut := newSut()

		got := sut()

		assert.NotContains(t, got, "\n  return 0\n", "a bare return aborts the rest of .zshrc")
	})
}
