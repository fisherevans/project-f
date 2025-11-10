package adventure

import "fisherevans.com/project/f/internal/game/rpg"

type EffectYieldElythium struct {
	instantEffect
	Amount int
}

func (e *EffectYieldElythium) Process(source EntityContext, s *State) bool {
	newValue := max(s.RunState().Get(rpg.RunStateKeyElythium).AsInt(0)+e.Amount, 0)
	s.runState.Set(rpg.RunStateKeyElythium, newValue)
	logEffectInfof(source, e, "elythium updated to: %d", newValue)
	return true
}
