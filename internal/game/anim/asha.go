package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func AshaIdle(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	sheet := "adventure/entities/asha_space/asha_idle"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, sheet+":idle_down"),
		input.Right: Load(atlas, sheet+":idle_right"),
		input.Up:    Load(atlas, sheet+":idle_up"),
		input.Left:  Load(atlas, sheet+":idle_left"),
	}
}

func AshaWalk(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	sheet := "adventure/entities/asha_space/asha_walk"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, sheet+":walk_down"),
		input.Right: Load(atlas, sheet+":walk_right"),
		input.Up:    Load(atlas, sheet+":walk_up"),
		input.Left:  Load(atlas, sheet+":walk_left"),
	}
}

func AshaRun(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	sheet := "adventure/entities/asha_space/asha_run"
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, sheet+":run_down"),
		input.Right: Load(atlas, sheet+":run_right"),
		input.Up:    Load(atlas, sheet+":run_up"),
		input.Left:  Load(atlas, sheet+":run_left"),
	}
}
