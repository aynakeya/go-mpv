//go:build !cgo
// +build !cgo

package mpv

type Node struct {
	Value  interface{}
	Format Format
}

type ByteArray []byte
