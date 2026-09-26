//go:build !cgo
// +build !cgo

package mpv

import "testing"

func TestNoCGOInvalidClientHandles(t *testing.T) {
	closed := newInitializedMpv(t)
	closed.Destroy()
	for _, m := range []*Mpv{{}, closed} {
		if err := m.Initialize(); err != ERROR_INVALID_PARAMETER {
			t.Fatalf("Initialize = %v, want ERROR_INVALID_PARAMETER", err)
		}
		if err := m.Command([]string{"playlist-clear"}); err != ERROR_INVALID_PARAMETER {
			t.Fatalf("Command = %v, want ERROR_INVALID_PARAMETER", err)
		}
		if err := m.SetProperty("pause", FORMAT_FLAG, true); err != ERROR_INVALID_PARAMETER {
			t.Fatalf("SetProperty = %v, want ERROR_INVALID_PARAMETER", err)
		}
		if _, err := m.GetProperty("pause", FORMAT_FLAG); err != ERROR_INVALID_PARAMETER {
			t.Fatalf("GetProperty = %v, want ERROR_INVALID_PARAMETER", err)
		}
		if ev := m.WaitEvent(0); ev != nil {
			t.Fatalf("WaitEvent = %+v, want nil for an invalid handle", ev)
		}
	}
}
