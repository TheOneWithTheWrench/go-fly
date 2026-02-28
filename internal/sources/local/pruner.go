package local

import (
	"io"
	"os"
	"os/exec"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func newDetachedPruner() internal.Pruner {
	return internal.PrunerFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		cmd := exec.Command(exe, "_prune")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			return
		}

		_ = cmd.Process.Release()
	})
}
