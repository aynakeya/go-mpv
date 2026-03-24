package mpv

import "testing"

func TestEventName(t *testing.T) {
	name := EventName(EVENT_SEEK)
	if name != "seek" {
		t.Fatalf("unexpected event name: %q", name)
	}
}
