package keepawake

import (
	"context"
	"errors"
	"os"

	"github.com/godbus/dbus/v5"
)

func acquire(ctx context.Context) (func() error, error) {
	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		return nil, err
	}
	closeOnError := func(err error) (func() error, error) {
		return nil, errors.Join(err, conn.Close())
	}
	if err := conn.Auth(nil); err != nil {
		return closeOnError(err)
	}
	if err := conn.Hello(); err != nil {
		return closeOnError(err)
	}
	var fd dbus.UnixFD
	err = conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").
		CallWithContext(ctx, "org.freedesktop.login1.Manager.Inhibit", 0,
			"idle:sleep", "keepawake", "Application is working", "block").Store(&fd)
	if err != nil {
		return closeOnError(err)
	}
	file := os.NewFile(uintptr(fd), "sleep-inhibitor")
	return func() error {
		return errors.Join(file.Close(), conn.Close())
	}, nil
}
