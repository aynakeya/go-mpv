//go:build !cgo
// +build !cgo

package mpv

import (
	"runtime"
	"testing"
	"time"
)

func TestNoCGOBackendBasicClientPath(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()

	if ClientApiVersion() == 0 {
		t.Fatal("ClientApiVersion returned 0")
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	if m.ClientName() == "" {
		t.Fatal("ClientName is empty")
	}
	if m.ClientId() <= 0 {
		t.Fatalf("ClientId invalid: %d", m.ClientId())
	}
}

func TestNoCGOBackendBasicCommandAndProperty(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if err := m.Command([]string{"playlist-clear"}); err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	v, err := m.GetProperty("idle-active", FORMAT_FLAG)
	if err != nil {
		t.Fatalf("GetProperty failed: %v", err)
	}
	if _, ok := v.(bool); !ok {
		t.Fatalf("unexpected type: %T", v)
	}
}

func waitEventNoCGO(t *testing.T, m *Mpv, timeout time.Duration, match func(*Event) bool) *Event {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ev := m.WaitEvent(0.1)
		if match(ev) {
			return ev
		}
	}
	t.Fatalf("waitEvent timeout after %s", timeout)
	return nil
}

func TestNoCGOBackendCommandRetAndNode(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	ret, err := m.CommandRet([]string{"expand-text", "${idle-active}"})
	if err != nil {
		t.Fatalf("CommandRet failed: %v", err)
	}
	if ret.Format != FORMAT_STRING {
		t.Fatalf("unexpected CommandRet format: %v", ret.Format)
	}

	nodeCmd := Node{
		Format: FORMAT_NODE_ARRAY,
		Value: []Node{
			{Format: FORMAT_STRING, Value: "expand-text"},
			{Format: FORMAT_STRING, Value: "${idle-active}"},
		},
	}
	nodeRet, err := m.CommandNode(nodeCmd)
	if err != nil {
		t.Fatalf("CommandNode failed: %v", err)
	}
	if nodeRet.Format != FORMAT_STRING {
		t.Fatalf("unexpected CommandNode format: %v", nodeRet.Format)
	}
}

func TestNoCGOBackendAsyncPropertyEvent(t *testing.T) {
	m := Create()
	if m == nil {
		t.Skip("libmpv not available for nocgo backend")
	}
	defer m.Destroy()
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	const ud = uint64(7001)
	if err := m.GetPropertyAsync("idle-active", ud, FORMAT_FLAG); err != nil {
		t.Fatalf("GetPropertyAsync failed: %v", err)
	}
	ev := waitEventNoCGO(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_GET_PROPERTY_REPLY && e.ReplyUserData == ud
	})
	if ev.Error != nil {
		t.Fatalf("event error: %v", ev.Error)
	}
	p := ev.Property()
	if p.Format != FORMAT_FLAG {
		t.Fatalf("unexpected event property format: %v", p.Format)
	}
	if _, ok := p.Data.(bool); !ok {
		t.Fatalf("unexpected event property type: %T", p.Data)
	}
}

func TestNoCGOLibCandidatesByOS(t *testing.T) {
	cands := libMPVCandidates(runtime.GOOS)
	if len(cands) == 0 {
		t.Fatal("libMPVCandidates returned empty slice")
	}
	switch runtime.GOOS {
	case "windows":
		if cands[0] != "mpv-2.dll" {
			t.Fatalf("unexpected windows first candidate: %s", cands[0])
		}
	case "darwin":
		if cands[0] != "libmpv.2.dylib" {
			t.Fatalf("unexpected darwin first candidate: %s", cands[0])
		}
	default:
		if cands[0] != "libmpv.so.2" {
			t.Fatalf("unexpected unix first candidate: %s", cands[0])
		}
	}
}
