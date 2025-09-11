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

func NewStaticAnimation(frame pixelutil.BoundedDrawable) *AnimatedSprite {
	return &AnimatedSprite{
		frames:          []pixelutil.BoundedDrawable{frame},
		framesPerSecond: 1,
	}
}

func NewStaticAnimationFromId(atlas *resources.Atlas, tileId resources.TilesheetSpriteId) *AnimatedSprite {
	return NewStaticAnimation(atlas.GetTilesheetSpriteById(tileId))
}

func (a *AnimatedSprite) ApplyPingPong() *AnimatedSprite {
	for i := len(a.frames) - 2; i > 0; i-- {
		a.frames = append(a.frames, a.frames[i])
	}
	return a
}

func FromTilesheetRow(atlas *resources.Atlas, tilesheet string, row int, framesPerSecond float64) *AnimatedSprite {
	ts := resources.GetTilesheet(tilesheet)
	return FromTilesheetRowPartial(atlas, tilesheet, row, 1, ts.Columns, framesPerSecond)
}

func FromTilesheetRowPartial(atlas *resources.Atlas, tilesheet string, row, colFrom, columns int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	for col := colFrom; col < colFrom+columns; col++ {
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
