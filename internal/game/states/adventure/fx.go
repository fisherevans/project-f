package adventure

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func (s *State) addBackgroundFx(f fx) {
	s.backgroundFxs = append(s.backgroundFxs, f)
}

func (s *State) addForegroundFx(f fx) {
	s.foregroundFxs = append(s.foregroundFxs, f)
}

func renderFx(target pixel.Target, cameraDelta pixel.Vec, timeDelta float64, fxs []fx) {
	game.DebugTLf("rendering %d fxs", len(fxs))
	for idx := 0; idx < len(fxs); idx++ {
		fxs[idx].Update(timeDelta)
		if fxs[idx].IsFinished() {
			fxs = append(fxs[:idx], fxs[idx+1:]...)
			idx--
		} else {
			moveDelta := fxs[idx].GetPreciseMapLocation().Scaled(resources.MapTileSize.Float())
			// todo don't render things off screen?A
			renderMatrix := pixel.IM.Moved(cameraDelta).Moved(moveDelta)
			fxs[idx].Render(target, renderMatrix)
		}
	}
}

type fx interface {
	Update(timeDelta float64)
	IsFinished() bool
	Render(target pixel.Target, matrix pixel.Matrix)
	GetPreciseMapLocation() pixel.Vec
}

type timedFx struct {
	preciseMapLocation pixel.Vec
	age                float64
	maxAge             float64
}

func (fx *timedFx) IsFinished() bool {
	return fx.age >= fx.maxAge
}

func (fx *timedFx) Update(timeDelta float64) {
	fx.age += timeDelta
}

func (fx *timedFx) GetPreciseMapLocation() pixel.Vec {
	return fx.preciseMapLocation
}

func (fx *timedFx) ageProgression() float64 {
	return min(fx.age/fx.maxAge, 1.0)
}

type fadingSpriteFx struct {
	*timedFx
	fromMask, toMask pixel.RGBA
	sprite           pixelutil.BoundedDrawable
}

func newFadingSpriteFx(location pixel.Vec, sprite pixelutil.BoundedDrawable, startMask pixel.RGBA, maxAge float64) *fadingSpriteFx {
	toMask := colors.LayerAlpha(startMask, 0)
	if sprite == nil {
		log.Fatal().Msgf("Tried to create fading sprite fx with nil sprite")
	}
	return &fadingSpriteFx{
		timedFx: &timedFx{
			preciseMapLocation: location,
			maxAge:             maxAge,
		},
		sprite:   sprite,
		fromMask: startMask,
		toMask:   toMask,
	}
}

func (fx *fadingSpriteFx) Render(target pixel.Target, matrix pixel.Matrix) {
	mask := colors.Lerp(fx.fromMask, fx.toMask, fx.ageProgression())
	fx.sprite.DrawColorMask(target, matrix, mask)
}

type starFx struct {
	// constructor
	preciseMapLocation pixel.Vec
	startX             float64
	endX               float64

	// randomized
	mask        pixel.RGBA
	sprite      pixelutil.BoundedDrawable
	speed       float64
	baseAlpha   float64
	flashAmount float64
	flashSpeed  float64

	// set with defauls
	age float64
}

func newStarFx(location pixel.Vec, startX, endX float64) *starFx {
	maskHue := float64(rand.Intn(325-188)+188) / 360.0
	return &starFx{
		preciseMapLocation: location,
		startX:             startX,
		endX:               endX,
		// randoms
		mask:        colors.HSLToRGBA(maskHue, 1, 1),
		sprite:      atlas.GetTilesheetSprite("adventure/hud/elythium_sparkle", rand.Intn(5)+1, 1),
		speed:       0.5 + 2*rand.Float64(),
		baseAlpha:   0.25 + rand.Float64()*0.75,
		flashAmount: rand.Float64(),
		flashSpeed:  4 + 16*rand.Float64(),
	}
}

func (fx *starFx) IsFinished() bool {
	return false
}

func (fx *starFx) GetPreciseMapLocation() pixel.Vec {
	return fx.preciseMapLocation
}

func (fx *starFx) Update(timeDelta float64) {
	fx.preciseMapLocation.X += fx.speed * timeDelta
	if fx.preciseMapLocation.X >= fx.endX {
		fx.preciseMapLocation.X -= fx.endX - fx.startX
	}
	fx.age += timeDelta
}

func (fx *starFx) Render(target pixel.Target, matrix pixel.Matrix) {
	alpha := fx.baseAlpha - (math.Sin(fx.age*fx.flashAmount)+1.0)/2.0*fx.flashAmount
	fx.sprite.DrawColorMask(target, matrix, colors.LayerAlpha(fx.mask, alpha))
}
