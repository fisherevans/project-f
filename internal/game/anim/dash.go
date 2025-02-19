package anim

import "fisherevans.com/project/f/internal/game/input"

func Dash() map[input.Direction]*AnimatedSprite {
	fps := 4.0
	sheet := "dash_anim"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(sheet, 1, fps),
		input.Right: FromTilesheetRow(sheet, 2, fps),
		input.Up:    FromTilesheetRow(sheet, 3, fps),
		input.Left:  FromTilesheetRow(sheet, 4, fps),
	}
}
