//go:build linux && cgo
// +build linux,cgo

package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"github.com/aynakeya/go-mpv"
	goglfw "github.com/go-gl/glfw/v3.4/glfw"
)

const defaultFPS = 60

func main() {
	var media string
	var fps int
	flag.StringVar(&media, "file", defaultMedia(), "video file path or URL to render")
	flag.IntVar(&fps, "fps", defaultFPS, "render loop frames per second")
	flag.Parse()
	if flag.NArg() > 0 {
		media = flag.Arg(0)
	}
	if fps <= 0 {
		fail("fps must be positive")
	}

	a := app.New()
	w := a.NewWindow("Fyne + go-mpv render API demo")
	w.Resize(fyne.NewSize(960, 540))
	// Fyne needs content to create and maintain the underlying GLFW viewport.
	// The video frame is rendered directly into that viewport after Fyne draws.
	w.SetContent(canvas.NewRectangle(color.Black))

	done := make(chan struct{})
	finished := make(chan struct{})
	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() { close(done) })
		// The worker waits cancellably for UI work, so joining it here does
		// not need the event loop and prevents enqueueing after Fyne stops.
		<-finished
	}
	w.SetOnClosed(stop)

	w.Show()
	go func() {
		defer close(finished)
		runDemo(w, media, fps, done, stop)
	}()
	a.Run()
	stop()
}

func defaultMedia() string {
	for _, p := range []string{"../../data/test.mp4", "data/test.mp4"} {
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err == nil {
				return abs
			}
			return p
		}
	}
	return "../../data/test.mp4"
}

func runDemo(w fyne.Window, media string, fps int, done <-chan struct{}, stop func()) {
	viewport, err := waitFyneGLFWWindow(w, done)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fyne.Do(func() { w.Close() })
		return
	}

	var player *glMPV
	if !doWhileOpen(done, func() {
		viewport.MakeContextCurrent()
		defer goglfw.DetachCurrentContext()
		player, err = newGLMPV(media)
		if err == nil {
			// Fyne invokes OnClosed before destroying its GLFW window. Free
			// libmpv here while its GL context and the UI thread still exist.
			w.SetOnClosed(func() {
				stop()
				previous := goglfw.GetCurrentContext()
				viewport.MakeContextCurrent()
				player.Close()
				if previous != nil {
					previous.MakeContextCurrent()
				} else {
					goglfw.DetachCurrentContext()
				}
			})
		}
	}) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fyne.Do(func() { w.Close() })
		return
	}

	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if !doWhileOpen(done, func() {
				// Event reads, rendering and cleanup all use the UI thread.
				viewport.MakeContextCurrent()
				defer goglfw.DetachCurrentContext()
				player.drainEvents()
				width, height := viewport.GetFramebufferSize()
				if width > 0 && height > 0 {
					if err := player.render(width, height); err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
					viewport.SwapBuffers()
					player.reportSwap()
				}
			}) {
				return
			}
		}
	}
}

// doWhileOpen avoids waiting for a callback after Fyne has stopped servicing
// its queue. Queued callbacks also check cancellation before touching GL.
// A true result means fn ran to completion, not just that it left the queue.
func doWhileOpen(done <-chan struct{}, fn func()) bool {
	select {
	case <-done:
		return false
	default:
	}
	completed := make(chan struct{})
	fyne.Do(func() {
		select {
		case <-done:
			return
		default:
			fn()
			close(completed)
		}
	})
	select {
	case <-done:
		return false
	case <-completed:
		return true
	}
}

func waitFyneGLFWWindow(w fyne.Window, done <-chan struct{}) (*goglfw.Window, error) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(5 * time.Second)
	for {
		var viewport *goglfw.Window
		var err error
		if !doWhileOpen(done, func() {
			viewport, err = reflectFyneGLFWWindow(w)
		}) {
			return nil, fmt.Errorf("window closed before Fyne GLFW viewport was available")
		}
		if err == nil && viewport != nil {
			return viewport, nil
		}

		select {
		case <-done:
			return nil, fmt.Errorf("window closed before Fyne GLFW viewport was available")
		case <-deadline:
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timed out waiting for Fyne GLFW viewport")
		case <-ticker.C:
		}
	}
}

