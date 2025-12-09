package combat

import (
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
)

type PrimortalOpponent struct {
	*CurrentCombatantSkills
	Statuses    *AppliedStatuses
	Health      *HealthState
	NextSkill   *rpg.SkillId
	HealthFlash *HealthFlash
	Renderer    *CombatantRenderer

	name         string
	skillChooser SkillChooser
}

func NewPrimortalOpponent(primortal rpg.PrimortalType) *PrimortalOpponent {
	p := rpg.Primortals[primortal]
	archetype, exists := primortal.Primortal().CombatArchetypes["default"]
	if !exists {
		panic("no default archetype for primortal: " + primortal)
	}
	syncVariance := archetype.AdditionalSyncVariance
	if syncVariance > 0 {
		syncVariance = rand.Intn(syncVariance)
	}
	return &PrimortalOpponent{
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Statuses:               NewAppliedStatuses(),
		Health:                 NewHealthState(p.BaseSync + archetype.AdditionalSync),
		HealthFlash:            NewDamageFlashMask(),
		name:                   p.Name,
		skillChooser:           NewSkillChooser(archetype.SkillPool),
		Renderer:               NewCombatantRenderer(anim.LoadTilesheetAnimation(atlas, "primortals/"+string(primortal), "default"), true),
	}
}

func (o *PrimortalOpponent) GetStats() rpg.CombatantStats {
	return rpg.CombatantStats{
		TempoLevel:   rpg.TempoLevel0,
		Stance:       o.GetCurrentSkill().GetCurrentStance(),
		StatusLevels: o.Statuses.GetLevels(),
	}
}

var _ Opponent = &PrimortalOpponent{}

func (o *PrimortalOpponent) IsPlayer() bool {
	return false
}

func (o *PrimortalOpponent) Update(timeDelta float64) {
	o.HealthFlash.Update(timeDelta)
}

func (o *PrimortalOpponent) GetColorMask() pixel.RGBA {
	return o.HealthFlash.getMask()
}

func (o *PrimortalOpponent) AdjustHealth(amount int) {
	o.Health.AdjustTarget(amount)
	o.HealthFlash.basedOnAdjust(amount)
}

func (o *PrimortalOpponent) GetRenderer() *CombatantRenderer {
	return o.Renderer
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
