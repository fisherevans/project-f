package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
)

type LightEntity struct {
	InnateEntity
	Light
	Animations []*anim.AnimatedSprite
	Passable
}

func (i *LightEntity) Update(adv *State, timeDelta float64) {
	i.Light.Update(timeDelta)
	for _, a := range i.Animations {
		a.Update(timeDelta)
	}
}

func (i *LightEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	for _, a := range i.Animations {
		a.Sprite().Draw(target, matrix)
	}
}

func (i *LightEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	i.Light.Render(target, matrix)
}
