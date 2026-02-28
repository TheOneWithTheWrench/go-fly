package remote

import (
	"io"
	"os"
	"os/exec"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func newDetachedRefresher() internal.Refresher {
	return internal.RefresherFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		cmd := exec.Command(exe, "refresh")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if err := cmd.Start(); err != nil {
			return
		}

		_ = cmd.Process.Release()
	})
}
