//go:build !darwin && !linux && !windows

package keepawake

import (
	"context"
	"fmt"
	"runtime"
)

func acquire(_ context.Context) (func() error, error) {
	return nil, fmt.Errorf("unsupported operating system %q", runtime.GOOS)
}
