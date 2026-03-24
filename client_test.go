//go:build cgo
// +build cgo

package mpv

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newInitializedMpv(t *testing.T) *Mpv {
	t.Helper()
	m := Create()
	if m == nil {
		t.Fatal("Create returned nil")
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Cleanup(m.Destroy)
	return m
}

func waitForEvent(t *testing.T, m *Mpv, timeout time.Duration, match func(*Event) bool) *Event {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ev := m.WaitEvent(0.1)
		if match(ev) {
			return ev
		}
	}
	t.Fatalf("waitForEvent timeout after %s", timeout)
	return nil
}

func TestClientApiVersion(t *testing.T) {
	if v := ClientApiVersion(); v == 0 {
		t.Fatal("ClientApiVersion returned 0")
	}
}

func TestMpvBasicLifecycleAndInfo(t *testing.T) {
	m := newInitializedMpv(t)
	if m.ClientName() == "" {
		t.Fatal("ClientName is empty")
	}
	if m.ClientId() <= 0 {
		t.Fatalf("ClientId invalid: %d", m.ClientId())
	}
	if m.GetTimeNS() <= 0 {
		t.Fatalf("GetTimeNS invalid: %d", m.GetTimeNS())
	}
	if m.GetTimeUS() <= 0 {
		t.Fatalf("GetTimeUS invalid: %d", m.GetTimeUS())
	}
}

func TestMpvPropertySync(t *testing.T) {
	m := newInitializedMpv(t)
	v, err := m.GetProperty("idle-active", FORMAT_FLAG)
	if err != nil {
		t.Fatalf("GetProperty(idle-active) failed: %v", err)
	}
	if _, ok := v.(bool); !ok {
		t.Fatalf("unexpected type for idle-active: %T", v)
	}
}

func TestMpvPropertyAsync(t *testing.T) {
	m := newInitializedMpv(t)
	const ud = uint64(1001)
	if err := m.GetPropertyAsync("idle-active", ud, FORMAT_FLAG); err != nil {
		t.Fatalf("GetPropertyAsync failed: %v", err)
	}
	ev := waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_GET_PROPERTY_REPLY && e.ReplyUserData == ud
	})
	if ev.Error != nil {
		t.Fatalf("async get property event error: %v", ev.Error)
	}
	p := ev.Property()
	if p.Format != FORMAT_FLAG {
		t.Fatalf("unexpected property format: %v", p.Format)
	}
	if _, ok := p.Data.(bool); !ok {
		t.Fatalf("unexpected property data type: %T", p.Data)
	}
	if _, err := ev.ToNode(); err != nil {
		t.Fatalf("Event.ToNode failed: %v", err)
	}
}

func TestMpvCommandVariants(t *testing.T) {
	m := newInitializedMpv(t)
	if err := m.Command([]string{"playlist-clear"}); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	ret, err := m.CommandRet([]string{"expand-text", "${idle-active}"})
	if err != nil {
		t.Fatalf("CommandRet failed: %v", err)
	}
	if ret.Format == FORMAT_NONE {
		t.Fatal("CommandRet returned FORMAT_NONE")
	}

	nodeCmd := Node{
		Format: FORMAT_NODE_ARRAY,
		Value: []Node{
			{Value: "expand-text", Format: FORMAT_STRING},
			{Value: "${idle-active}", Format: FORMAT_STRING},
		},
	}
	nodeRet, err := m.CommandNode(nodeCmd)
	if err != nil {
		t.Fatalf("CommandNode failed: %v", err)
	}
	if nodeRet.Format == FORMAT_NONE {
		t.Fatal("CommandNode returned FORMAT_NONE")
	}
}

func TestMpvCommandAsyncVariants(t *testing.T) {
	m := newInitializedMpv(t)
	const cmdUD = uint64(2001)
	if err := m.CommandAsync(cmdUD, []string{"expand-text", "${idle-active}"}); err != nil {
		t.Fatalf("CommandAsync failed: %v", err)
	}
	ev := waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_COMMAND_REPLY && e.ReplyUserData == cmdUD
	})
	if ev.Error != nil {
		t.Fatalf("command async event error: %v", ev.Error)
	}
	if ev.Command().Result.Format == FORMAT_NONE {
		t.Fatal("CommandAsync returned empty result")
	}

	const nodeUD = uint64(2002)
	nodeCmd := Node{
		Format: FORMAT_NODE_ARRAY,
		Value: []Node{
			{Value: "expand-text", Format: FORMAT_STRING},
			{Value: "${idle-active}", Format: FORMAT_STRING},
		},
	}
	if err := m.CommandNodeAsync(nodeUD, nodeCmd); err != nil {
		t.Fatalf("CommandNodeAsync failed: %v", err)
	}
	ev = waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_COMMAND_REPLY && e.ReplyUserData == nodeUD
	})
	if ev.Error != nil {
		t.Fatalf("command node async event error: %v", ev.Error)
	}
	if ev.Command().Result.Format == FORMAT_NONE {
		t.Fatal("CommandNodeAsync returned empty result")
	}
}

func TestMpvHookAdd(t *testing.T) {
	m := newInitializedMpv(t)
	if err := m.HookAdd(3001, "on_load", 0); err != nil {
		t.Fatalf("HookAdd failed: %v", err)
	}
}

