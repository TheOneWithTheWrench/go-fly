package local

import (
	"os"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func newDetachedPruner() internal.Pruner {
	return internal.PrunerFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		_ = internal.StartDetached(exe, "_prune")
	})
}
