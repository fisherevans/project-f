package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func AshaIdle(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	fps := 4.0
	sheet := "adventure/entities/asha/asha_idle"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(atlas, sheet, 1, fps),
		input.Right: FromTilesheetRow(atlas, sheet, 2, fps),
		input.Up:    FromTilesheetRow(atlas, sheet, 3, fps),
		input.Left:  FromTilesheetRow(atlas, sheet, 4, fps),
	}
}

func AshaWalk(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	fps := 5.0
	sheet := "adventure/entities/asha/asha_walk"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(atlas, sheet, 1, fps),
		input.Right: FromTilesheetRow(atlas, sheet, 2, fps),
		input.Up:    FromTilesheetRow(atlas, sheet, 3, fps),
		input.Left:  FromTilesheetRow(atlas, sheet, 4, fps),
	}
}

func AshaRun(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	fps := 2.5
	sheet := "adventure/entities/asha/asha_run"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  FromTilesheetRow(atlas, sheet, 1, fps),
		input.Right: FromTilesheetRow(atlas, sheet, 2, fps),
		input.Up:    FromTilesheetRow(atlas, sheet, 3, fps),
		input.Left:  FromTilesheetRow(atlas, sheet, 4, fps),
	}
}
