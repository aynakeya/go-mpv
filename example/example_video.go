package main

import (
	"log"
	"mpv"
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

func main() {
	m := mpv.Create()
	c := eventListener(m)
	log.Println("volume", m.SetOption("volume", mpv.FORMAT_INT64, 64))
	log.Println("terminal", m.SetOptionString("terminal", "no"))

	//log.Println("video", m.SetOption("video", mpv.FORMAT_STRING, "no"))
	log.Println("vo=null", m.SetOptionString("vo", "null"))
	//log.Println("vid", m.SetOption("vid", mpv.FORMAT_STRING, "no"))

	err := m.Initialize()
	if err != nil {
		log.Println("Mpv init:", err.Error())
		return
	}
	//Set video file
	log.Println("loadfile", m.Command([]string{"loadfile", ymca}))

	for {
		e := <-c
		//log.Println(e)
		if e.EventId == mpv.EVENT_END_FILE {
			break
		}

	}
	m.TerminateDestroy()
}
