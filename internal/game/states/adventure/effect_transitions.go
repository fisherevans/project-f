package adventure

import (
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

type EffectFade struct {
	FadeId          string `auto_generate:"true"`
	DurationSeconds float64
	Transitions     int // number of transitions (1 = fade out, 2 = fade out+in, etc.)
	AutoDeactivate  *bool
	FromColor       *string
	ToColor         *string
}

func (e *EffectFade) CompletionID() string {
	if e.FadeId == "" {
		return ""
	}
	return "fade:" + e.FadeId
}

func (e *EffectFade) WithColors(from, to string) *EffectFade {
	e.FromColor = &from
	e.ToColor = &to
	return e
}

func (e *EffectFade) Process(source EntityContext, s *State) bool {

	// Determine colors
	fromColor := pixel.RGBA{R: 0, G: 0, B: 0, A: 0} // transparent
	toColor := pixel.RGBA{R: 0, G: 0, B: 0, A: 1}   // black

	if e.FromColor != nil {
		fromColor = colors.HexString(*e.FromColor)
	}
	if e.ToColor != nil {
		toColor = colors.HexString(*e.ToColor)
	}

	transitions := e.Transitions
	if transitions < 1 {
		transitions = 1
	}

	autoDeactivate := true
	if e.AutoDeactivate != nil {
		autoDeactivate = *e.AutoDeactivate
	}
	// Create base overlay with auto-complete
	base := NewBaseOverlay(e.FadeId, e.CompletionID(), e.DurationSeconds, autoDeactivate)

	// Create fade overlay
	fade := NewFadeOverlay(fromColor, toColor, transitions, base)

	s.overlays.Add(fade)
	logEffectInfof(source, e, "fade overlay added")
	return true
}

type EffectDeactivateFade struct {
	instantEffect
	FadeId string
}

func (e *EffectDeactivateFade) Process(source EntityContext, s *State) bool {
	s.overlays.Deactivate(e.FadeId)
	logEffectInfof(source, e, "fade deactivated")
	return true
}
