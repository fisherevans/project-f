package adventure

import (
	"math"
	"math/rand"

	"github.com/gopxl/pixel/v2"
)

var light = atlas.GetSprite("lights/amber_linear")

type LightRenderDetails struct {
	SizeScale     float64
	ColorMask     pixel.RGBA
	PositionDelta pixel.Vec
}

type LightModifier interface {
	Update(timeDelta float64)
	Apply(*LightRenderDetails)
}

type LightModifierJitterUpdate struct {
	FrequencySeconds   float64
	FrequencyVariation float64

	secondsLeft float64
}

func (f *LightModifierJitterUpdate) Update(timeDelta float64) bool {
	f.secondsLeft -= timeDelta
	if f.secondsLeft <= 0 {
		f.secondsLeft = f.FrequencyVariation/0.5 - rand.Float64()*f.FrequencyVariation + f.FrequencySeconds
		return true
	}
	return false
}

type LightModifierFlicker struct {
	Jitter              *LightModifierJitterUpdate
	SizeVariation       float64
	BrightnessVariation float64

	// todo parameterize
	sizeMultiplier       float64
	brightnessMultiplier float64
}

func (l *LightModifierFlicker) Update(timeDelta float64) {
	if l.Jitter.Update(timeDelta) {
		l.sizeMultiplier = 1.0 - rand.Float64()*l.SizeVariation
		l.brightnessMultiplier = 1.0 - rand.Float64()*l.BrightnessVariation
	}
}

func (l *LightModifierFlicker) Apply(details *LightRenderDetails) {
	details.SizeScale *= l.sizeMultiplier
	details.ColorMask = details.ColorMask.Scaled(l.brightnessMultiplier)
}

type LightModifierPulse struct {
	PeriodSeconds       float64
	SizeIntensity       float64
	BrightnessIntensity float64

	elapsedSeconds       float64
	sizeMultiplier       float64
	brightnessMultiplier float64
}

func (l *LightModifierPulse) Update(timeDelta float64) {
	l.elapsedSeconds = math.Remainder(l.elapsedSeconds+timeDelta, math.Pi)
	scale := (math.Sin(l.elapsedSeconds*l.PeriodSeconds*math.Pi*2) + 1.0) / 2.0
	l.sizeMultiplier = 1.0 - scale*l.SizeIntensity
	l.brightnessMultiplier = 1.0 - scale*l.BrightnessIntensity
}

func (l *LightModifierPulse) Apply(details *LightRenderDetails) {
	details.SizeScale *= l.sizeMultiplier
	details.ColorMask = details.ColorMask.Scaled(l.brightnessMultiplier)
}

type Light struct {
	RenderDetails LightRenderDetails
	Modifiers     []LightModifier
}

func (l *Light) Update(timeDelta float64) {
	for _, m := range l.Modifiers {
		m.Update(timeDelta)
	}
}

func (l *Light) Render(target pixel.Target, matrix pixel.Matrix) {
	renderDetails := l.RenderDetails
	for _, m := range l.Modifiers {
		m.Apply(&renderDetails)
	}
	light.DrawColorMask(
		target,
		pixel.IM.
			Moved(renderDetails.PositionDelta).
			Scaled(pixel.ZV, renderDetails.SizeScale).
			Chained(matrix),
		renderDetails.ColorMask)
}
