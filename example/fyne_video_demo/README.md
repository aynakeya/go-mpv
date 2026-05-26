# Fyne video render demo

This example renders libmpv video frames into Fyne's GLFW OpenGL framebuffer using the go-mpv render API.

It is a separate Go module so Fyne and GLFW dependencies do not affect the root module.

## Run

```bash
cd example/fyne_video_demo
CGO_ENABLED=1 go run . -file ../../data/test.mp4
```

You can also pass a file or URL as the first positional argument.

## Notes

- The example is limited to `linux && cgo` because it reflects Fyne's current GLFW window internals and renders directly into that OpenGL context.
- libmpv is accessed through `github.com/aynakeya/go-mpv`; the demo does not call `mpv_render_*` C functions directly.
