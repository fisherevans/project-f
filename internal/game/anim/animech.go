package anim

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
)

func AnimechIdle(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/altered_knight_idle:idle_down"),
		input.Right: Load(atlas, "adventure/entities/animech/altered_knight_idle:idle_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/altered_knight_idle:idle_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/altered_knight_idle:idle_left"),
	}
}

func AnimechWalk(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/altered_knight_walk:walk_down"),
		input.Right: Load(atlas, "adventure/entities/animech/altered_knight_walk:walk_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/altered_knight_walk:walk_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/altered_knight_walk:walk_left"),
	}
}

func AnimechRun(atlas *resources.Atlas) map[input.Direction]*AnimatedSprite {
	return map[input.Direction]*AnimatedSprite{
		input.Down:  Load(atlas, "adventure/entities/animech/altered_knight_run:run_down"),
		input.Right: Load(atlas, "adventure/entities/animech/altered_knight_run:run_right"),
		input.Up:    Load(atlas, "adventure/entities/animech/altered_knight_run:run_up"),
		input.Left:  Load(atlas, "adventure/entities/animech/altered_knight_run:run_left"),
	}
}
