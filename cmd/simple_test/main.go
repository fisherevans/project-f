package main

import (
	"fmt"
	"os"
	"time"

	"fisherevans.com/project/f/internal/game/audio/speech"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const sampleRate = beep.SampleRate(48000)

func main() {
	// Initialize speaker
	if err := speaker.Init(sampleRate, sampleRate.N(time.Second/10)); err != nil {
		fmt.Printf("Failed to initialize speaker: %v\n", err)
		os.Exit(1)
	}

	// Create hollow tick generator
	gen := speech.NewTickGenerator(sampleRate)

	// Generate and play
	fmt.Println("Playing hollow tick: 'hello world'")
	sound := gen.Speak("hello world")

	done := make(chan bool)
	speaker.Play(beep.Seq(sound, beep.Callback(func() {
		done <- true
	})))

	// Wait for playback to complete
	<-done
	fmt.Println("Done!")
}
