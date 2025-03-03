package anim

import "fisherevans.com/project/f/internal/game/input"

func AshaIdle() map[input.Direction]*AnimatedSprite {
	fps := 4.0
	sheet := "adventure/entities/asha/asha_idle"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(sheet, 1, fps),
		input.Right: FromTilesheetRow(sheet, 2, fps),
		input.Up:    FromTilesheetRow(sheet, 3, fps),
		input.Left:  FromTilesheetRow(sheet, 4, fps),
	}
}

func AshaWalk() map[input.Direction]*AnimatedSprite {
	fps := 5.0
	sheet := "adventure/entities/asha/asha_walk"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(sheet, 1, fps),
		input.Right: FromTilesheetRow(sheet, 2, fps),
		input.Up:    FromTilesheetRow(sheet, 3, fps),
		input.Left:  FromTilesheetRow(sheet, 4, fps),
	}
}

func AshaRun() map[input.Direction]*AnimatedSprite {
	fps := 2.5
	sheet := "adventure/entities/asha/asha_run"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(sheet, 1, fps),
		input.Right: FromTilesheetRow(sheet, 2, fps),
		input.Up:    FromTilesheetRow(sheet, 3, fps),
		input.Left:  FromTilesheetRow(sheet, 4, fps),
	}
}
