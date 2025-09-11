package combat

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/rpg"
)

type Opponent interface {
	Combatant
	GetHealth() *HealthState
}

type Wall struct {
	Health    *HealthState
	Tempo     *Tempo
	NextSkill *rpg.SkillId
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

func (w *Wall) PeekNextSkill() *rpg.SkillId {
	if w.NextSkill == nil {
		w.NextSkill = randomSkill()
	}
	return w.NextSkill
}

func (w *Wall) PopNextSkill() *rpg.SkillId {
	if w.NextSkill == nil {
		w.NextSkill = randomSkill()
	}
	popped := w.NextSkill
	w.NextSkill = randomSkill()
	return popped
}

func randomSkill() *rpg.SkillId {
	switch rand.Intn(3) {
	case 0:
		return &rpg.Skill_Block.Id
	case 1:
		return &rpg.Skill_Shunt.Id
	default:
		return &rpg.Skill_Crush.Id
	}
}
