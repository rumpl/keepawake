// Package keepawake inhibits idle system sleep without preventing display sleep.
package keepawake

import (
	"context"
	"fmt"
	"sync"
)

// Acquire holds an OS power assertion until release is called. The assertion
// outlives ctx; the OS also releases it when the process exits.
func Acquire(ctx context.Context) (release func() error, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	release, err = acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("keep awake: %w", err)
	}
	return sync.OnceValue(release), nil
}
