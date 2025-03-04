package anim

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

type AnimatedSprite struct {
	frames          []pixelutil.BoundedDrawable
	framesPerSecond float64
	progression     float64
	currentFrame    int
}

func (a *AnimatedSprite) ApplyPingPong() *AnimatedSprite {
	for i := len(a.frames) - 2; i > 0; i-- {
		a.frames = append(a.frames, a.frames[i])
	}
	return a
}

func FromTilesheetRow(atlas *resources.Atlas, tilesheet string, row int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	ts := resources.GetTilesheet(tilesheet)
	for col := 1; col <= ts.Columns; col++ {
		animated.frames = append(animated.frames, atlas.GetTilesheetSprite(tilesheet, col, row))
	}
	return animated
}

func (a *AnimatedSprite) Sprite() pixelutil.BoundedDrawable {
	return a.frames[a.currentFrame]
}

func (a *AnimatedSprite) Update(timeDelta float64) {
	a.progression += timeDelta
	secondsPerFrame := 1.0 / a.framesPerSecond
	for a.progression >= secondsPerFrame {
		a.progression -= secondsPerFrame
		a.currentFrame++
	}
	a.currentFrame = a.currentFrame % len(a.frames)
}

func (a *AnimatedSprite) Reset() {
	a.currentFrame = 0
	a.progression = 0
}
