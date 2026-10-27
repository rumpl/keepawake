package keepawake

import (
	"context"
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"
)

func acquire(_ context.Context) (func() error, error) {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("SetThreadExecutionState")
	if err := proc.Find(); err != nil {
		return nil, err
	}
	started := make(chan error, 1)
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		// Execution state belongs to an OS thread, not a Go goroutine.
		runtime.LockOSThread()
		// Exiting without UnlockOSThread destroys the dedicated thread,
		// ensuring its power state cannot leak into the runtime's pool.
		const continuous = 0x80000000
		const systemRequired = 0x00000001
		if result, _, err := proc.Call(continuous | systemRequired); result == 0 {
			started <- fmt.Errorf("SetThreadExecutionState: %w", err)
			return
		}
		started <- nil
		<-stop
		if result, _, err := proc.Call(continuous); result == 0 {
			done <- fmt.Errorf("SetThreadExecutionState: %w", err)
			return
		}
		done <- nil
	}()
	if err := <-started; err != nil {
		return nil, err
	}
	return func() error {
		close(stop)
		return <-done
	}, nil
}
