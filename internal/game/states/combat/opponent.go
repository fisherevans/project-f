package combat

import (
	"fisherevans.com/project/f/internal/game/rpg"
)

type Opponent interface {
	Combatant
	GetHealth() *HealthState
}

type Wall struct {
	Health *HealthState
	Tempo  *Tempo
}

func (w *Wall) GetStats() CombatantStats {
	return CombatantStats{
		BodyType:        rpg.BodyTypeRock,
		PhysicalAttack:  10,
		PhysicalDefense: 10,
		AetherAttack:    10,
		AetherDefense:   10,
	}
}

var _ Opponent = &Wall{}
var _ Combatant = &Wall{}

func (w *Wall) ApplyDamage(result rpg.DamageResult) {
	w.Health.AdjustTarget(-result.TotalDamage)
}

func (w *Wall) Name() string {
	return "Squishy Wall"
}

func (w *Wall) GetHealth() *HealthState {
	return w.Health
}

func (w *Wall) GetTempo() *Tempo {
	return w.Tempo
}
