# keepawake

A small Go library that prevents idle system sleep while your application is working. No helper executables or cgo required.

```sh
go get github.com/rumpl/keepawake
```

```go
release, err := keepawake.Acquire(ctx)
if err != nil {
    return err
}
defer release()

// Do work. Call release when finished.
```

`Acquire` returns an idempotent, concurrency-safe release function. Check its error when explicit cleanup matters. The context controls acquisition only: cancellation after acquisition does not release the inhibitor. Always call the release function; the OS also cleans up when the process exits.

## Platforms

- **macOS:** IOKit `IOPMAssertionCreateWithName`, loaded through [purego](https://github.com/ebitengine/purego).
- **Linux:** systemd-logind's D-Bus `Inhibit` API. Requires a running system bus and logind, Unix FD passing, and permission to acquire an `idle:sleep` inhibitor. Acquisition returns an error if unavailable, including in many containers and headless environments.
- **Windows:** `SetThreadExecutionState` on a dedicated OS thread.
- Other platforms return an unsupported-platform error.

The display can still turn off. This is not a guarantee against explicit sleep, lid closure, shutdown, or administrator policy. Linux's block inhibitor also participates in logind-managed sleep requests, subject to policy and privileged overrides.

Requires Go 1.26 or newer.

## Development

```sh
go test -race ./...
go vet ./...
KEEPAWAKE_INTEGRATION=1 go test -race ./...
```

The integration test briefly acquires a real power assertion. Run it on each supported OS; Linux needs access to logind.
