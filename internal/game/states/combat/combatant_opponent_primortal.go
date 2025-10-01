package combat

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
)

type PrimortalOpponent struct {
	*CurrentCombatantSkills
	Statuses        *AppliedStatuses
	Health          *HealthState
	Tempo           *Tempo
	NextSkill       *rpg.SkillId
	DamageFlashMask *DamageFlashMask

	name          string
	skillChooser  SkillChooser
	baseAnimation *anim.AnimatedSprite
}

func NewPrimortalOpponent(primortal rpg.PrimortalType) *PrimortalOpponent {
	p := rpg.Primortals[primortal]
	var skills []rpg.SkillId
	for _, us := range p.UnlockableSkills {
		skills = append(skills, us.SkillId)
	}
	for _, s := range p.AdditionalCombatSkills {
		skills = append(skills, s)
	}
	return &PrimortalOpponent{
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Statuses:               NewAppliedStatuses(),
		Health:                 NewHealthState(rand.Intn(25) + 50),
		Tempo:                  &Tempo{},
		DamageFlashMask:        NewDamageFlashMask(),
		name:                   p.Name,
		skillChooser:           NewRandomSkillChooserEven(skills),
		baseAnimation:          anim.Load(atlas, "primortals/"+string(primortal), "default"),
	}
}

func (o *PrimortalOpponent) GetStats() CombatantStats {
	return CombatantStats{
		Stance: o.GetCurrentSkill().GetCurrentStance(),
	}
}

var _ Opponent = &PrimortalOpponent{}

func (o *PrimortalOpponent) IsPlayer() bool {
	return false
}

func (o *PrimortalOpponent) Update(timeDelta float64) {
	o.DamageFlashMask.Update(timeDelta)
}

func (o *PrimortalOpponent) GetColorMask() pixel.RGBA {
	return o.DamageFlashMask.getMask()
}

func (o *PrimortalOpponent) ApplyDamage(damage int) {
	o.Health.AdjustTarget(-damage)
	o.DamageFlashMask.damaged()
}

func (o *PrimortalOpponent) GetStatuses() *AppliedStatuses {
	return o.Statuses
}

func (o *PrimortalOpponent) Name() string {
	return o.name
}

func (o *PrimortalOpponent) GetHealth() *HealthState {
	return o.Health
}

func (o *PrimortalOpponent) GetTotalMaxHealth() int {
	return o.Health.Max
}

func (o *PrimortalOpponent) IsDead() bool {
	return o.Health.Current < 1
}

func (o *PrimortalOpponent) GetTempo() *Tempo {
	return o.Tempo
}

func (o *PrimortalOpponent) PeekNextSkill() *rpg.SkillId {
	if o.NextSkill == nil {
		o.NextSkill = o.skillChooser.NextSkill()
	}
	return o.NextSkill
}

func (o *PrimortalOpponent) PopNextSkill() *rpg.SkillId {
	if o.NextSkill == nil {
		o.NextSkill = o.skillChooser.NextSkill()
	}
	popped := o.NextSkill
	o.NextSkill = o.skillChooser.NextSkill()
	return popped
}

func (o *PrimortalOpponent) IsNextSkillCommitted() bool {
	return true
}

func (o *PrimortalOpponent) GetAnimation() *anim.AnimatedSprite {
	return o.baseAnimation
}
