package adventure

import (
	"image/color"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var (
	missingSprite = atlas.GetSprite("2x2")
)

type AnimatedMoveableEntity struct {
	MoveableEntity
	Animations map[MoveState]map[input.Direction]*anim.AnimatedSprite
	Lights     map[MoveState]*Light
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

func (a *AnimatedMoveableEntity) currentLight() *Light {
	if a.Lights == nil {
		return nil
	}
	l, _ := a.Lights[a.MoveState]
	return l
}

func (a *AnimatedMoveableEntity) Update(adv *State, timeDelta float64) {
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
	light := a.currentLight()
	if light != nil {
		light.Update(timeDelta)
	}
}

func (a *AnimatedMoveableEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	var sprite pixelutil.BoundedDrawable
	animation := a.currentAnimation()
	if animation == nil {
		sprite = missingSprite
	} else {
		sprite = a.currentAnimation().Sprite()
	}
	sprite.DrawColorMask(target, matrix, a.ColorMask)
}

func (a *AnimatedMoveableEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	l := a.currentLight()
	if l == nil {
		return
	}
	l.Render(target, matrix)
}

func (a *AnimatedMoveableEntity) RenderMapLocation() pixel.Vec {
	location := a.MoveableEntity.RenderMapLocation()
	location = location.Add(pixel.V(0, 0.25))
	return location
}
