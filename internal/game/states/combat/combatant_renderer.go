package combat

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
)

type CombatantRenderer struct {
	animation       *anim.AnimatedSprite
	transformations []CombatantRenderTransformation
	flip            bool
}

func NewCombatantRenderer(animation *anim.AnimatedSprite, flip bool) *CombatantRenderer {
	return &CombatantRenderer{
		animation:       animation,
		transformations: []CombatantRenderTransformation{},
		flip:            flip,
	}
}

func (r *CombatantRenderer) AddTransformation(transformation CombatantRenderTransformation) {
	r.transformations = append(r.transformations, transformation)
}

func (r *CombatantRenderer) AddFromConfig(ts []rpg.SkillTickCombatantTransformation) {
	for _, t := range ts {
		speed := 1.0
		if t.Speed != 0 {
			speed = t.Speed
		}
		var transform func() CombatantRenderTransformation
		switch t.Type {
		// todo repetitions + speed
		case rpg.SkillTickCombatantTransformationTypePounce:
			transform = func() CombatantRenderTransformation { return NewCombatantRenderTransformationPounce(speed) }
		case rpg.SkillTickCombatantTransformationTypeRecoil:
			transform = func() CombatantRenderTransformation { return NewCombatantRenderTransformationRecoil(speed) }
		case rpg.SkillTickCombatantTransformationTypeWiggle:
			transform = func() CombatantRenderTransformation { return NewCombatantRenderTransformationWiggle(speed) }
		case rpg.SkillTickCombatantTransformationTypeHop:
			transform = func() CombatantRenderTransformation { return NewCombatantRenderTransformationHop(speed) }
		}
		if transform == nil {
			continue
		}
		if t.Repetitions <= 1 {
			r.AddTransformation(transform())
		}
		var chain []CombatantRenderTransformation
		for i := 0; i < t.Repetitions; i++ {
			chain = append(chain, transform())
		}
		r.AddTransformation(newChainedTransformations(chain...))
	}
}

func (r *CombatantRenderer) Render(target pixel.Target, timeDelta float64, com Combatant) {
	colorMask := com.GetColorMask()
	if com.IsDead() {
		colorMask = colors.MixColor(colorMask, colors.HexString("#af8686"))
	} else {
		r.animation.Update(timeDelta)
	}

	var position pixel.Vec
	rotateDirection := 1.0
	if r.flip {
		position = pixel.V(math.Floor(game.GameWidth*0.8), math.Floor(game.GameHeight*0.4))
		rotateDirection = -1
	} else {
		position = pixel.V(math.Floor(game.GameWidth*0.2), math.Floor(game.GameHeight*0.4))
	}

	m := pixel.IM
	if com.IsDead() {
		h := r.animation.Sprite().Bounds().H()
		ry := -h / 3.0
		m = m.Rotated(pixel.V(0, ry), math.Pi/2.0*rotateDirection)
	} else {
		for id := 0; id < len(r.transformations); id++ {
			transformation := r.transformations[id]
			if transformation.Update(timeDelta) {
				r.transformations = append(r.transformations[:id], r.transformations[id+1:]...)
				id--
			}
			m = transformation.Apply(m, r.flip)
		}
	}
	m = m.Moved(position)

	r.animation.Sprite().DrawColorMask(target, m, colorMask)
}

type CombatantRenderTransformation interface {
	Update(timeDelta float64) bool
	Apply(matrix pixel.Matrix, flip bool) pixel.Matrix
}

type baseCombatantRenderEffect struct {
	duration float64
	elapsed  float64
}

func newBaseCombatantRenderEffect(duration float64) *baseCombatantRenderEffect {
	return &baseCombatantRenderEffect{
		duration: duration,
	}
}

func (e *baseCombatantRenderEffect) Update(timeDelta float64) bool {
	e.elapsed = min(e.elapsed+timeDelta, e.duration)
	return e.elapsed >= e.duration
}

