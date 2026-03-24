# go-mpv

Go bindings for [libmpv](https://mpv.io/).

> Why reinventing the wheel?
> 
> Because the old wheel didn't work for me
>
> And enhance memory safety

----

## install

linux
```
sudo apt install libmpv-dev
go get github.com/aynakeya/go-mpv
```


macos
```
brew install mpv
go get github.com/aynakeya/go-mpv
```

windows
1. compile or download mpv-2.dll.
2. copy mpv-2.dll and mpv folder to system path
```
go get github.com/aynakeya/go-mpv
```

# Known bugs

## Node memory ownership and safety

`mpv_node` memory ownership is split in two cases:

1. Nodes returned by libmpv APIs (for example `mpv_get_property(..., MPV_FORMAT_NODE, ...)` and command results):
- Release with `mpv_free_node_contents()`.
- Do not write into pointers owned by libmpv.

2. Nodes created by this binding (for example from `Node.CNode()` for set/command input):
- Release with binding-owned cleanup (`free_node()` via `freeMpvDataPointer(FORMAT_NODE, ...)`).
- Do not pass these nodes to `mpv_free_node_contents()` directly.

Reason:
- libmpv can only safely free memory it allocated itself.
- binding-created node trees use custom allocations (including nested map/array keys and byte arrays), and must be released by binding-owned free logic.
