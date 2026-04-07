package anim

import (
	"fisherevans.com/project/f/internal/resources"
)

func Torch(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "snowhex_base:torch")
}

func TorchLeft(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "snowhex_base:torch_left")
}

func TorchRight(atlas *resources.Atlas) *AnimatedSprite {
	return Load(atlas, "snowhex_base:torch_right")
}
