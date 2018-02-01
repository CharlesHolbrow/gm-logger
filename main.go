package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CharlesHolbrow/gm"
	"github.com/rakyll/portmidi"
)

var inputDeviceInt = flag.Int("device", -1, "ID of the device to use")

func main() {
	flag.Parse()
	gm.PrintDevices()

	// Select an input device
	if *inputDeviceInt == -1 {
		*inputDeviceInt = int(portmidi.DefaultInputDeviceID())
		fmt.Println("No device specified. Using the default device")
	}

	// Tell the user which device will be used
	inputID := portmidi.DeviceID(*inputDeviceInt)
	deviceInfo := *portmidi.Info(inputID)
	if !deviceInfo.IsInputAvailable {
		panic(fmt.Sprintf("Device %d is not an input device", inputID))
	}
	fmt.Printf("Listening for MIDI on %d - %s %v\n\n", inputID, deviceInfo.Name, deviceInfo.IsInputAvailable)

	// Open the device, and start logging
	portmidi.Initialize()
	ms, err := gm.MakeMidiStream(int(inputID), NewMidiLogger())
	if err != nil {
		log.Panicln("Error Making Midi Stream", err)
	}
	defer ms.Close()

	// Wait for an exit signal
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal
}
