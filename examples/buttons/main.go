package main

import phatbeat "phatbeat/lib"

import "log"

func main() {

	log.Print("pHAT BEAT: Buttons \n Shows off the various ways you can configure buttons on PHAT BEAT. \n Press Ctrl+C to exit!\n")

	phatbeat.On(phatbeat.FastForward, true, func(p int) {
		log.Println("Fast Forward")
	})

	phatbeat.Hold(phatbeat.FastForward, false, 2, func(b int) {
		log.Println("FF Held")
	})

	phatbeat.On(phatbeat.PlayPause, false, func(b int) {
		log.Println("PP")
	})

	phatbeat.Hold(phatbeat.PlayPause, false, 2, func(b int) {
		log.Println("PP held")
	})

	phatbeat.On(phatbeat.VolDown, true, func(p int) {
		log.Println("VolDown")
	})

	phatbeat.Hold(phatbeat.VolDown, false, 2, func(b int) {
		log.Println("VolDown")
	})

	phatbeat.On(phatbeat.VolUp, true, func(p int) {
		log.Println("VolUp")
	})

	phatbeat.Hold(phatbeat.VolUp, false, 2, func(b int) {
		log.Println("VolUp")
	})

	phatbeat.On(phatbeat.Rewind, true, func(p int) {
		log.Println("Rewind")
	})

	phatbeat.Hold(phatbeat.Rewind, false, 2, func(b int) {
		log.Println("Rewind")
	})

	phatbeat.On(phatbeat.OnOff, true, func(p int) {
		log.Println("OnOff")
	})

	phatbeat.Hold(phatbeat.OnOff, false, 2, func(b int) {
		log.Println("OnOff")
	})

	defer phatbeat.Clean()

	select {}
}
