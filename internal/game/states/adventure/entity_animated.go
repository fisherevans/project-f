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
	Animations map[MoveState]map[input.Direction]*anim.AnimatedSprite
	ColorMask  color.Color

	lastUpdateDirection input.Direction
	lastUpdateState     MoveState
}

func (a *AnimatedMoveableEntity) currentAnimation() *anim.AnimatedSprite {
	moveAnimations, exists := a.Animations[a.MoveState]
	if !exists {
		moveAnimations, exists = a.Animations[MoveStateIdle]
	}
	animation, animExists := moveAnimations[a.FacingDirection]
	if !animExists {
		animation = moveAnimations[input.Down]
	}
	return animation
}

func (a *AnimatedMoveableEntity) Update(ctx *game.Context, adv *State, timeDelta float64) {
	animation := a.currentAnimation()
	if animation == nil {
		return
	}
	if a.lastUpdateDirection != a.FacingDirection || a.lastUpdateState != a.MoveState {
		animation.Reset()
	}
	a.lastUpdateDirection = a.FacingDirection
	a.lastUpdateState = a.MoveState
	if a.IsMoving() {
		animation.Update(timeDelta * a.GetCurrentSpeed())
	} else {
		animation.Update(timeDelta)
	}
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
