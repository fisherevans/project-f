package anim

import (
	"fisherevans.com/project/f/resources"
	"github.com/gopxl/pixel/v2"
)

type AnimatedSprite struct {
	frames       []*pixel.Sprite
	timePerFrame float64
	progression  float64
	currentFrame int
}

func FromSprites(tilesheet string, row, startCol, endCol int, timePerFrame float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		timePerFrame: timePerFrame,
	}
	for col := startCol; col <= endCol; col++ {
		ref := resources.Sprites[resources.SpriteId{
			Tilesheet: tilesheet,
			Row:       row,
			Column:    col,
		}]
		animated.frames = append(animated.frames, ref.Sprite)
	}
	return animated
}

func FromTilesheetRow(tilesheet string, row int, timePerFrame float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		timePerFrame: timePerFrame,
	}
	ts := resources.Tilesheets[tilesheet]
	for col := 1; col <= ts.Columns; col++ {
		ref := resources.Sprites[resources.SpriteId{
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
	for a.progression >= a.timePerFrame {
		a.progression -= a.timePerFrame
		a.currentFrame++
	}
	a.currentFrame = a.currentFrame % len(a.frames)
}

func (a *AnimatedSprite) Reset() {
	a.currentFrame = 0
	a.progression = 0
}
