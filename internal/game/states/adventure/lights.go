package adventure

import (
	"math"
	"math/rand"

	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type Light struct {
	RenderDetails []LightRenderDetails
	Modifiers     []LightModifier
}

func (l *Light) Update(timeDelta float64) {
	if l == nil {
		return
	}
	for _, m := range l.Modifiers {
		m.Update(timeDelta)
	}
}

func (l *Light) Render(target pixel.Target, matrix pixel.Matrix) {
	l.RenderAlpha(target, matrix, 1)
}

func (l *Light) RenderAlpha(target pixel.Target, matrix pixel.Matrix, alpha float64) {
	if l == nil {
		return
	}
	for _, renderDetails := range l.RenderDetails {
		for _, m := range l.Modifiers {
			m.Apply(&renderDetails)
		}
		mask := colors.WithAlpha(renderDetails.ColorMask, alpha)
		renderDetails.Sprite.DrawColorMask(
			target,
			pixel.IM.
				Moved(renderDetails.PositionDelta).
				Scaled(pixel.ZV, renderDetails.SizeScale).
				Chained(matrix),
			mask)

	}
}

func NewLight(color pixel.RGBA, size float64) *Light {
	return NewLightWithModifier(color, size, "")
}

func NewLightWithModifier(color pixel.RGBA, size float64, modifier string) *Light {
	l := &Light{
		RenderDetails: []LightRenderDetails{
			{
				Sprite:    atlas.GetSprite("lights/white_5x5"),
				SizeScale: size,
				ColorMask: color,
			},
			//{
			//	Sprite:    atlas.GetSprite("lights/white_hard_5x5"),
			//	SizeScale: size,
			//	ColorMask: colors.WithAlpha(color, 0.5),
			//},
			//{
			//	Sprite:    atlas.GetSprite("lights/white_hard_5x5"),
			//	SizeScale: size * 0.5,
			//	ColorMask: colors.WithAlpha(color, 0.5),
			//},
		},
	}
	pulse := func(periodSeconds float64) LightModifier {
		return &LightModifierPulse{
			PeriodSeconds:       periodSeconds,
			SizeIntensity:       0.1,
			BrightnessIntensity: 0.4,
		}
	}
	flicker := func(freqSeconds float64) LightModifier {
		return &LightModifierFlicker{
			Jitter: &LightModifierJitterUpdate{
				FrequencySeconds:   freqSeconds,
				FrequencyVariation: freqSeconds * 0.333,
			},
			SizeVariation:       0.1,
			BrightnessVariation: 0.1,
		}
	}
	switch modifier {
	case "":
	case "pulse_slow":
		l.Modifiers = append(l.Modifiers, pulse(6))
	case "pulse_medium", "pulse":
		l.Modifiers = append(l.Modifiers, pulse(4))
	case "pulse_fast":
		l.Modifiers = append(l.Modifiers, pulse(2))
	case "flicker_slow":
		l.Modifiers = append(l.Modifiers, flicker(0.35))
	case "flicker_medium", "flicker":
		l.Modifiers = append(l.Modifiers, flicker(0.15))
	case "flicker_fast":
		l.Modifiers = append(l.Modifiers, flicker(0.075))
	default:
		log.Error().Str("modifier", modifier).Msg("unknown light modifier")
	}
	return l
}

type LightRenderDetails struct {
	Sprite        pixelutil.BoundedDrawable
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
	// Advance time and wrap by the period to avoid float growth (no discontinuity in phase).
	l.elapsedSeconds += timeDelta
	l.elapsedSeconds = math.Mod(l.elapsedSeconds, l.PeriodSeconds)
	if l.elapsedSeconds < 0 { // handle negative dt just in case
		l.elapsedSeconds += l.PeriodSeconds
	}

	// Convert to phase [0,1), then to angle.
	phase := l.elapsedSeconds / l.PeriodSeconds      // 0..1
	scale := 0.5 * (math.Sin(2*math.Pi*phase) + 1.0) // 0..1

	l.sizeMultiplier = 1.0 - scale*l.SizeIntensity
	l.brightnessMultiplier = 1.0 - scale*l.BrightnessIntensity

}

func (l *LightModifierPulse) Apply(details *LightRenderDetails) {
	details.SizeScale *= l.sizeMultiplier
	details.ColorMask = details.ColorMask.Scaled(l.brightnessMultiplier)
}
