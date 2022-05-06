package mpv

import (
	"fmt"
	"testing"
)

func TestNode(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.Initialize())
	fmt.Println(FORMAT_NODE_ARRAY, FORMAT_NODE_MAP)
	n, _ := mpv.GetProperty("audio-device-list", FORMAT_NODE)
	fmt.Println(n.(Node))
	renode := NewNode(n.(Node).CNode())
	fmt.Println(renode)
	mpv.Destroy()
}
