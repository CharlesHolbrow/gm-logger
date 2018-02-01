package main

import (
	"fmt"
	"time"

	"github.com/CharlesHolbrow/gm"
)

// MidiLogger is a Midi handler that logs some info a bout incoming midi events.
// It make a reasonable effort to keep track of how many 16th notes have elapsed
// on the incoming midi clock. However, it is impossible to keep perfect track
// as esplained in the source code comments.
type MidiLogger struct {
	keys [16][128]uint8
	ccs  [16][128]uint8

	isPlaying bool

	// When receiving MIDI clock, when playback begins, the host sends a `start`
	// message instantly followed by a `tick` message. The first `tick` received
	// does not always indicate that the song position pointer has advanced.
	// When looping, for example we should count all ticks as an advance. When
	// stopped, the first tick does not signify an advance.
	isTicking bool

	// Try to keep track of how many sixteenth notes have elapsed in the song.
	// Unfortunately there is a significant barrier. SPP is very coarse. It's
	// resolution is only 16th notes, so when the transport moves in the host,
	// slaves do not know precisely where to seek to.
	//
	// When reaper pauses and continues it sends messages like this:
	// 1. Tick
	// 2. Stop
	// 3. SongPosition (to the CLOSEST 16th note - possibly ahead or behind)
	// -- unpause --
	// 4. Continue
	// 5. Tick (not immediately, but on the next scheduled tick)
	// The unreliability in step 3 makes it difficult to follow the actual
	// position of the pause.
	//
	// When reaper loops on sixteenth notes, it sends messages like this:
	// 1. Clock tick (at the very end)
	// 2. Note off for hanging notes
	// 3. Note on for notes that begin at the begining of the loop (weird)
	// 4. Midi stop
	// -- reaper transport jumps back - no clock tick sent --
	// 5. Midi start (wait for one tick)
	// 6. Clock tick
	// Steps 1, 2, 3, 4, 5 all happem within a millisecond
	//
	// The implementation here assumes that we only seek exactly to 18th notes.
	// If the track is pausing or seeking to anything other than 16th notes,
	// both sixtheenthsElapsed and beatPosition should be treated as estimates.
	sixteenthsElapsed int

	// There are 24 midi ticks per quarter, so there are 6 per 16th note.
	// beatPosition counts from 0 to 5 repeatedly.
	beatPosition int

	startTime time.Time
	stopTime  time.Time
}

// NewMidiLogger creates a New MidiLogger with the start time initialized.
func NewMidiLogger() *MidiLogger {
	return &MidiLogger{
		startTime: time.Now(),
	}
}

// HandleNote processing incoming Note ons and Note Offs
func (mh *MidiLogger) HandleNote(n gm.Note) {
	if n.Vel == 0 {
		n.On = false
	}
	notes := mh.keys[int(n.Ch)]

	// Update our note maps maps
	if n.On {
		notes[n.Note] = n.Vel
	} else {
		notes[n.Note] = 0
	}
	fmt.Printf("%s - %d.%d (%v)\n", n, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
}

// HandleMisc sync messages including start, stop, continue, spp
func (mh *MidiLogger) HandleMisc(event interface{}) {
	switch t := event.(type) {
	case gm.Clock:
		if !mh.isPlaying {
			break
		}

		// At bpm = 120, ticks are about 23 ms appart. There's no perfect wy to
		// do this, but this allows us to "loop" in reaper. See the longer
		// comment above for why this is needed.
		if !mh.isTicking {
			mh.isTicking = true
			break
		}

		mh.beatPosition = (mh.beatPosition + 1) % 6
		if mh.beatPosition == 0 {
			mh.sixteenthsElapsed++
			mh.onSixteenth()
		}
		if mh.sixteenthsElapsed%16 == 0 {
			mh.onWhole()
		}
	case gm.SPP:
		mh.beatPosition = 0
		mh.sixteenthsElapsed = int(t)
		// When we move the cursor in reaper, reaper smartly sends note off events
		// for active notes in midi items. However, it does not send note off
		// events for notes held on the keyboard.
	case gm.Start:
		mh.startTime = time.Now()
		mh.beatPosition = 0
		mh.sixteenthsElapsed = 0
		mh.isPlaying = true
	case gm.Continue:
		mh.isPlaying = true
		if time.Since(mh.stopTime) > time.Millisecond {
			mh.isTicking = false
		}
	case gm.Stop:
		mh.stopTime = time.Now()
		mh.isPlaying = false

	default:
		fmt.Println("midiHandler.handleMisc received", t)
	}
	fmt.Printf("%s - %d.%d (%v)\n", event, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
}

// HandleCC handles incoming CC events. Part of the MidiHandler interface.
func (mh *MidiLogger) HandleCC(cc gm.CC) {
	mh.ccs[cc.Ch][cc.Number] = cc.Value
	switch cc.Number {
	case 64:
		if cc.Value >= 64 {
			fmt.Printf("Pedal Down - %s - %d.%d (%v)\n", cc, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
		} else {
			fmt.Printf("Pedal Up - %s - %d.%d (%v)\n", cc, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
		}
	case 123:
		fmt.Printf("All Notes Off - %s - %d.%d (%v)\n", cc, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
	default:
		fmt.Printf("%s - %d.%d (%v)\n", cc, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
	}
}

func (mh *MidiLogger) onSixteenth() {
}

func (mh *MidiLogger) onWhole() {
}

// HandlePW handles incoming Pitch Wehll events. Part of the MidiHandler interface.
func (mh *MidiLogger) HandlePW(pw gm.PitchWheel) {
	fmt.Printf("%s - %d.%d (%v)\n", pw, mh.sixteenthsElapsed, mh.beatPosition, time.Since(mh.startTime))
}
