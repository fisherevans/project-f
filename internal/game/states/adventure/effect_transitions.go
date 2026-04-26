package adventure

import (
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

func init() {
	registerStepConverter("teleport_player", func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		e := NewTeleportPlayerEffect()
		if ref := mapStr(m, "to_reference"); ref != "" {
			e = e.WithToReference(ref)
		}
		if toEntity := mapStr(m, "to_entity"); toEntity != "" {
			e = e.WithToEntityId(toEntity)
		}
		if style := mapStr(m, "transition_style"); style != "" {
			e = e.WithTransitionStyle(style)
		}
		if interstitialSteps, ok := m["interstitial"].([]any); ok {
			var interstitialEffects []Effect
			for _, rawStep := range interstitialSteps {
				stepMap, ok := rawStep.(map[string]any)
				if !ok {
					continue
				}
				for k, v := range stepMap {
					interstitialEffects = append(interstitialEffects, convertStep(&StepNode{Kind: k, Params: v}, tc, sequences)...)
				}
			}
			e = e.WithInterstitialEffects(interstitialEffects)
		}
		return []Effect{e}
	})
	registerStepConverter("fade", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		m := resolveMap(step.Params, tc)
		e := NewFadeEffect(mapFloat(m, "duration", 0.33), mapInt(m, "transitions", 1))
		if from := mapStr(m, "from_color"); from != "" {
			e = e.WithFromColor(from)
		}
		if to := mapStr(m, "to_color"); to != "" {
			e = e.WithToColor(to)
		}
		if mapHas(m, "auto_deactivate") {
			e = e.WithAutoDeactivate(mapBool(m, "auto_deactivate"))
		}
		return []Effect{e}
	})
	registerStepConverter("deactivate_fade", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		return []Effect{NewDeactivateFadeEffect(resolveString(step.Params, tc))}
	})
}

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

func (e *EffectFade) Process(source EntityReader, s *State) bool {

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
	return true
}

type EffectDeactivateFade struct {
	instantEffect
	FadeId string
}

func (e *EffectDeactivateFade) Process(source EntityReader, s *State) bool {
	s.overlays.Deactivate(e.FadeId)
	return true
}
