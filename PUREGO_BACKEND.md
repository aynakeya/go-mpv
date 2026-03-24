# PureGo backend notes (`!cgo`)

This document describes the current status and safety boundaries of the `purego` backend.

## Build selection

- `cgo` backend: files with `//go:build cgo`
- `purego` backend: files with `//go:build !cgo`

Typical command:

```bash
CGO_ENABLED=0 go test ./...
```

## Current support (client API)

Implemented and tested:

- lifecycle: `Create`, `Initialize`, `Destroy`, `TerminateDestroy`
- info: `ClientApiVersion`, `ClientName`, `ClientId`, `GetTimeUS`, `GetTimeNS`
- properties/options:
  - `SetOption`, `SetOptionString`
  - `SetProperty`, `SetPropertyString`, `SetPropertyAsync`
  - `GetProperty`, `GetPropertyString`, `GetPropertyOsdString`, `GetPropertyAsync`
  - `DelProperty`, `ObserveProperty`, `UnObserveProperty`
- commands:
  - `Command`, `CommandAsync`, `CommandString`
  - `CommandRet`, `CommandNode`, `CommandNodeAsync`
- events:
  - `WaitEvent`, `Wakeup`, `WaitAsyncRequests`
  - `Event.Property`, `Event.LogMessage`, `Event.Command`, `Event.Hook`, `Event.ToNode`
- hooks: `HookAdd`, `HookContinue`

## ABI and memory boundaries

The purego backend manually mirrors C structs from `client.h`:

- `mpv_event`
- `mpv_event_property`
- `mpv_event_log_message`
- `mpv_event_hook`
- `mpv_event_command`
- `mpv_node`
- `mpv_node_list`
- `mpv_byte_array`

Key rules:

1. Values returned by libmpv are released with libmpv APIs:
- strings via `mpv_free`
- node trees via `mpv_free_node_contents`

2. Values marshaled from Go for `CommandNode` and property/option writes are held alive across the FFI call with Go references (`runtime.KeepAlive` and arena keeps).

3. The current implementation assumes the common `libmpv` ABI layout used by mainstream builds (Linux/macOS x64/arm64). If a platform ships a divergent ABI, the purego backend can break.

## Known risks

1. Struct layout drift risk
- This backend depends on exact struct/union layout compatibility with `client.h`.
- Upgrading libmpv major versions should be verified with tests on target platforms.

2. Pointer lifetime risk
- Any new FFI call that takes pointers must keep backing Go memory alive until the call returns.

3. Cross-platform DLL name variance
- Loader currently probes:
  - `libmpv.so.2`
  - `libmpv.so`
  - `libmpv.dylib`
  - `mpv-2.dll`
- If packaging uses other names/paths, loading fails.

## Recommended validation when changing purego backend

1. Run both backends:
- `go test ./...`
- `GOCACHE=/tmp/go-build CGO_ENABLED=0 go test ./...`

2. Validate on real target OS/arch pairs (not only one Linux host).

3. Prefer adding tests for every newly bound symbol before exposing it publicly.
