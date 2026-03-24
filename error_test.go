//go:build cgo
// +build cgo

package mpv

import "testing"

func TestErrorHelpers(t *testing.T) {
	if ErrorString(ERROR_COMMAND) == "" {
		t.Fatal("ErrorString(ERROR_COMMAND) is empty")
	}
	if newError(0) != nil {
		t.Fatal("newError(0) must be nil")
	}
	if newError(-1) != ERROR_EVENT_QUEUE_FULL {
		t.Fatalf("newError(-1) mismatch: %v", newError(-1))
	}
}
