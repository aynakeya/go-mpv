# Fyne video render demo

This example renders libmpv video frames into Fyne's GLFW OpenGL framebuffer using the go-mpv render API.

It is a separate Go module so Fyne and GLFW dependencies do not affect the root module.
This module and the root library require Go 1.25 or later.

## Run

```bash
cd example/fyne_video_demo
CGO_ENABLED=1 go run . -file ../../data/test.mp4
```

You can also pass a file or URL as the first positional argument.

## Notes

- The example is limited to `linux && cgo` because it reflects Fyne's current GLFW window internals and renders directly into that OpenGL context.
- Fyne v2.8.1 uses GLFW 3.4. Keep this example's GLFW module aligned with Fyne: its reflected viewport must have the same Go type, not just a similar memory layout.
- All player access runs on the UI thread. On close, the demo cancels and joins its worker, then frees the render context before Fyne destroys the GLFW window. Do not move that cleanup after `a.Run()`.
- Direct rendering into Fyne's internal framebuffer remains experimental; Fyne can repaint over the video. This is not a supported Fyne video widget API.
- libmpv is accessed through `github.com/aynakeya/go-mpv`; the demo does not call `mpv_render_*` C functions directly.
