package startup

import (
	"cmp"
	"math"
	"slices"

	"fisherevans.com/project/f/internal/game/audio"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/interp"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/util/colors"
)

type letter struct {
	sprite                  pixelutil.BoundedDrawable
	dx                      int
	timeOffset              float64
	duration                float64
	animationMatrixFunction func(float64) pixel.Matrix
	colorFunction           func(float64) pixel.RGBA
}

type State struct {
	game.BaseState
	elapsed     float64
	initialized bool

	letters  []letter
	thanadox pixelutil.BoundedDrawable

	spriteBatch  *pixel.Batch
	spriteCanvas *opengl.Canvas

	rayBatch  *pixel.Batch
	rayCanvas *opengl.Canvas

	blendCanvas *shaders.Canvas
	control     *audio.PlaybackControl
}

const baseOffset = 0.15
const normalAnimationDuration = 1.5

func NewDevice(_ game.StartupDeviceIntent) game.State {
	s := &State{
		spriteBatch:  atlas.NewBatch(),
		spriteCanvas: opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),
		rayBatch:     atlas.NewBatch(),
		rayCanvas:    opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight)),
		blendCanvas:  shaders.NewCanvas(game.GameWidth, game.GameHeight),
		thanadox:     atlas.GetSprite("startup/device_thanadox"),
		letters: []letter{
			{ // n
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 1, 1),
				dx:                      0,
				timeOffset:              baseOffset * 7,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(true),
				colorFunction:           rainbowColor,
			},
			{ // a
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 2, 1),
				dx:                      29 - 8,
				timeOffset:              baseOffset * 5,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(true),
				colorFunction:           rainbowColor,
			},
			{ // n
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 3, 1),
				dx:                      53 - 4,
				timeOffset:              baseOffset * 3,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(true),
				colorFunction:           rainbowColor,
			},
			{ // O
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 4, 1),
				dx:                      77 - 9,
				timeOffset:              baseOffset * 1,
				duration:                normalAnimationDuration * 0.85,
				animationMatrixFunction: oAnimation,
				colorFunction:           rainbowColor,
			},
			{ // d
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 5, 1),
				dx:                      88,
				timeOffset:              baseOffset * 2,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(false),
				colorFunction:           rainbowColor,
			},
			{ // e
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 6, 1),
				dx:                      122 - 9,
				timeOffset:              baseOffset * 4,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(false),
				colorFunction:           rainbowColor,
			},
			{ // c
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 7, 1),
				dx:                      142 - 9,
				timeOffset:              baseOffset * 6,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(false),
				colorFunction:           rainbowColor,
			},
			{ // k
				sprite:                  atlas.GetTilesheetSprite("startup/device_letters", 8, 1),
				dx:                      161 - 8,
				timeOffset:              baseOffset * 8,
				duration:                normalAnimationDuration,
				animationMatrixFunction: rotateIn(false),
				colorFunction:           rainbowColor,
			},
		},
	}
	slices.SortFunc(s.letters, func(a, b letter) int {
		return cmp.Compare(a.timeOffset, b.timeOffset)
	})
	s.blendCanvas.SetMaskBrightenShader(s.rayCanvas.Texture())
	return s
}

const initializeDeviceAfter = 0.3

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.elapsed += timeDelta
	if !s.initialized && s.elapsed > initializeDeviceAfter {
		s.initialized = true
		s.control = audio.GetSystem().PlaySFX("startup/device", 1)
	}

	if game.Controls[*State]().ButtonB().JustPressed() {
		game.SetActiveStateIntent(game.StartupDeviceIntent{})
		s.control.Stop()
	}
	if s.elapsed > baseOffset*35 || game.Controls[*State]().ButtonA().JustPressed() {
		game.SetActiveStateIntent(game.StartupCopyrightsIntent{})
		s.control.Stop()
	}

	// generate rays

	s.rayBatch.Clear()
	drawRays(s.elapsed, s.rayBatch)
	s.rayCanvas.Clear(colors.Black.RGBA)
	s.rayBatch.Draw(s.rayCanvas)

	// sprites

	s.spriteBatch.Clear()
	tAlpha := 0.0
	tStart, tEnd := baseOffset*2, baseOffset*8
	if s.elapsed > tStart {
		p := (s.elapsed - tStart) / (tEnd - tStart)
		tAlpha = interp.Smoothstep(p)
	}
	tMask := colors.FromString("#af0068")
	tMask = colors.WithAlpha(tMask, tAlpha)
	s.thanadox.DrawColorMask(s.spriteBatch, gfx.Moved(73, 35).Moved(gfx.BottomLeft.Align(s.thanadox)), tMask)

	x, y := 26+20, 82+16 // botom left of N, accounting for centering of sprite
	for _, l := range s.letters {
		progress := max(min((s.elapsed-l.timeOffset)/l.duration, 1.0), 0.0)
		m := gfx.Moved(x+l.dx, y)
		m = l.animationMatrixFunction(progress).Chained(m)
		mask := l.colorFunction(progress)
		l.sprite.DrawColorMask(s.spriteBatch, m, mask)
	}

	s.spriteCanvas.Clear(pixel.RGBA{})
	s.spriteBatch.Draw(s.spriteCanvas)

	// clear background on target

	bgProgress := interp.Smoothstep(min(s.elapsed/(baseOffset*6), 1.0))
	bgFrom := colors.FromString("#444")
	bgTo := colors.FromString("#eee")
	bg := colors.Lerp(bgFrom, bgTo, bgProgress)
	target.Clear(bg)

	// draw sprites to target

	s.blendCanvas.Clear(pixel.RGBA{})
	s.spriteCanvas.Draw(s.blendCanvas, centerMatrix)
	s.blendCanvas.Draw(target, centerMatrix) // shader setup references sprite texture
}

