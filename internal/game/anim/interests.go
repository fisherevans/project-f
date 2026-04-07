package anim

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/tiles"
)

func PDA(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "space_base:pda")
}

func PDADisabled(atlas *resources.Atlas) *AnimatedSprite {
	return NewStaticAnimationFromId(atlas, tiles.PDADisabled)
}
