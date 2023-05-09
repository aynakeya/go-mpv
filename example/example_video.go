package main

import (
	"fmt"
	"github.com/aynakeya/go-mpv"
	"log"
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
var localfile = "/home/aynakeya/Videos/igotsmoke.mp4"

func main() {
	m := mpv.Create()
	c := eventListener(m)
	log.Println("volume", m.SetOption("volume", mpv.FORMAT_INT64, 100))
	log.Println("terminal", m.SetOptionString("terminal", "no"))

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
	log.Println("loadfile", m.Command([]string{"loadfile", localfile}))

	// getting log messages
	//m.RequestLogMessages(mpv.LOG_LEVEL_INFO)

	//m.ObserveProperty(1, "time-pos", mpv.FORMAT_NODE)
	m.ObserveProperty(1, "time-pos", mpv.FORMAT_STRING)

	fmt.Println(123)
	fmt.Println(m.GetProperty("volume", mpv.FORMAT_NODE))

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
