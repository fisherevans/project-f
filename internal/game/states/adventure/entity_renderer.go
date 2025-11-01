package adventure

import (
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

type EntityRenderer interface {
	Update(timeDelta float64)
	RenderToScene(target pixel.Target, matrix pixel.Matrix)
	RenderToLightMap(target pixel.Target, matrix pixel.Matrix)
	ZPriority() int
}

type AnimationSpeedScaler interface {
	GetAnimationSpeedScale() float64
}

type MoveAnimationSpeedScaler struct {
	entity Entity
}

func NewMoveAnimationSpeedScaler(entity Entity) *MoveAnimationSpeedScaler {
	return &MoveAnimationSpeedScaler{
		entity: entity,
	}
}

func (s *MoveAnimationSpeedScaler) GetAnimationSpeedScale() float64 {
	return s.entity.GetMovementSpeed(s.entity.GetMovementState())
}

type BasicEntityRenderer struct {
	entity                Entity
	zPriority             int
	lights                []*Light
	lightOriginOffset     pixel.Vec
	animations            []ColorMaskAnimation
	animationOriginOffset pixel.Vec
	animationSpeedScaler  AnimationSpeedScaler
}

func NewBasicEntityRenderer(entity Entity) *BasicEntityRenderer {
	return &BasicEntityRenderer{
		entity: entity,
	}
}

type ColorMaskAnimation struct {
	Animation *anim.AnimatedSprite
	ColorMask *pixel.RGBA
}

func NewColorMaskAnimation(animation *anim.AnimatedSprite) ColorMaskAnimation {
	return ColorMaskAnimation{
		Animation: animation,
	}
}

func (cma ColorMaskAnimation) WithColorMask(mask pixel.RGBA) ColorMaskAnimation {
	cma.ColorMask = &mask
	return cma
}

func (r *BasicEntityRenderer) WithZPriority(zPriority int) *BasicEntityRenderer {
	r.zPriority = zPriority
	return r
}

func (r *BasicEntityRenderer) WithLights(lights ...*Light) *BasicEntityRenderer {
	r.lights = lights
	return r
}

func (r *BasicEntityRenderer) WithLightOriginOffset(offset pixel.Vec) *BasicEntityRenderer {
	r.lightOriginOffset = offset
	return r
}

func (r *BasicEntityRenderer) WithAnimations(animations ...*anim.AnimatedSprite) *BasicEntityRenderer {
	var cmas []ColorMaskAnimation
	for _, a := range animations {
		cmas = append(cmas, ColorMaskAnimation{Animation: a})
	}
	return r.WithColorMaskAnimations(cmas...)
}

func (r *BasicEntityRenderer) WithColorMaskAnimations(animations ...ColorMaskAnimation) *BasicEntityRenderer {
	r.animations = animations
	return r
}

func (r *BasicEntityRenderer) WithAnimationOriginOffset(offset pixel.Vec) *BasicEntityRenderer {
	r.animationOriginOffset = offset
	return r
}

func (r *BasicEntityRenderer) WithAnimationSpeedScaler(speedScaler AnimationSpeedScaler) *BasicEntityRenderer {
	r.animationSpeedScaler = speedScaler
	return r
}

func (r *BasicEntityRenderer) Reset() {
	for _, a := range r.animations {
		a.Animation.Reset()
	}
}

func (r *BasicEntityRenderer) ZPriority() int {
	return r.zPriority
}

func (r *BasicEntityRenderer) Update(timeDelta float64) {
	animationSpeedScale := 1.0
	if r.animationSpeedScaler != nil {
		animationSpeedScale = r.animationSpeedScaler.GetAnimationSpeedScale()
	}
	for _, a := range r.animations {
		a.Animation.Update(timeDelta * animationSpeedScale)
	}
	for _, l := range r.lights {
		l.Update(timeDelta)
	}
}

func (r *BasicEntityRenderer) RenderToScene(target pixel.Target, matrix pixel.Matrix) {
	r.RenderToSceneAlpha(target, matrix, 1)
}

func (r *BasicEntityRenderer) RenderToSceneAlpha(target pixel.Target, matrix pixel.Matrix, alpha float64) {
	alpha = max(0, min(1, alpha))
	m := matrix.Moved(r.animationOriginOffset.Scaled(resources.MapTileSize.Float()))
	for _, a := range r.animations {
		mask := colors.White.RGBA
		if a.ColorMask != nil {
			mask = *a.ColorMask
		}
		mask = colors.WithAlpha(mask, alpha)
		a.Animation.Sprite().DrawColorMask(target, m, mask)
	}
}

func (r *BasicEntityRenderer) RenderToLightMap(target pixel.Target, matrix pixel.Matrix) {
	r.RenderToLightMapAlpha(target, matrix, 1)
}

func (r *BasicEntityRenderer) RenderToLightMapAlpha(target pixel.Target, matrix pixel.Matrix, alpha float64) {
	for _, l := range r.lights {
		l.RenderAlpha(target, matrix.Moved(r.lightOriginOffset), alpha)
	}
}
