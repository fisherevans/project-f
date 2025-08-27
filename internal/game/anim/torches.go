package anim

import (
	"fisherevans.com/project/f/internal/resources"
)

func TorchTop(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "snowhex_base", 36, 86, 93, 4)
}
