package anim

import (
	"fisherevans.com/project/f/internal/resources"
)

func Torch(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "snowhex_base", 36, 96, 8, 8)
}

func TorchLeft(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "snowhex_base", 40, 96, 8, 8)
}

func TorchRight(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "snowhex_base", 40, 104, 8, 8)
}
