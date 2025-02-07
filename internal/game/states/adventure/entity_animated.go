package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"github.com/gopxl/pixel/v2"
	"image/color"
)

type AnimatedMoveableEntity struct {
	MoveableEntity
	Animations map[input.Direction]*anim.AnimatedSprite
	ColorMask  color.Color
}

func (a *AnimatedMoveableEntity) currentAnimation() *anim.AnimatedSprite {
	animation, exists := a.Animations[a.FacingDirection]
	if !exists {
		animation, exists = a.Animations[input.Down]
	}
	return animation
}

func (a *AnimatedMoveableEntity) Update(ctx *game.Context, adv *State, timeDelta float64) {
	animation := a.currentAnimation()
	if animation == nil {
		return
	}
	if a.Moving {
		animation.Update(timeDelta * a.MoveSpeed)
		return
	}
	animation.Reset()
}

func (a *AnimatedMoveableEntity) Render(target pixel.Target, matrix pixel.Matrix) {
	var sprite *pixel.Sprite
	animation := a.currentAnimation()
	if animation == nil {
		sprite = util.MissingSprite
	} else {
		sprite = a.currentAnimation().Sprite()
	}
	sprite.DrawColorMask(target, matrix, a.ColorMask)
}

func (a *AnimatedMoveableEntity) RenderMapLocation() pixel.Vec {
	location := a.MoveableEntity.RenderMapLocation()
	location = location.Add(pixel.V(0, 0.25))
	return location
}
