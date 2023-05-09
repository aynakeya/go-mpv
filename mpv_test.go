package mpv

import (
	"fmt"
	"testing"
)

func TestClientApiVersion(t *testing.T) {
	fmt.Println(ClientApiVersion())
}

func TestMpv(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.ClientName())
	fmt.Println(mpv.Initialize())
	mpv.Destroy()
}

func TestMpv_GetProperty(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.Initialize())
	fmt.Println(mpv.GetPropertyString("idle-actffive"))
	fmt.Println(mpv.GetPropertyString("idle-active"))
	fmt.Println(mpv.GetPropertyOsdString("idle-active"))
	fmt.Println(mpv.GetProperty("idle-active", FORMAT_STRING))
	fmt.Println(mpv.GetProperty("volume", FORMAT_STRING))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	fmt.Println(mpv.GetProperty("audio-device-list", FORMAT_STRING))
	fmt.Println(mpv.GetPropertyString("audio-device-list"))
	mpv.Destroy()
}

func TestMpv_SetProperty(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.Initialize())
	fmt.Println(mpv.GetProperty("volume", FORMAT_STRING))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	fmt.Println(mpv.SetProperty("volume", FORMAT_INT64, 50))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	mpv.Destroy()
}

func TestMpv_SetProperty2(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.Initialize())
	fmt.Println(mpv.GetProperty("volume", FORMAT_STRING))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	fmt.Println(mpv.SetProperty("volume", FORMAT_INT64, 50))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	fmt.Println(mpv.SetPropertyString("volume", "61"))
	//fmt.Println(mpv.SetProperty("volume", FORMAT_STRING, "61"))
	fmt.Println(mpv.GetProperty("volume", FORMAT_INT64))
	mpv.Destroy()
}

func TestMpv_GetProperty2(t *testing.T) {
	mpv := Create()
	fmt.Println(mpv.Initialize())
	fmt.Println(FORMAT_NODE_ARRAY, FORMAT_NODE_MAP)
	n, _ := mpv.GetProperty("audio-device-list", FORMAT_NODE)
	fmt.Println(n.(Node))
	fmt.Println(n.(Node).Format)
	for _, val := range n.(Node).Value.([]Node) {
		fmt.Println(val)
		for key, val2 := range val.Value.(map[string]Node) {
			fmt.Println(key, val2)
		}
	}
	fmt.Println(mpv.GetProperty("audio-device-list", FORMAT_STRING))
	fmt.Println(mpv.GetProperty("audio-device", FORMAT_STRING))
	mpv.Destroy()
}
