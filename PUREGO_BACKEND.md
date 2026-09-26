# PureGo backend notes (`!cgo`)

This document describes the current status and safety boundaries of the `purego` backend.

## Build selection

- `cgo` backend: files with `//go:build cgo`
- `purego` backend: files with `//go:build !cgo`

The module requires Go 1.25 or later, matching purego v0.11.1.

Typical command:

```bash
CGO_ENABLED=0 go test ./...
```

## Current support (client API)

Implemented and tested:

- lifecycle: `Create`, `Initialize`, `Destroy`, `TerminateDestroy`
- info: `ClientApiVersion`, `ClientName`, `ClientId`, `GetTimeUS`, `GetTimeNS`
- properties/options:
  - `SetOption`, `SetOptionString` (including `FORMAT_NODE` writes)
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

## Render API

- Both backends implement render context creation, rendering, updates, callbacks, swap reporting, parameters and frame info.
- Software rendering is tested with the video in `data/test.mp4` and demonstrated in `example/render`. OpenGL integration is demonstrated in the separate `example/fyne_video_demo` module.
- This does not imply cross-platform ABI validation. Windows loading and render parameter passing still need implementation and real-platform testing.

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

`Create` keeps its nil-on-failure contract; it does not expose a creation error.
Callers must check creation results before calling methods; nil `*Mpv`
receivers are not supported. A zero-value or destroyed handle in the nocgo
backend is rejected with
`ERROR_INVALID_PARAMETER` by error-returning methods;
`WaitEvent` returns nil, not a timeout event. Neither backend permits concurrent
destruction and use of a handle.

1. Struct layout drift risk
- This backend depends on exact struct/union layout compatibility with `client.h`.
- Upgrading libmpv major versions should be verified with tests on target platforms.

2. Pointer lifetime risk
- Any new FFI call that takes pointers must keep backing Go memory alive until the call returns.
- Event payloads are borrowed until the next `WaitEvent` on the same handle or its destruction. Convert them using payload accessors or `ToNode` before handing data to another goroutine; copying `Event` does not extend the lifetime.
- Stop handle users and free render contexts before destruction. Repeated `Destroy`/`TerminateDestroy` calls are safe when serialized, not concurrently.

3. Platform loading
- Library candidates vary by OS (see `libMPVCandidates`). Packaging with other names/paths can fail to load.
- The loader still uses Unix `Dlopen`/`Dlsym`. Windows DLL candidate names alone do not make the current backend build or run on Windows.

## Recommended validation when changing purego backend

1. Run the relevant shared behavior tests on both backends. The suite requires
libmpv and the checked-in video; loading failures are test failures, not skips.
For changes spanning client and render APIs, run the complete suites:
- `go test ./...`
- `GOCACHE=/tmp/go-build CGO_ENABLED=0 go test ./...`

2. Validate on real target OS/arch pairs (not only one Linux host).

3. Prefer adding tests for every newly bound symbol before exposing it publicly.
