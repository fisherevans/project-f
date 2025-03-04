package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func Dash(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	fps := 4.0
	sheet := "adventure/entities/asha/dash_anim"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(atlas, sheet, 1, fps),
		input.Right: FromTilesheetRow(atlas, sheet, 2, fps),
		input.Up:    FromTilesheetRow(atlas, sheet, 3, fps),
		input.Left:  FromTilesheetRow(atlas, sheet, 4, fps),
	}
}
