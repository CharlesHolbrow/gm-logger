package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/CharlesHolbrow/gm"
	"github.com/rakyll/portmidi"
)

const midiInputDeviceID = 0

func main() {
	gm.PrintDevices()

	deviceInfo := *portmidi.Info(midiInputDeviceID)
	if !deviceInfo.IsInputAvailable {
		panic(fmt.Sprintf("device %d is not an input device", midiInputDeviceID))
	}

	portmidi.Initialize()

	fmt.Printf("Listening for MIDI on %d - %s\n", midiInputDeviceID, deviceInfo.Name)
	myMidiHandler := NewMidiLogger()
	ms, err := gm.MakeMidiStream(midiInputDeviceID, myMidiHandler)
	if err != nil {
		log.Panicln("Error Making Midi Stream", err)
	}
	defer ms.Close()

	out, err := portmidi.NewOutputStream(2, 1024, 0)
	if err != nil {
		log.Fatal(err)
	}
	out.WriteShort(gm.Note{On: true, Note: 64, Vel: 127}.Midi())
	time.Sleep(time.Second)
	out.WriteShort(gm.Note{Note: 64, Vel: 127}.Midi())

	g := sync.WaitGroup{}
	g.Add(1)
	g.Wait()
}
