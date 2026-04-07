package anim

import (
	"fisherevans.com/project/f/internal/resources"
)

func RedCoin(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "snowhex_base:red_coin")
}
