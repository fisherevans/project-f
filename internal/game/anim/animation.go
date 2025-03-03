package anim

import (
	"fisherevans.com/project/f/internal/resources"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type AnimatedSprite struct {
	frames          []*pixel.Sprite
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

func FromSprites(tilesheet string, row, startCol, endCol int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	for col := startCol; col <= endCol; col++ {
		ref := resources.GetTilesheetSprite(tilesheet, col, row)
		animated.frames = append(animated.frames, ref.Sprite)
	}
	return animated
}

func FromTilesheetRow(tilesheet string, row int, framesPerSecond float64) *AnimatedSprite {
	animated := &AnimatedSprite{
		framesPerSecond: framesPerSecond,
	}
	ts := resources.GetTilesheet(tilesheet)
	for col := 1; col <= ts.Columns; col++ {
		spriteId := resources.TilesheetSpriteId{
			Tilesheet: tilesheet,
			Row:       row,
			Column:    col,
		}
		ref := resources.GetTilesheetSpriteById(spriteId)
		if ref == nil {
			log.Fatal().Str("tilesheet", tilesheet).Int("row", row).Int("col", col).Msg("sprite not found when making animation")
		}
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