func (e *baseCombatantRenderEffect) progress() float64 {
	return max(0, min(1, e.elapsed/e.duration))
}

type CombatantRenderTransformationBasic struct {
	*baseCombatantRenderEffect
	xKeys *interp.Keys
	yKeys *interp.Keys
}

func NewCombatantRenderTransformationPounce(speed float64) *CombatantRenderTransformationBasic {
	return &CombatantRenderTransformationBasic{
		baseCombatantRenderEffect: newBaseCombatantRenderEffect(1.0 / speed),
		xKeys: interp.NewKeys().WithDefaultFunction(interp.Smootherstep).
			WithKey(0, 0).
			WithKey(0.33, 20).
			WithKey(1, 0),
		yKeys: interp.NewKeys().
			WithKey(0, 0).
			WithKeyFn(0.18, 15, interp.ParabolaUpLeft).
			WithKeyFn(0.33, 0, interp.ParabolaUpRight).
			WithKey(1, 0),
	}
}

func NewCombatantRenderTransformationRecoil(speed float64) *CombatantRenderTransformationBasic {
	return &CombatantRenderTransformationBasic{
		baseCombatantRenderEffect: newBaseCombatantRenderEffect(1.0 / speed),
		xKeys: interp.NewKeys().WithDefaultFunction(interp.Smootherstep).
			WithKey(0, 0).
			WithKey(0.25, 0).
			WithKey(0.33, -10).
			WithKey(0.5, 0).
			WithKey(1, 0),
		yKeys: interp.NewKeys().
			WithKey(0, 0).
			WithKey(1, 0),
	}
}

func NewCombatantRenderTransformationHop(speed float64) *CombatantRenderTransformationBasic {
	return &CombatantRenderTransformationBasic{
		baseCombatantRenderEffect: newBaseCombatantRenderEffect(0.5 / speed),
		xKeys: interp.NewKeys().WithDefaultFunction(interp.Smootherstep).
			WithKey(0, 0).
			WithKey(1, 0),
		yKeys: interp.NewKeys().
			WithKey(0, 0).
			WithKeyFn(0.5, 10, interp.ParabolaUpLeft).
			WithKeyFn(1, 0, interp.ParabolaUpRight),
	}
}

func NewCombatantRenderTransformationWiggle(speed float64) *CombatantRenderTransformationBasic {
	return &CombatantRenderTransformationBasic{
		baseCombatantRenderEffect: newBaseCombatantRenderEffect(0.5 / speed),
		xKeys: interp.NewKeys().WithDefaultFunction(interp.Smootherstep).
			WithKey(0, 0).
			WithKey(0.25, -3).
			WithKey(0.75, 3).
			WithKey(1, 0),
		yKeys: interp.NewKeys().
			WithKey(0, 0).
			WithKey(1, 0),
	}
}

func (c *CombatantRenderTransformationBasic) Apply(matrix pixel.Matrix, flip bool) pixel.Matrix {
	x := c.xKeys.Interpolate(c.progress())
	if flip {
		x = -1.0 * x
	}
	y := c.yKeys.Interpolate(c.progress())
	delta := pixel.V(x, y)
	return matrix.Moved(delta)
}

type chainedTransformations struct {
	transformations []CombatantRenderTransformation
}

func newChainedTransformations(transformations ...CombatantRenderTransformation) *chainedTransformations {
	return &chainedTransformations{
		transformations: transformations,
	}
}

func (c *chainedTransformations) Update(timeDelta float64) bool {
	if len(c.transformations) > 0 {
		if c.transformations[0].Update(timeDelta) {
			c.transformations = c.transformations[1:]
		}
	}
	return len(c.transformations) == 0
}

func (c *chainedTransformations) Apply(matrix pixel.Matrix, flip bool) pixel.Matrix {
	if len(c.transformations) == 0 {
		return matrix
	}
	return c.transformations[0].Apply(matrix, flip)
}
