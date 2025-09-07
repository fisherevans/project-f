package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
)

type LightEntity struct {
	InnateEntity
	Light
	Animation *anim.AnimatedSprite
}

func (i *LightEntity) Update(ctx *game.Context, adv *State, timeDelta float64) {
	i.Light.Update(timeDelta)
	if i.Animation != nil {
		i.Animation.Update(timeDelta)
	}
}

func (i *LightEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if i.Animation != nil {
		i.Animation.Sprite().Draw(target, matrix)
	}
}

func (i *LightEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	i.Light.Render(target, matrix)
}
