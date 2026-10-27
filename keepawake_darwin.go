package keepawake

import (
	"context"
	"errors"
	"fmt"

	"github.com/ebitengine/purego"
)

func acquire(_ context.Context) (func() error, error) {
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return nil, err
	}
	io, err := purego.Dlopen("/System/Library/Frameworks/IOKit.framework/IOKit", purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return nil, errors.Join(err, purego.Dlclose(cf))
	}
	closeLibraries := func() error {
		return errors.Join(purego.Dlclose(io), purego.Dlclose(cf))
	}
	var createString func(uintptr, string, uint32) uintptr
	var releaseString func(uintptr)
	var createAssertion func(uintptr, uint32, uintptr, *uint32) int32
	var releaseAssertion func(uint32) int32
	for _, symbol := range []struct {
		library uintptr
		name    string
		target  any
	}{
		{
			cf,
			"CFStringCreateWithCString",
			&createString,
		},
		{
			cf,
			"CFRelease",
			&releaseString,
		},
		{
			io,
			"IOPMAssertionCreateWithName",
			&createAssertion,
		},
		{
			io,
			"IOPMAssertionRelease",
			&releaseAssertion,
		},
	} {
		address, err := purego.Dlsym(symbol.library, symbol.name)
		if err != nil {
			return nil, errors.Join(err, closeLibraries())
		}
		purego.RegisterFunc(symbol.target, address)
	}

	const utf8 = 0x08000100
	kind := createString(0, "PreventUserIdleSystemSleep", utf8)
	reason := createString(0, "Application is working", utf8)
	var id uint32
	code := createAssertion(kind, 255, reason, &id)
	releaseString(kind)
	releaseString(reason)
	if code != 0 {
		return nil, errors.Join(fmt.Errorf("IOPMAssertionCreateWithName: %d", code), closeLibraries())
	}
	return func() error {
		code := releaseAssertion(id)
		closeErr := closeLibraries()
		if code != 0 {
			return errors.Join(fmt.Errorf("IOPMAssertionRelease: %d", code), closeErr)
		}
		return closeErr
	}, nil
}
