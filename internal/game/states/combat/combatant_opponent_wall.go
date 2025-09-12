package combat

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/rpg"
)

type Wall struct {
	*CurrentCombatantSkills
	Health          *HealthState
	Tempo           *Tempo
	NextSkill       *rpg.SkillId
	DamageFlashMask *DamageFlashMask
}

func NewWall() *Wall {
	return &Wall{
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Health:                 NewHealthState(100),
		Tempo:                  &Tempo{},
		DamageFlashMask:        NewDamageFlashMask(),
	}
}

func (w *Wall) GetStats() CombatantStats {
	return CombatantStats{
		BodyType:        rpg.BodyTypeRock,
		PhysicalAttack:  10,
		PhysicalDefense: 10,
		AetherAttack:    10,
		AetherDefense:   10,
		Stance:          w.GetCurrentSkill().GetCurrentStance(),
	}
}

var _ Opponent = &Wall{}

func (w *Wall) Update(timeDelta float64) {
	w.DamageFlashMask.Update(timeDelta)
}

func (w *Wall) GetColorMask() pixel.RGBA {
	return w.DamageFlashMask.getMask()
}

func (w *Wall) ApplyDamage(result rpg.DamageResult) {
	w.Health.AdjustTarget(-result.TotalDamage)
	w.DamageFlashMask.damaged()
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
