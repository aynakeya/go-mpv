package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aynakeya/go-mpv"
)

var ymca = "https://ia600809.us.archive.org/19/items/VillagePeopleYMCAOFFICIALMusicVideo1978/Village%20People%20-%20YMCA%20OFFICIAL%20Music%20Video%201978.mp4"

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
	if m == nil {
		log.Println("mpv.Create returned nil")
		return
	}
	defer m.TerminateDestroy()
	log.Println("audio-client-name", m.SetOptionString("audio-client-name", "AynaMpvCore"))
	log.Println("volume", m.SetOption("volume", mpv.FORMAT_INT64, 30))
	log.Println("terminal", m.SetOptionString("terminal", "no"))
	if dev := audioDevice(); dev != "" {
		log.Println("audio-device", m.SetOptionString("audio-device", dev))
	}

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
	if err := m.Command([]string{"loadfile", videoPath()}); err != nil {
		log.Println("loadfile:", err)
		return
	}

	// getting log messages
	//m.RequestLogMessages(mpv.LOG_LEVEL_INFO)

	//m.ObserveProperty(1, "time-pos", mpv.FORMAT_NODE)
	if err := m.ObserveProperty(1, "time-pos", mpv.FORMAT_STRING); err != nil {
		log.Println("ObserveProperty:", err)
		return
	}
	fmt.Println(m.GetProperty("volume", mpv.FORMAT_NODE))
	fmt.Println(m.GetProperty("ao-volume", mpv.FORMAT_NODE))

	for {
		// Event payloads belong to libmpv until the next WaitEvent. Process
		// them here so neither another wait nor destruction can race a read.
		e := m.WaitEvent(1)
		//log.Println(e)
		if e.EventId == mpv.EVENT_LOG_MESSAGE {
			fmt.Println(e.LogMessage())
		}
		if e.EventId == mpv.EVENT_PROPERTY_CHANGE {
			fmt.Println(e.Property())
		}
		if e.EventId == mpv.EVENT_END_FILE || e.EventId == mpv.EVENT_SHUTDOWN {
			fmt.Println("end file")
			break
		}

	}
}
