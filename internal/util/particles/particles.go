package particles

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type PositionFactory func() pixel.Vec

func RandomPositionFactory(originLocation gfx.OriginLocation, width, height float64) PositionFactory {
	return func() pixel.Vec {
		bottomLeft := originLocation.AlignFrom(gfx.BottomLeft, width, height)
		return pixel.V(bottomLeft.X+width*rand.Float64(), bottomLeft.Y+height*rand.Float64())
	}
}

type VelocityFactory func(pos pixel.Vec) pixel.Vec

func RandomVelocityFactory(minSpeed, maxSpeed float64) VelocityFactory {
	return func(pos pixel.Vec) pixel.Vec {
		return pixel.V(minSpeed+rand.Float64()*(maxSpeed-minSpeed), 0).Rotated(2 * math.Pi * rand.Float64())
	}
}

type AnimationFactory func() *anim.AnimatedSprite
type AgeFactory func() float64

func RandomAgeFactory(minAge, maxAge float64) AgeFactory {
	return func() float64 {
		return minAge + rand.Float64()*(maxAge-minAge)
	}
}

type NewParticleOpt func(p *particle)

func WithFading(pctIn, pctOut float64) NewParticleOpt {
	return func(p *particle) {
		total := pctIn + pctOut
		if pctIn < 0 || pctOut < 0 || total > 1 {
			log.Error().Msgf("Invalid fade percentages: %f, %f", pctIn, pctOut)
			pctIn = 0.5
			pctOut = 0.5
		}
		p.fadeInDuration = util.Ptr(pctIn * p.maxAge)
		p.fadeOutDuration = util.Ptr(pctOut * p.maxAge)
	}
}

func WithColorMask(colorMask pixel.RGBA) NewParticleOpt {
	return WithColorMaskSupplier(func() pixel.RGBA {
		return colorMask
	})
}

func WithColorMaskSupplier(sup func() pixel.RGBA) NewParticleOpt {
	return func(p *particle) {
		c := sup()
		p.colorMask = &c
	}
}

type particle struct {
	position  pixel.Vec
	velocity  pixel.Vec
	maxAge    float64
	age       float64
	animation *anim.AnimatedSprite

	colorMask       *pixel.RGBA
	fadeInDuration  *float64
	fadeOutDuration *float64
}

func (p *particle) Update(globalAcceleration pixel.Vec, timeDelta float64) {
	p.age += timeDelta
	p.velocity = p.velocity.Add(globalAcceleration.Scaled(timeDelta))
	p.position = p.position.Add(p.velocity.Scaled(timeDelta))
	if p.animation != nil {
		p.animation.Update(timeDelta)
	}
}

func (p *particle) Render(target pixel.Target, matrix pixel.Matrix) {
	if p.age >= p.maxAge || p.animation == nil {
		return
	}
	colorMask := colors.White.RGBA
	alpha := 1.0
	if p.colorMask != nil {
		colorMask = *p.colorMask
	}
	if p.fadeInDuration != nil && p.age < *p.fadeInDuration {
		alpha = interp.Smootherstep(p.age / *p.fadeInDuration)
	}
	if p.fadeOutDuration != nil && p.age > p.maxAge-*p.fadeOutDuration {
		alpha = interp.Smootherstep((p.maxAge - p.age) / *p.fadeOutDuration)
	}
	if alpha != 1 {
		colorMask = colors.WithAlpha(colorMask, alpha)
	}
	p.animation.Sprite().DrawColorMask(target, matrix.Moved(p.position), colorMask)
}

func (p *particle) IsDead() bool {
	return p.age >= p.maxAge
}

type Group struct {
	particles          []*particle
	positionFactory    PositionFactory
	velocityFactory    VelocityFactory
	animationFactory   AnimationFactory
	ageFactory         AgeFactory
	particleOpts       []NewParticleOpt
	globalAcceleration pixel.Vec
}

func NewGroup(positionFactory PositionFactory, velocityFactory VelocityFactory, animationFactory AnimationFactory, ageFactory AgeFactory, globalAcceleration pixel.Vec, particleOpts ...NewParticleOpt) *Group {
	return &Group{
		positionFactory:    positionFactory,
		velocityFactory:    velocityFactory,
		animationFactory:   animationFactory,
		ageFactory:         ageFactory,
		globalAcceleration: globalAcceleration,
		particleOpts:       particleOpts,
	}
}

func (g *Group) AddParticle() {
	pos := g.positionFactory()
	p := &particle{
		position:  pos,
		velocity:  g.velocityFactory(pos),
		maxAge:    g.ageFactory(),
		animation: g.animationFactory(),
	}
	for _, opt := range g.particleOpts {
		opt(p)
	}
	g.particles = append(g.particles, p)
}

func (g *Group) Render(target pixel.Target, matrix pixel.Matrix, timeDelta float64) {
	for i := 0; i < len(g.particles); i++ {
		p := g.particles[i]
		p.Update(g.globalAcceleration, timeDelta)
		if p.IsDead() {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
			i--
			continue
		}
		p.Render(target, matrix)
	}
}
