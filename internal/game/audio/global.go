package audio

import (
	"fmt"
	"time"
)

var system *System

func init() {
	var err error
	system, err = newSystem()
	if err != nil {
		panic(fmt.Sprintf("failed to start audio system: %v", err))
	}
	system.SetBusGain(system.Buses.Music, -2)
}

func GetSystem() *System {
	return system
}

// WaitUntilReady blocks until the audio speaker has started streaming.
// This ensures the initial audio buffer is filled and sounds won't be clipped.
func WaitUntilReady() {
	if system == nil {
		panic("audio system not initialized")
	}
	// Wait for 2-3 buffer cycles to ensure speaker thread has started
	// and initial buffers are filled. Buffer size is 50ms, so 150ms is safe.
	time.Sleep(150 * time.Millisecond)
}
