package anim

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
)

type AnimatedSprite struct {
	frames          []*pixel.Sprite
	framesPerSecond float64
	progression     float64
	currentFrame    int
}

func FromSprites(tilesheet string, row, startCol, endCol int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	for col := startCol; col <= endCol; col++ {
		ref := resources.TilesheetSprites[resources.TilesheetSpriteId{
			Tilesheet: tilesheet,
			Row:       row,
			Column:    col,
		}]
		animated.frames = append(animated.frames, ref.Sprite)
	}
	return animated
}

func FromTilesheetRow(tilesheet string, row int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	ts := resources.Tilesheets[tilesheet]
	for col := 1; col <= ts.Columns; col++ {
		ref := resources.TilesheetSprites[resources.TilesheetSpriteId{
			Tilesheet: tilesheet,
			Row:       row,
			Column:    col,
		}]
		animated.frames = append(animated.frames, ref.Sprite)
	}
	return animated
}

func (a *AnimatedSprite) Sprite() *pixel.Sprite {
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
