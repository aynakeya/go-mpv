//go:build cgo
// +build cgo

package mpv

import (
	"testing"
	"unsafe"
)

func TestNodeRoundTrip(t *testing.T) {
	n := Node{
		Format: FORMAT_NODE_MAP,
		Value: map[string]Node{
			"name": {Format: FORMAT_STRING, Value: "test"},
			"ok":   {Format: FORMAT_FLAG, Value: true},
			"raw":  {Format: FORMAT_BYTE_ARRAY, Value: ByteArray{0x1, 0x2}},
		},
	}
	cn := n.CNode()
	defer freeMpvDataPointer(FORMAT_NODE, unsafe.Pointer(cn))

	round := newNode(cn)
	if round.Format != FORMAT_NODE_MAP {
		t.Fatalf("unexpected format: %v", round.Format)
	}
	m := round.Value.(map[string]Node)
	if m["name"].Value.(string) != "test" {
		t.Fatalf("unexpected string value: %v", m["name"])
	}
	if m["ok"].Value.(bool) != true {
		t.Fatalf("unexpected bool value: %v", m["ok"])
	}
	raw := []byte(m["raw"].Value.(ByteArray))
	if len(raw) != 2 || raw[0] != 0x1 || raw[1] != 0x2 {
		t.Fatalf("unexpected byte array: %v", raw)
	}
}

func TestNodeFromProperty(t *testing.T) {
	m := newInitializedMpv(t)
	v, err := m.GetProperty("track-list", FORMAT_NODE)
	if err != nil {
		t.Fatalf("GetProperty(track-list) failed: %v", err)
	}
	n := v.(Node)
	if n.Format != FORMAT_NODE_ARRAY {
		t.Fatalf("unexpected node format: %v", n.Format)
	}
}
