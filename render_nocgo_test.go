//go:build !cgo
// +build !cgo

package mpv

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
	"unsafe"
)

func TestNoCGORenderParamConstructors(t *testing.T) {
	raw := int32(1)
	if p := NewRenderParam(RENDER_PARAM_DEPTH, unsafe.Pointer(&raw)); p.Type != RENDER_PARAM_DEPTH || p.Data != unsafe.Pointer(&raw) {
		t.Fatalf("unexpected raw render param: %+v", p)
	}

	if p := RenderParamAPIType(RENDER_API_TYPE_SW); p.Type != RENDER_PARAM_API_TYPE {
		t.Fatalf("unexpected API type param: %+v", p)
	}

	fbo := &OpenGLFBO{FBO: 1, W: 2, H: 3, InternalFormat: 4}
	if p := RenderParamOpenGLFBO(fbo); p.Type != RENDER_PARAM_OPENGL_FBO || p.Data != unsafe.Pointer(fbo) {
		t.Fatalf("unexpected OpenGL FBO param: %+v", p)
	}

	drm := &OpenGLDRMParams{FD: -1, RenderFD: -1}
	if p := RenderParamDRMDisplay(drm); p.Type != RENDER_PARAM_DRM_DISPLAY || p.Data != unsafe.Pointer(drm) {
		t.Fatalf("unexpected DRM display param: %+v", p)
	}

	size := &OpenGLDRMDrawSurfaceSize{Width: 640, Height: 360}
	if p := RenderParamDRMDrawSurfaceSize(size); p.Type != RENDER_PARAM_DRM_DRAW_SURFACE_SIZE || p.Data != unsafe.Pointer(size) {
		t.Fatalf("unexpected DRM draw surface size param: %+v", p)
	}

	drmV2 := &OpenGLDRMParamsV2{FD: -1, RenderFD: -1}
	if p := RenderParamDRMDisplayV2(drmV2); p.Type != RENDER_PARAM_DRM_DISPLAY_V2 || p.Data != unsafe.Pointer(drmV2) {
		t.Fatalf("unexpected DRM display v2 param: %+v", p)
	}

	swSize := &[2]int32{16, 16}
	if p := RenderParamSWSize(swSize); p.Type != RENDER_PARAM_SW_SIZE || p.Data != unsafe.Pointer(swSize) {
		t.Fatalf("unexpected SW size param: %+v", p)
	}
	if p := RenderParamSWFormat("rgb0"); p.Type != RENDER_PARAM_SW_FORMAT {
		t.Fatalf("unexpected SW format param: %+v", p)
	}
	stride := uintptr(64)
	if p := RenderParamSWStride(&stride); p.Type != RENDER_PARAM_SW_STRIDE || p.Data != unsafe.Pointer(&stride) {
		t.Fatalf("unexpected SW stride param: %+v", p)
	}
	buffer := make([]byte, 16)
	if p := RenderParamSWPointer(unsafe.Pointer(&buffer[0])); p.Type != RENDER_PARAM_SW_POINTER || p.Data != unsafe.Pointer(&buffer[0]) {
		t.Fatalf("unexpected SW pointer param: %+v", p)
	}

	info := &RenderFrameInfo{}
	if p := RenderParamNextFrameInfo(info); p.Type != RENDER_PARAM_NEXT_FRAME_INFO || p.Data != unsafe.Pointer(info) {
		t.Fatalf("unexpected next frame info param: %+v", p)
	}
}

func noCGOTestVideoPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	p := filepath.Join(wd, "data", "test.mp4")
	if _, err := os.Stat(p); err != nil {
		t.Skipf("test video not found at %s: %v", p, err)
	}
	return p
}

func TestNoCGORenderUnavailableInputs(t *testing.T) {
	var m Mpv
	if rc, err := m.CreateRenderContext([]RenderParam{RenderParamAPIType(RENDER_API_TYPE_SW)}); !errors.Is(err, ErrCGODisabled) || rc != nil {
		t.Fatalf("CreateRenderContext = (%v, %v), want nil ErrCGODisabled", rc, err)
	}

	rc := &RenderContext{}
	if err := rc.SetParameter(RenderParamInt(RENDER_PARAM_DEPTH, new(int32))); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("SetParameter err = %v, want ErrCGODisabled", err)
	}
	if err := rc.GetInfo(RenderParamNextFrameInfo(&RenderFrameInfo{})); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("GetInfo err = %v, want ErrCGODisabled", err)
	}
	if err := rc.Render(nil); !errors.Is(err, ErrCGODisabled) {
		t.Fatalf("Render err = %v, want ErrCGODisabled", err)
	}
	if flags := rc.Update(); flags != 0 {
		t.Fatalf("Update flags = %d, want 0", flags)
	}
	rc.SetUpdateCallback(func() {
		t.Fatal("nocgo render update callback should not be invoked")
	})
	rc.ReportSwap()
	rc.Free()

	if params := NewOpenGLInitParams(nil); params != nil {
		t.Fatalf("NewOpenGLInitParams(nil) = %v, want nil", params)
	}
}

