package game

import (
	"math"

	"fisherevans.com/project/f/internal/game/audio"
)

type ContextUtils struct {
}

func (u ContextUtils) TimeCycleSin(speed float64) float64 {
	return (math.Sin(ctx.elapsed*speed*2.0*math.Pi) + 1.0) / 2.0
}

// GetAudioSystem returns the global audio system
func GetAudioSystem() *audio.System {
	return audio.GetSystem()
}
