package anim

import (
	"fisherevans.com/project/f/internal/resources"
)

func RedCoin(atlas *resources.Atlas) *AnimatedSprite {
	return FromTilesheetRowPartial(atlas, "snowhex_base", 48, 46, 8, 8)
}
