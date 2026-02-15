package travel

import (
	"math"
	"math/rand"

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
	x := rand.Float64() * (game.GameWidth)
	y := rand.Float64()*(maxY-minY) + minY
	maskHue := float64(rand.Intn(325-188)+188) / 360.0
	return &bgSprite{
		position: pixel.V(x, y),
		// randoms
		mask:        colors.HSLToRGBA(maskHue, 1, 1),
		sprite:      atlas.GetTilesheetSprite("adventure/hud/elythium_sparkle", rand.Intn(5)+1, 1),
		speed:       0.5 + 10*rand.Float64(),
		baseAlpha:   0.3 + rand.Float64()*0.7,
		flashAmount: rand.Float64(),
		flashSpeed:  0.5 + 4*rand.Float64(),
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
