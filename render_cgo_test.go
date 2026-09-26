//go:build cgo
// +build cgo

package mpv

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestRenderContextSoftwareRejectsTargetSizeOverflow(t *testing.T) {
	m := newInitializedMpv(t)

	rc, err := m.CreateRenderContext([]RenderParam{
		RenderParamAPIType(RENDER_API_TYPE_SW),
	})
	if err != nil {
		t.Fatalf("CreateRenderContext(sw) failed: %v", err)
	}
	defer rc.Free()

	size := [2]int32{16, 16}
	stride := ^uintptr(0)/uintptr(size[1]) + 5
	buffer := bytes.Repeat([]byte{0xa5}, int(size[0])*int(size[1])*4)
	before := append([]byte(nil), buffer...)

	// Reject overflow before allocating or copying the temporary C target.
	err = rc.Render([]RenderParam{
		RenderParamSWSize(&size),
		RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride),
		RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
	})
	if err == nil {
		t.Fatalf("Render with invalid stride %d succeeded", stride)
	}
	if !bytes.Equal(buffer, before) {
		t.Fatal("Render error overwrote software buffer")
	}
}
