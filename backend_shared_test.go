package mpv

import (
	"bytes"
	"reflect"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

// Both backends must exercise these contracts against an installed libmpv.
func newHeadlessClient(t *testing.T) *Mpv {
	t.Helper()
	m := Create()
	if m == nil {
		t.Fatal("Create returned nil; these tests require libmpv")
	}
	t.Cleanup(m.Destroy)
	for name, value := range map[string]string{
		"config": "no", "terminal": "no", "vo": "libmpv", "ao": "null",
	} {
		if err := m.SetOptionString(name, value); err != nil {
			t.Fatalf("SetOptionString(%s): %v", name, err)
		}
	}
	return m
}

func TestMpvRepeatedDestruction(t *testing.T) {
	for _, initialized := range []bool{false, true} {
		for _, terminate := range []bool{false, true} {
			name := "uninitialized"
			if initialized {
				name = "initialized"
			}
			if terminate {
				name += "/terminate"
			} else {
				name += "/destroy"
			}
			t.Run(name, func(t *testing.T) {
				m := newHeadlessClient(t)
				if initialized {
					if err := m.Initialize(); err != nil {
						t.Fatal(err)
					}
				}
				if terminate {
					m.TerminateDestroy()
				} else {
					m.Destroy()
				}
				m.Destroy()
				m.TerminateDestroy()
			})
		}
	}
}

func TestMpvNodeWrites(t *testing.T) {
	m := newHeadlessClient(t)
	if err := m.SetOption("pause", FORMAT_NODE, Node{Format: FORMAT_FLAG, Value: true}); err != nil {
		t.Fatalf("SetOption(node): %v", err)
	}
	if err := m.Initialize(); err != nil {
		t.Fatal(err)
	}
	if got, err := m.GetProperty("pause", FORMAT_FLAG); err != nil || got != true {
		t.Fatalf("pause = %v, %v; want true", got, err)
	}

	want := Node{Format: FORMAT_NODE_MAP, Value: map[string]Node{
		"title": {Format: FORMAT_STRING, Value: "go-mpv"},
		"items": {Format: FORMAT_NODE_ARRAY, Value: []Node{
			{Format: FORMAT_INT64, Value: int64(42)},
			{Format: FORMAT_FLAG, Value: true},
			{Format: FORMAT_DOUBLE, Value: 1.5},
		}},
	}}
	const property = "user-data/go-mpv-test"
	if err := m.SetProperty(property, FORMAT_NODE, want); err != nil {
		t.Fatalf("SetProperty(node): %v", err)
	}
	if got, err := m.GetProperty(property, FORMAT_NODE); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProperty = %#v, %v; want %#v", got, err, want)
	}

	want.Value.(map[string]Node)["title"] = Node{Format: FORMAT_STRING, Value: "async"}
	const userdata = 1234
	if err := m.SetPropertyAsync(property, userdata, FORMAT_NODE, want); err != nil {
		t.Fatalf("SetPropertyAsync(node): %v", err)
	}
	// The async API must own its copy after returning, not depend on the
	// temporary marshaling arena surviving until the reply is consumed.
	runtime.GC()
	ev := waitForEvent(t, m, 5*time.Second, func(e *Event) bool {
		return e.EventId == EVENT_SET_PROPERTY_REPLY && e.ReplyUserData == userdata
	})
	if ev.Error != nil {
		t.Fatalf("async node write: %v", ev.Error)
	}
	if got, err := m.GetProperty(property, FORMAT_NODE); err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("GetProperty after async write = %#v, %v; want %#v", got, err, want)
	}
}

func TestRenderSoftwareSkipPreservesTarget(t *testing.T) {
	m := newInitializedMpv(t)
	rc, err := m.CreateRenderContext([]RenderParam{RenderParamAPIType(RENDER_API_TYPE_SW)})
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Free()
	size := [2]int32{16, 16}
	stride := uintptr(size[0] * 4)
	buffer := bytes.Repeat([]byte{0xa5}, int(stride)*int(size[1]))
	before := append([]byte(nil), buffer...)
	target := []RenderParam{
		RenderParamSWSize(&size), RenderParamSWFormat("rgb0"),
		RenderParamSWStride(&stride), RenderParamSWPointer(unsafe.Pointer(&buffer[0])),
	}
	for _, tc := range []struct {
		name   string
		params []RenderParam
	}{
		{"complete-target", target},
		{"incomplete-target", []RenderParam{RenderParamSWPointer(unsafe.Pointer(&buffer[0]))}},
		{"no-target", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			skip := int32(1)
			params := append([]RenderParam{RenderParamInt(RENDER_PARAM_SKIP_RENDERING, &skip)}, tc.params...)
			if err := rc.Render(params); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buffer, before) {
				t.Fatal("skipped render modified the target")
			}
		})
	}
	skip := int32(0)
	if err := rc.Render(append(target, RenderParamInt(RENDER_PARAM_SKIP_RENDERING, &skip))); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(buffer, before) {
		t.Fatal("rendering with skip=0 did not write the target")
	}
}
