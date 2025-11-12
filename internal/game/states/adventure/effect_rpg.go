package adventure

import "fisherevans.com/project/f/internal/game/rpg"

type EffectYieldElythium struct {
	instantEffect
	Amount int
}

func (e *EffectYieldElythium) Process(source EntityReader, s *State) bool {
	newValue := max(s.globals.Get(rpg.GlobalKeyElythium).AsInt(0)+e.Amount, 0)
	s.globals.Set(rpg.GlobalKeyElythium, newValue)
	return true
}
