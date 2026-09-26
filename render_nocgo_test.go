//go:build !cgo
// +build !cgo

package mpv

import (
	"errors"
	"testing"
)

func TestNoCGORenderInvalidHandles(t *testing.T) {
	var m Mpv
	if rc, err := m.CreateRenderContext([]RenderParam{RenderParamAPIType(RENDER_API_TYPE_SW)}); !errors.Is(err, ERROR_INVALID_PARAMETER) || rc != nil {
		t.Fatalf("CreateRenderContext = (%v, %v), want nil ERROR_INVALID_PARAMETER", rc, err)
	}

	rc := &RenderContext{}
	if err := rc.SetParameter(RenderParamInt(RENDER_PARAM_DEPTH, new(int32))); !errors.Is(err, ERROR_INVALID_PARAMETER) {
		t.Fatalf("SetParameter err = %v, want ERROR_INVALID_PARAMETER", err)
	}
	if err := rc.GetInfo(RenderParamNextFrameInfo(&RenderFrameInfo{})); !errors.Is(err, ERROR_INVALID_PARAMETER) {
		t.Fatalf("GetInfo err = %v, want ERROR_INVALID_PARAMETER", err)
	}
	if err := rc.Render(nil); !errors.Is(err, ERROR_INVALID_PARAMETER) {
		t.Fatalf("Render err = %v, want ERROR_INVALID_PARAMETER", err)
	}
	if flags := rc.Update(); flags != 0 {
		t.Fatalf("Update flags = %d, want 0", flags)
	}
	rc.SetUpdateCallback(func() {
		t.Fatal("nocgo render update callback should not be invoked")
	})
	rc.ReportSwap()
	rc.Free()
}
