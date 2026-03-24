//go:build !cgo
// +build !cgo

package mpv

type Format int

const (
	FORMAT_NONE       Format = 0
	FORMAT_STRING     Format = 1
	FORMAT_OSD_STRING Format = 2
	FORMAT_FLAG       Format = 3
	FORMAT_INT64      Format = 4
	FORMAT_DOUBLE     Format = 5
	FORMAT_NODE       Format = 6
	FORMAT_NODE_ARRAY Format = 7
	FORMAT_NODE_MAP   Format = 8
	FORMAT_BYTE_ARRAY Format = 9
)
