//go:build cgo
// +build cgo

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/aynakeya/go-mpv"
)

const (
	renderWidth  = int32(80)
	renderHeight = int32(45)
)

func videoPath() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	for _, p := range []string{"data/test.mp4", "../../data/test.mp4"} {
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err == nil {
				return abs
			}
			return p
		}
	}
	return "data/test.mp4"
}

func main() {
	m := mpv.Create()
	if m == nil {
		log.Fatal("mpv.Create returned nil")
	}
	defer m.Destroy()

	must("set terminal", m.SetOptionString("terminal", "no"))
	must("set vo", m.SetOptionString("vo", "libmpv"))
	must("set ao", m.SetOptionString("ao", "null"))
	must("initialize", m.Initialize())

	rc, err := m.CreateRenderContext([]mpv.RenderParam{
		mpv.RenderParamAPIType(mpv.RENDER_API_TYPE_SW),
	})
	if err != nil {
		log.Fatalf("create render context: %v", err)
	}
	defer rc.Free()

	updates := make(chan struct{}, 1)
	rc.SetUpdateCallback(func() {
		select {
		case updates <- struct{}{}:
		default:
		}
	})
	defer rc.SetUpdateCallback(nil)

	must("loadfile", m.Command([]string{"loadfile", videoPath(), "replace"}))

	fmt.Print("\x1b[2J\x1b[?25l")
	defer fmt.Print("\x1b[0m\x1b[?25h\n")

	var frames int
	lastDraw := time.Now()
	for {
		ev := m.WaitEvent(0.01)
		switch ev.EventId {
		case mpv.EVENT_END_FILE, mpv.EVENT_SHUTDOWN:
			return
		}

		if !drainRenderUpdate(updates) {
			continue
		}
		if rc.Update()&uint64(mpv.RENDER_UPDATE_FRAME) == 0 {
			continue
		}

		frame, err := renderSoftwareFrame(rc)
		if err != nil {
			log.Fatalf("render frame: %v", err)
		}
		frames++
		drawTerminalFrame(frame, frames, time.Since(lastDraw))
		lastDraw = time.Now()
		rc.ReportSwap()
	}
}

func must(action string, err error) {
	if err != nil {
		log.Fatalf("%s: %v", action, err)
	}
}

func drainRenderUpdate(updates <-chan struct{}) bool {
	updated := false
	for {
		select {
		case <-updates:
			updated = true
		default:
			return updated
		}
	}
}

func renderSoftwareFrame(rc *mpv.RenderContext) ([]byte, error) {
	size := [2]int32{renderWidth, renderHeight}
	stride := uintptr(size[0] * 4)
	buffer := make([]byte, int(stride)*int(size[1]))
	blockForTargetTime := int32(0)
	err := rc.Render([]mpv.RenderParam{
		mpv.RenderParamSWSize(&size),
		mpv.RenderParamSWFormat("rgb0"),
		mpv.RenderParamSWStride(&stride),
		mpv.RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
		mpv.RenderParamInt(mpv.RENDER_PARAM_BLOCK_FOR_TARGET_TIME, &blockForTargetTime),
	})
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

func drawTerminalFrame(buffer []byte, frames int, elapsed time.Duration) {
	var b strings.Builder
	b.WriteString("\x1b[H")
	for y := int32(0); y < renderHeight; y++ {
		for x := int32(0); x < renderWidth; x++ {
			i := int((y*renderWidth + x) * 4)
			r, g, bl := buffer[i], buffer[i+1], buffer[i+2]
			fmt.Fprintf(&b, "\x1b[48;2;%d;%d;%dm ", r, g, bl)
		}
		b.WriteString("\x1b[0m\n")
	}
	fmt.Fprintf(&b, "frames=%d frame_time=%s\n", frames, elapsed.Truncate(time.Millisecond))
	fmt.Print(b.String())
}