func TestMpvCreateClientVariants(t *testing.T) {
	m := newInitializedMpv(t)

	c1 := m.CreateClient("client_test_child")
	if c1 == nil {
		t.Fatal("CreateClient returned nil")
	}
	t.Cleanup(c1.Destroy)

	c2 := m.CreateWeakClient("client_test_weak")
	if c2 == nil {
		t.Fatal("CreateWeakClient returned nil")
	}
	t.Cleanup(c2.Destroy)

	if c1.ClientName() == "" || c2.ClientName() == "" {
		t.Fatal("child client name is empty")
	}
	if c1.ClientId() <= 0 || c2.ClientId() <= 0 {
		t.Fatalf("invalid child ids: %d, %d", c1.ClientId(), c2.ClientId())
	}
}

func TestMpvCommandEmptyInput(t *testing.T) {
	m := newInitializedMpv(t)
	if err := m.Command([]string{}); err == nil {
		t.Fatal("Command(empty) should return error")
	}
	if err := m.CommandAsync(4901, []string{}); err == nil {
		t.Fatal("CommandAsync(empty) should return error")
	}
	if _, err := m.CommandRet([]string{}); err == nil {
		t.Fatal("CommandRet(empty) should return error")
	}
}

func TestMpvObserveUnobserve(t *testing.T) {
	m := newInitializedMpv(t)
	const ud = uint64(5001)
	if err := m.ObserveProperty(ud, "pause", FORMAT_FLAG); err != nil {
		t.Fatalf("ObserveProperty failed: %v", err)
	}

	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_PROPERTY_CHANGE && e.ReplyUserData == ud
	})

	if err := m.UnObserveProperty(ud); err != nil {
		t.Fatalf("UnObserveProperty failed: %v", err)
	}

	if err := m.SetProperty("pause", FORMAT_FLAG, true); err != nil {
		t.Fatalf("SetProperty(pause=true) failed: %v", err)
	}
	if err := m.SetProperty("pause", FORMAT_FLAG, false); err != nil {
		t.Fatalf("SetProperty(pause=false) failed: %v", err)
	}

	deadline := time.Now().Add(700 * time.Millisecond)
	for time.Now().Before(deadline) {
		ev := m.WaitEvent(0.05)
		if ev.EventId == EVENT_PROPERTY_CHANGE && ev.ReplyUserData == ud {
			t.Fatalf("received property event after unobserve: %+v", ev)
		}
	}
}

func testVideoPath(t *testing.T) string {
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

func TestMpvLoadFileAndPlaybackEvents(t *testing.T) {
	video := testVideoPath(t)
	m := Create()
	if m == nil {
		t.Fatal("Create returned nil")
	}
	if err := m.SetOptionString("terminal", "no"); err != nil {
		t.Fatalf("SetOptionString(terminal) failed: %v", err)
	}
	if err := m.SetOptionString("vo", "null"); err != nil {
		t.Fatalf("SetOptionString(vo) failed: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("SetOptionString(ao) failed: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Cleanup(m.Destroy)

	if err := m.ObserveProperty(4001, "time-pos", FORMAT_DOUBLE); err != nil {
		t.Fatalf("ObserveProperty(time-pos) failed: %v", err)
	}

	if err := m.Command([]string{"loadfile", video, "replace"}); err != nil {
		t.Fatalf("Command(loadfile) failed: %v", err)
	}

	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_START_FILE
	})
	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_FILE_LOADED
	})
	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		if e.EventId != EVENT_PROPERTY_CHANGE || e.ReplyUserData != 4001 {
			return false
		}
		p := e.Property()
		if p.Format != FORMAT_DOUBLE {
			return false
		}
		_, ok := p.Data.(float64)
		return ok
	})

	if err := m.Command([]string{"stop"}); err != nil {
		t.Fatalf("Command(stop) failed: %v", err)
	}
	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_END_FILE
	})
}

func TestMpvHookContinueOnLoad(t *testing.T) {
	video := testVideoPath(t)
	m := Create()
	if m == nil {
		t.Fatal("Create returned nil")
	}
	if err := m.SetOptionString("terminal", "no"); err != nil {
		t.Fatalf("SetOptionString(terminal) failed: %v", err)
	}
	if err := m.SetOptionString("vo", "null"); err != nil {
		t.Fatalf("SetOptionString(vo) failed: %v", err)
	}
	if err := m.SetOptionString("ao", "null"); err != nil {
		t.Fatalf("SetOptionString(ao) failed: %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	t.Cleanup(m.Destroy)

	const ud = uint64(6001)
	if err := m.HookAdd(ud, "on_load", 0); err != nil {
		t.Fatalf("HookAdd(on_load) failed: %v", err)
	}

	if err := m.Command([]string{"loadfile", video, "replace"}); err != nil {
		t.Fatalf("Command(loadfile) failed: %v", err)
	}

	ev := waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_HOOK && e.ReplyUserData == ud
	})
	h := ev.Hook()
	if h.ID == 0 {
		t.Fatalf("invalid hook id: %+v", h)
	}
	if err := m.HookContinue(h.ID); err != nil {
		t.Fatalf("HookContinue failed: %v", err)
	}

	waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_FILE_LOADED
	})
}
