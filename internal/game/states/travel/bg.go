package travel

import (
	"fisherevans.com/project/f/internal/util/rng"
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
)

var (
	atlas = resources.DefaultAtlas()
	minY  = -5.0
	maxY  = float64(game.GameHeight + 5)
)

type bgSprite struct {
	sprite      pixelutil.BoundedDrawable
	position    pixel.Vec
	speed       float64
	mask        pixel.RGBA
	baseAlpha   float64
	flashAmount float64
	flashSpeed  float64

	age float64
}

func newStarBgSprite() *bgSprite {
	x := rng.Float64() * (game.GameWidth)
	y := rng.Float64()*(maxY-minY) + minY
	maskHue := float64(rng.Intn(325-188)+188) / 360.0
	return &bgSprite{
		position: pixel.V(x, y),
		// randoms
		mask:        colors.HSLToRGBA(maskHue, 1, 1),
		sprite:      atlas.GetTilesheetSprite("adventure/hud/elythium_sparkle", rng.Intn(5)+1, 1),
		speed:       0.5 + 10*rng.Float64(),
		baseAlpha:   0.3 + rng.Float64()*0.7,
		flashAmount: rng.Float64(),
		flashSpeed:  0.5 + 4*rng.Float64(),
	}
}

func (s *bgSprite) Update(timeDelta, speedTimeDelta float64) {
	s.position = s.position.Add(pixel.V(0, -s.speed*speedTimeDelta))
	if s.position.Y < minY {
		s.position.Y = maxY
	}
	s.age += timeDelta
}

func (s *bgSprite) Render(target pixel.Target, matrix pixel.Matrix) {
	alpha := s.baseAlpha - (math.Sin(s.age*s.flashSpeed)+1.0)/2.0*s.flashAmount
	s.sprite.DrawColorMask(target, matrix, colors.LayerAlpha(s.mask, alpha))
}
