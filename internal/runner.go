package internal

import "context"

type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}
