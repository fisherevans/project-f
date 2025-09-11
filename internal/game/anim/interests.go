package anim

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/tiles"
)

func PDA(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "space_base", 29, 1, 8, 6)
}

func PDADisabled(atlas *resources.Atlas) *AnimatedSprite {
	return NewStaticAnimationFromId(atlas, tiles.PDADisabled)
}