type ray struct {
	w, p, a float64
}

var rays []ray
var raysWidth float64

func init() {
	start := []ray{
		{w: 1, p: 6, a: 0.5},
		{w: 1, p: 6, a: 0.6},
		{w: 2, p: 7, a: 0.7},
		{w: 4, p: 3, a: 0.8},
		{w: 6, p: 3, a: 0.8},
		{w: 10, p: 3, a: 0.9},
	}
	middle := ray{w: 20, p: 2, a: 1.0}
	end := make([]ray, len(start))
	copy(end, start)
	slices.Reverse(end)
	rays = append(start, middle)
	rays = append(start, end...)
	raysWidth = 0.0
	for _, r := range rays {
		raysWidth += r.p*2 + r.w
	}
}

const rayMaxEffect = 0.33
const rayAngle = 60 * math.Pi / 180
const rayPanPct = 0.175
const rayDx = -55

func drawRays(elapsed float64, target pixel.Target) {
	raysFrom, raysTo := baseOffset*12, baseOffset*30
	p := max(min((elapsed-raysFrom)/(raysTo-raysFrom), 1.0), 0.0)
	x := float64(game.GameWidth)*p*rayPanPct - raysWidth/2 + (1-rayPanPct)/2.0*float64(game.GameWidth) + rayDx
	y := float64(game.GameHeight / 2)
	easeInOut := 0.5 * (1 - math.Cos(2*math.Pi*p)) // 0 > 1 > 0
	widthScale := 1.0 + p*3
	paddingScale := 1.0 + p

	angleRange := rayAngle * 0.5
	angleDelta := (angleRange * p) - angleRange/2.0
	angle := rayAngle + angleDelta

	for _, r := range rays {
		pad, renderWidth := r.p*paddingScale, r.w*widthScale
		x += pad
		m := pixel.IM.ScaledXY(pixel.ZV, pixel.V(game.GameWidth, renderWidth)).
			Rotated(pixel.ZV, angle).
			Moved(pixel.V(x, y))
		effect := rayMaxEffect * r.a
		amount := effect * easeInOut
		color := pixel.RGBA{
			R: amount,
			G: amount,
			B: amount,
			A: amount,
		}
		atlas.GetSprite("1x1").DrawColorMask(target, m, color)
		x += r.w + pad
	}
}

const rainbowFadeInPeriod = 0.5
const rainbowTargetHue = 0.65
const rainbowHueSwing = 0.175

func rainbowColor(p float64) pixel.RGBA {
	alpha := 1.0
	if p < rainbowFadeInPeriod {
		alpha = interp.Smootherstep(p / rainbowFadeInPeriod)
	}
	hue := math.Mod(rainbowTargetHue+math.Sin(-(1.0-p)*math.Pi*2)*rainbowHueSwing, 1.0)
	return colors.WithAlpha(colors.HSLToRGBA(hue, 1.0, 0.5), alpha)
}

const rotateInScaleFrom = 3.5
const rotateInScalePeriod = 0.75
const rotateInArcPeriod = 0.9
const rotateInArcRadius = 100.0
const rotateInArcAmount = 0.4

func rotateIn(fromTop bool) func(p float64) pixel.Matrix {
	return func(p float64) pixel.Matrix {
		p = interp.Smootherstep(p)
		// Scale animation
		scale := 1.0
		if p < rotateInScalePeriod {
			scale = rotateInScaleFrom - (p / rotateInScalePeriod * (rotateInScaleFrom - 1.0))
		}

		// Arc motion: travel along a circular arc and end at (0, 0)
		dx, dy := 0.0, 0.0
		if p < rotateInArcPeriod {
			progress := p / rotateInArcPeriod
			// Start at the beginning of the arc, end at (0,0)
			// angle goes from rotateInArcAmount * 2π down to 0
			angleProgress := -(1 - progress) * rotateInArcAmount * 2 * math.Pi
			size := rotateInArcRadius
			if fromTop {
				angleProgress += math.Pi
			}

			// Position along the arc
			dx = math.Cos(angleProgress) * size * (1 - progress)
			dy = math.Sin(angleProgress) * size * (1 - progress)
		}

		return pixel.IM.Scaled(pixel.ZV, scale).Moved(pixel.V(dx, dy))
	}
}

func oAnimation(p float64) pixel.Matrix {
	scale := p + math.Sin(math.Pi*p)*1.0
	return pixel.IM.Scaled(pixel.ZV, scale)
}
