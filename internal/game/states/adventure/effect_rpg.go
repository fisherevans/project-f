package adventure

import "fisherevans.com/project/f/internal/game/rpg"

func init() {
	registerStepConverter("yield_elythium", func(step *StepNode, tc *TemplateContext, _ map[string]*SequenceDef) []Effect {
		amount := 0
		switch v := step.Params.(type) {
		case int:
			amount = v
		case float64:
			amount = int(v)
		default:
			m := resolveMap(step.Params, tc)
			amount = mapInt(m, "amount", 0)
		}
		return []Effect{NewYieldElythiumEffect(amount)}
	})
}

type EffectYieldElythium struct {
	instantEffect
	Amount int
}

func (e *EffectYieldElythium) Process(source EntityReader, s *State) bool {
	newValue := max(s.globals.Get(rpg.GlobalKeyElythium).AsInt(0)+e.Amount, 0)
	s.globals.Set(rpg.GlobalKeyElythium, newValue)
	return true
}
