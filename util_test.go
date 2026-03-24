//go:build cgo
// +build cgo

package mpv

import "testing"

func TestMallocAndFreeMpvDataPointer(t *testing.T) {
	cases := []struct {
		name   string
		format Format
		data   interface{}
	}{
		{name: "none", format: FORMAT_NONE, data: nil},
		{name: "string", format: FORMAT_STRING, data: "hello"},
		{name: "flag", format: FORMAT_FLAG, data: true},
		{name: "int64", format: FORMAT_INT64, data: int64(42)},
		{name: "double", format: FORMAT_DOUBLE, data: 3.14},
		{name: "byte_array", format: FORMAT_BYTE_ARRAY, data: ByteArray{1, 2, 3}},
		{
			name:   "node",
			format: FORMAT_NODE,
			data: Node{
				Format: FORMAT_NODE_ARRAY,
				Value: []Node{
					{Format: FORMAT_STRING, Value: "v"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ptr := mallocMpvDataPointer(tc.format, tc.data)
			freeMpvDataPointer(tc.format, ptr)
		})
	}
}

func TestFreeMpvDataPointerNilIsSafe(t *testing.T) {
	freeMpvDataPointer(FORMAT_NONE, nil)
}
