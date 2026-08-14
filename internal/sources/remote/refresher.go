package remote

import (
	"os"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func newDetachedRefresher() internal.Refresher {
	return internal.RefresherFunc(func() {
		exe, err := os.Executable()
		if err != nil {
			return
		}

		_ = internal.StartDetached(exe, "refresh")
	})
}