func TestNoCGORenderContextSoftwareLifecycle(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()

	if err := m.SetOptionString("terminal", "no"); err != nil {
		t.Fatalf("SetOptionString(terminal) failed: %v", err)
	}
	if err := m.SetOptionString("vo", "libmpv"); err != nil {
		t.Fatalf("SetOptionString(vo) failed: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	rc, err := m.CreateRenderContext([]RenderParam{
		RenderParamAPIType(RENDER_API_TYPE_SW),
	})
	if err != nil {
		t.Fatalf("CreateRenderContext(sw) failed: %v", err)
	}
	defer rc.Free()

	var callbackCount int
	rc.SetUpdateCallback(func() {
		callbackCount++
	})
	rc.SetUpdateCallback(nil)

	size := [2]int32{16, 16}
	stride := uintptr(size[0] * 4)
	buffer := make([]byte, int(stride)*int(size[1]))
	renderParams := []RenderParam{
		RenderParamSWSize(&size),
		RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride),
		RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
	}
	if err := rc.Render(renderParams); err != nil {
		t.Fatalf("Render(sw) failed: %v", err)
	}
	if err := rc.Render(renderParams); err != nil {
		t.Fatalf("Render(sw) with reused params failed: %v", err)
	}
	rc.ReportSwap()

	var info RenderFrameInfo
	if err := rc.GetInfo(RenderParamNextFrameInfo(&info)); err != nil {
		t.Fatalf("GetInfo(next-frame-info) failed: %v", err)
	}
	ambientLight := int32(100)
	if err := rc.SetParameter(RenderParamInt(RENDER_PARAM_AMBIENT_LIGHT, &ambientLight)); err != nil && err != ERROR_NOT_IMPLEMENTED {
		t.Fatalf("SetParameter(ambient-light) failed: %v", err)
	}
	_ = rc.Update()
	rc.Free()
	rc.Free()
	_ = callbackCount

	params := NewOpenGLInitParams(func(name string) unsafe.Pointer { return nil })
	if params == nil {
		t.Fatal("NewOpenGLInitParams returned nil for non-nil callback")
	}
	params.Free()
	params.Free()
}

func TestNoCGORenderContextSoftwareRenderRejectsInvalidStride(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()

	if err := m.SetOptionString("terminal", "no"); err != nil {
		t.Fatalf("SetOptionString(terminal) failed: %v", err)
	}
	if err := m.SetOptionString("vo", "libmpv"); err != nil {
		t.Fatalf("SetOptionString(vo) failed: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	rc, err := m.CreateRenderContext([]RenderParam{
		RenderParamAPIType(RENDER_API_TYPE_SW),
	})
	if err != nil {
		t.Fatalf("CreateRenderContext(sw) failed: %v", err)
	}
	defer rc.Free()

	size := [2]int32{16, 16}
	stride := uintptr(1)
	buffer := bytes.Repeat([]byte{0xa5}, int(size[0])*int(size[1])*4)

	err = rc.Render([]RenderParam{
		RenderParamSWSize(&size),
		RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride),
		RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
	})
	if err == nil {
		t.Fatal("Render with invalid stride succeeded")
	}
}

func TestNoCGORenderContextSoftwareRendersTestVideo(t *testing.T) {
	video := noCGOTestVideoPath(t)
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()

	if err := m.SetOptionString("terminal", "no"); err != nil {
		t.Fatalf("SetOptionString(terminal) failed: %v", err)
	}
	if err := m.SetOptionString("vo", "libmpv"); err != nil {
		t.Fatalf("SetOptionString(vo) failed: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("SetOptionString(ao) failed: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	rc, err := m.CreateRenderContext([]RenderParam{
		RenderParamAPIType(RENDER_API_TYPE_SW),
	})
	if err != nil {
		t.Fatalf("CreateRenderContext(sw) failed: %v", err)
	}
	defer rc.Free()

	updates := make(chan struct{}, 8)
	rc.SetUpdateCallback(func() {
		select {
		case updates <- struct{}{}:
		default:
		}
	})
	defer rc.SetUpdateCallback(nil)

	if err := m.Command([]string{"loadfile", video, "replace"}); err != nil {
		t.Fatalf("Command(loadfile) failed: %v", err)
	}
	waitEventNoCGO(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_FILE_LOADED
	})

	size := [2]int32{96, 54}
	stride := uintptr(size[0] * 4)
	buffer := make([]byte, int(stride)*int(size[1]))
	blockForTargetTime := int32(0)
	renderParams := []RenderParam{
		RenderParamSWSize(&size),
		RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride),
		RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
		RenderParamInt(RENDER_PARAM_BLOCK_FOR_TARGET_TIME, &blockForTargetTime),
	}

	fill := bytes.Repeat([]byte{0xa5}, len(buffer))
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-updates:
		default:
			m.WaitEvent(0.05)
		}

		if rc.Update()&uint64(RENDER_UPDATE_FRAME) == 0 {
			continue
		}

		copy(buffer, fill)
		if err := rc.Render(renderParams); err != nil {
			t.Fatalf("Render(test video) failed: %v", err)
		}
		rc.ReportSwap()
		if !bytes.Equal(buffer, fill) {
			return
		}
	}
	t.Fatal("rendered test video did not write to software buffer")
}
