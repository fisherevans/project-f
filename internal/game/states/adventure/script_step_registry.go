package adventure

import "github.com/rs/zerolog/log"

type StepConverter func(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect

var stepConverters = map[string]StepConverter{}

func registerStepConverter(kind string, converter StepConverter) {
	if _, exists := stepConverters[kind]; exists {
		log.Fatal().Str("kind", kind).Msg("duplicate step converter registration")
	}
	stepConverters[kind] = converter
}

func convertStep(step *StepNode, tc *TemplateContext, sequences map[string]*SequenceDef) []Effect {
	if converter, ok := stepConverters[step.Kind]; ok {
		return converter(step, tc, sequences)
	}
	log.Error().Str("kind", step.Kind).Msg("unknown step kind (should have been caught by validation)")
	return nil
}
