package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aynakeya/go-mpv"
)

func eventListener(m *mpv.Mpv) chan *mpv.Event {
	c := make(chan *mpv.Event)
	go func() {
		for {
			e := m.WaitEvent(1)
			c <- e
		}
	}()
	return c
}

var ymca = "https://ia600809.us.archive.org/19/items/VillagePeopleYMCAOFFICIALMusicVideo1978/Village%20People%20-%20YMCA%20OFFICIAL%20Music%20Video%201978.mp4"
var rickroll = "https://fwesh.yonle.repl.co"

func videoPath() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	for _, p := range []string{"data/test.mp4", "../../data/test.mp4"} {
		if _, err := os.Stat(p); err == nil {
			abs, err := filepath.Abs(p)
			if err == nil {
				return abs
			}
			return p
		}
	}
	return ymca
}

func audioDevice() string {
	if len(os.Args) > 2 {
		return os.Args[2]
	}
	return os.Getenv("MPV_AUDIO_DEVICE")
}

func main() {
	m := mpv.Create()
	c := eventListener(m)
	log.Println("audio-client-name", m.SetOptionString("audio-client-name", "AynaMpvCore"))
	log.Println("volume", m.SetOption("volume", mpv.FORMAT_INT64, 30))
	log.Println("terminal", m.SetOptionString("terminal", "no"))
	if dev := audioDevice(); dev != "" {
		log.Println("audio-device", m.SetOptionString("audio-device", dev))
	}
	m.SetPropertyString("audio-device", "pulse/alsa_output.pci-0000_75_00.6.analog-stereo")

	//log.Println("video", m.SetOption("video", mpv.FORMAT_STRING, "no"))
	//log.Println("vo=null", m.SetOption("vo", mpv.FORMAT_STRING, "null"))
	//log.Println("vo=null", m.SetOptionString("vo", "null"))
	//log.Println("vo=null", m.SetPropertyString("vo", "null"))
	//log.Println("vid", m.SetOption("vid", mpv.FORMAT_STRING, "no"))

	err := m.Initialize()

	if err != nil {
		log.Println("Mpv init:", err.Error())
		return
	}
	//Set video file
	log.Println("loadfile", m.Command([]string{"loadfile", videoPath()}))

	// getting log messages
	//m.RequestLogMessages(mpv.LOG_LEVEL_INFO)

	//m.ObserveProperty(1, "time-pos", mpv.FORMAT_NODE)
	m.ObserveProperty(1, "time-pos", mpv.FORMAT_STRING)

	fmt.Println(123)
	fmt.Println(m.GetProperty("volume", mpv.FORMAT_NODE))
	fmt.Println(m.GetProperty("ao-volume", mpv.FORMAT_NODE))

	for {
		e := <-c
		//log.Println(e)
		if e.EventId == mpv.EVENT_LOG_MESSAGE {
			fmt.Println(e.LogMessage())
		}
		if e.EventId == mpv.EVENT_PROPERTY_CHANGE {
			fmt.Println(e.Property())
		}
		if e.EventId == mpv.EVENT_END_FILE {
			fmt.Println("end file")
			break
		}

	}
	m.TerminateDestroy()
}
