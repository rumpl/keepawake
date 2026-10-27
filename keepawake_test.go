package keepawake

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
)

func TestAcquireCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	release, err := Acquire(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if release != nil {
		t.Fatal("failed acquisition returned a release function")
	}
}

func TestAcquireNative(t *testing.T) {
	if os.Getenv("KEEPAWAKE_INTEGRATION") != "1" {
		t.Skip("set KEEPAWAKE_INTEGRATION=1 to test native power assertions")
	}
	ctx, cancel := context.WithCancel(t.Context())
	release, err := Acquire(ctx)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := release(); err != nil {
			t.Error(err)
		}
	})
	// Cancellation after acquisition must not release or invalidate the assertion.
	cancel()
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			if err := release(); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
}
