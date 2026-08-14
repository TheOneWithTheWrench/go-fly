package zoxide

import (
	"bytes"
	"context"
	"fmt"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

func defaultRunner() internal.Runner {
	return internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		var stderr bytes.Buffer

		cmd := internal.NewIsolatedCommand(ctx, name, args...)
		cmd.Stderr = &stderr

		output, err := cmd.Output()
		if err != nil {
			return output, fmt.Errorf("run %s %v: %w: %s", name, args, err, stderr.Bytes())
		}

		return output, nil
	})
}