func reflectFyneGLFWWindow(w fyne.Window) (*goglfw.Window, error) {
	v := reflect.ValueOf(w)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, fmt.Errorf("unexpected Fyne window type %T", w)
	}
	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("unexpected Fyne window element type %s", elem.Kind())
	}
	field := elem.FieldByName("viewport")
	if !field.IsValid() {
		return nil, fmt.Errorf("Fyne window type %T has no viewport field", w)
	}
	if field.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("Fyne viewport field has unexpected kind %s", field.Kind())
	}
	if field.IsNil() {
		return nil, nil
	}
	return (*goglfw.Window)(unsafe.Pointer(field.Pointer())), nil
}

type glMPV struct {
	mpv    *mpv.Mpv
	ctx    *mpv.RenderContext
	glInit *mpv.OpenGLInitParams
	closed bool // All player operations run on the UI thread.
}

func newGLMPV(media string) (*glMPV, error) {
	m := mpv.Create()
	if m == nil {
		return nil, fmt.Errorf("mpv.Create returned nil")
	}
	p := &glMPV{mpv: m}

	for name, value := range map[string]string{
		"config":   "no",
		"terminal": "yes",
		"vo":       "libmpv",
		"hwdec":    "no",
	} {
		if err := m.SetOptionString(name, value); err != nil {
			p.Close()
			return nil, fmt.Errorf("set %s: %w", name, err)
		}
	}
	if err := m.RequestLogMessages(mpv.LOG_LEVEL_WARN); err != nil {
		p.Close()
		return nil, fmt.Errorf("request mpv logs: %w", err)
	}
	if err := m.Initialize(); err != nil {
		p.Close()
		return nil, fmt.Errorf("initialize mpv: %w", err)
	}

	p.glInit = mpv.NewOpenGLInitParams(goglfw.GetProcAddress)
	if p.glInit == nil {
		p.Close()
		return nil, fmt.Errorf("create OpenGL init params failed")
	}
	ctx, err := m.CreateRenderContext([]mpv.RenderParam{
		mpv.RenderParamAPIType(mpv.RENDER_API_TYPE_OPENGL),
		mpv.RenderParamOpenGLInitParams(p.glInit),
	})
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("create mpv OpenGL render context: %w", err)
	}
	p.ctx = ctx

	if err := m.Command([]string{"loadfile", media, "replace"}); err != nil {
		p.Close()
		return nil, fmt.Errorf("load media: %w", err)
	}
	return p, nil
}

func (p *glMPV) Close() {
	if p == nil || p.closed {
		return
	}
	p.closed = true
	if p.ctx != nil {
		p.ctx.Free()
		p.ctx = nil
	}
	if p.glInit != nil {
		p.glInit.Free()
		p.glInit = nil
	}
	if p.mpv != nil {
		p.mpv.TerminateDestroy()
		p.mpv = nil
	}
}

func (p *glMPV) render(width int, height int) error {
	if p.closed {
		return fmt.Errorf("mpv is closed")
	}
	fbo := &mpv.OpenGLFBO{FBO: 0, W: int32(width), H: int32(height)}
	flipY := int32(1)
	return p.ctx.Render([]mpv.RenderParam{
		mpv.RenderParamOpenGLFBO(fbo),
		mpv.RenderParamInt(mpv.RENDER_PARAM_FLIP_Y, &flipY),
	})
}

func (p *glMPV) reportSwap() {
	if p != nil && p.ctx != nil && !p.closed {
		p.ctx.ReportSwap()
	}
}

func (p *glMPV) drainEvents() {
	for {
		event := p.mpv.WaitEvent(0)
		if event == nil || event.EventId == mpv.EVENT_NONE {
			return
		}
		switch event.EventId {
		case mpv.EVENT_END_FILE:
			fmt.Fprintln(os.Stderr, "mpv: end-file")
		case mpv.EVENT_SHUTDOWN:
			fmt.Fprintln(os.Stderr, "mpv: shutdown")
			return
		case mpv.EVENT_LOG_MESSAGE:
			msg := event.LogMessage()
			fmt.Fprintf(os.Stderr, "mpv[%s] %s", msg.Level, msg.Text)
		}
	}
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
