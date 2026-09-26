package mpv

import (
	"bytes"
	"testing"
	"time"
	"unsafe"
)

func TestRenderContextSoftwareLifecycle(t *testing.T) {
	m := newInitializedMpv(t)

	rc, err := m.CreateRenderContext([]RenderParam{
		RenderParamAPIType(RENDER_API_TYPE_SW),
	})
	if err != nil {
		t.Fatalf("CreateRenderContext(sw) failed: %v", err)
	}
	defer rc.Free()

	rc.SetUpdateCallback(func() {})
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

	info := RenderFrameInfo{Flags: ^uint64(0)}
	if err := rc.GetInfo(RenderParamNextFrameInfo(&info)); err != nil {
		t.Fatalf("GetInfo(next-frame-info) failed: %v", err)
	}
	if info.Flags != 0 {
		t.Fatalf("frame flags without a loaded video = %d, want 0", info.Flags)
	}
	ambientLight := int32(100)
	if err := rc.SetParameter(RenderParamInt(RENDER_PARAM_AMBIENT_LIGHT, &ambientLight)); err != nil && err != ERROR_NOT_IMPLEMENTED {
		t.Fatalf("SetParameter(ambient-light) failed: %v", err)
	}
	_ = rc.Update()
	rc.Free()
	rc.Free()
}

func TestRenderContextSoftwareRendersTestVideo(t *testing.T) {
	video := testVideoPath(t)
	m := newInitializedMpv(t)

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
	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
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

func TestOpenGLInitParamsLifecycle(t *testing.T) {
	if params := NewOpenGLInitParams(nil); params != nil {
		t.Fatal("NewOpenGLInitParams(nil) must return nil")
	}
	params := NewOpenGLInitParams(func(name string) unsafe.Pointer { return nil })
	if params == nil {
		t.Fatal("NewOpenGLInitParams returned nil")
	}
	params.Free()
	params.Free()
}

func TestRenderContextSoftwareRejectsInvalidStride(t *testing.T) {
	m := newInitializedMpv(t)

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
	before := append([]byte(nil), buffer...)

	err = rc.Render([]RenderParam{
		RenderParamSWSize(&size),
		RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride),
		RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
	})
	if err == nil {
		t.Fatal("Render with invalid stride succeeded")
	}
	if !bytes.Equal(buffer, before) {
		t.Fatal("Render error overwrote software buffer")
	}
}
