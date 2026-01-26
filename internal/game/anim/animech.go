package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func AnimechIdle(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:idle_down"),
		input.Right: Load(atlas, "adventure/entities/animech/cybermonkey_robot:idle_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/cybermonkey_robot:idle_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:idle_left"),
	}
}

func AnimechWalk(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:walk_down"),
		input.Right: Load(atlas, "adventure/entities/animech/cybermonkey_robot:walk_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/cybermonkey_robot:walk_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:walk_left"),
	}
}

func AnimechRun(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:run_down"),
		input.Right: Load(atlas, "adventure/entities/animech/cybermonkey_robot:run_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/cybermonkey_robot:run_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/cybermonkey_robot:run_left"),
	}
}
