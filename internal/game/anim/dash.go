package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func Dash(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	sheet := "adventure/entities/asha/dash_anim"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, sheet+":dash_down"),
		input.Right: Load(atlas, sheet+":dash_right"),
		input.Up:    Load(atlas, sheet+":dash_up"),
		input.Left:  Load(atlas, sheet+":dash_left"),
	}
}
