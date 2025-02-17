package combat

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
)

type Opponent interface {
	Combatant
	GetHealth() HealthState
}

type Wall struct {
	MaxHealth     int
	CurrentHealth int
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

func (w *Wall) ApplyDamage(result rpg.DamageResult) DamageOutcome {
	w.CurrentHealth = util.MaxInt(0, w.CurrentHealth-result.TotalDamage)
	return DamageOutcome{
		Perished: w.CurrentHealth == 0,
	}
}

var _ Opponent = &Wall{}

func (w *Wall) Name() string {
	return "Squishy Wall"
}

func (w *Wall) GetHealth() HealthState {
	return HealthState{
		Max:     w.MaxHealth,
		Current: w.CurrentHealth,
	}
}
