package combat

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
)

type PlayerHealth struct {
	Sync   HealthState
	Shield HealthState
}

type PlayerCombatant interface {
	Combatant
	GetFightOption(slot int) *rpg.SkillId
	GetCurrentShield() *HealthState
	GetCurrentSync() *HealthState
}

type Player struct {
	*rpg.Animech
	*CurrentCombatantSkills
	Statuses           *AppliedStatuses
	CurrentPrimortal   int
	NextSkill          *rpg.SkillId
	NextSkillCommitted bool
	Tempo              *Tempo

	Shield      *HealthState
	Syncs       map[int]*HealthState
	HealthFlash *HealthFlash

	baseAnimation *anim.AnimatedSprite
}

func NewPlayer(animech *rpg.Animech) *Player {
	return &Player{
		Animech:                animech,
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Statuses:               NewAppliedStatuses(),
		Tempo:                  &Tempo{},
		Shield:                 NewHealthState(animech.GetMaxShield()),
		Syncs:                  make(map[int]*HealthState),
		HealthFlash:            NewDamageFlashMask(),
		baseAnimation:          anim.LoadTilesheetAnimation(atlas, "animech/combat_animech", "default"),
	}
}

func (p *Player) Update(timeDelta float64) {
	p.HealthFlash.Update(timeDelta)
}

func (p *Player) GetColorMask() pixel.RGBA {
	return p.HealthFlash.getMask()
}

func (p *Player) PeekNextSkill() *rpg.SkillId {
	return p.NextSkill
}

func (p *Player) GetStatuses() *AppliedStatuses {
	return p.Statuses
}

func (p *Player) PopNextSkill() *rpg.SkillId {
	if p.NextSkill == nil {
		return nil
	}
	next := p.NextSkill
	// keep selection and just keep hitting a, or p.NextSkill = nil and force re selection
	p.NextSkillCommitted = false
	return next
}

func (p *Player) IsNextSkillCommitted() bool {
	return p.NextSkillCommitted
}

func (p *Player) GetTempo() *Tempo {
	return p.Tempo
}

func (p *Player) getAnimech() *rpg.Animech {
	return p.Animech
}

var _ PlayerCombatant = &Player{}

func (p *Player) GetStats() rpg.CombatantStats {
	return rpg.CombatantStats{
		Tempo:        p.GetTempo().GetCurrent(),
		Stance:       p.GetCurrentSkill().GetCurrentStance(),
		StatusLevels: p.Statuses.GetLevels(),
	}
}

func (p *Player) AdjustHealth(amount int) {
	if amount == 0 {
		return
	}
	p.HealthFlash.basedOnAdjust(amount)
	if amount > 0 { // healing
		amount = p.GetCurrentSync().AdjustTarget(amount)
		if amount != 0 {
			p.Shield.AdjustTarget(amount)
		}
	} else { // damaging
		amount = p.Shield.AdjustTarget(amount)
		if amount != 0 {
			p.GetCurrentSync().AdjustTarget(amount)
		}
	}
}

func (p *Player) GetFightOption(slot int) *rpg.SkillId {
	animech := p.getAnimech()
	var id rpg.SkillId
	switch slot {
	case 0:
		id = animech.SkillSet.Skill1
	case 1:
		id = animech.SkillSet.Skill2
	case 2:
		id = animech.SkillSet.Skill3
	case 3:
		id = animech.SkillSet.Skill4
	}
	if id == rpg.UnsetSkillId {
		return nil
	}
	return &id
}

func (p *Player) GetCurrentSync() *HealthState {
	sync, exists := p.Syncs[p.CurrentPrimortal]
	if !exists {
		sync = NewHealthState(p.Animech.GetMaxSync())
		p.Syncs[p.CurrentPrimortal] = sync
	}
	return sync
}

func (p *Player) GetCurrentShield() *HealthState {
	return p.Shield
}

func (p *Player) GetTotalMaxHealth() int {
	return p.Shield.Max + p.GetCurrentSync().Max
}

func (p *Player) IsDead() bool {
	return p.GetCurrentSync().Current < 1
}
func (p *Player) IsPlayer() bool {
	return true
}

func (p *Player) Name() string {
	return "Animech" // todo plumb player name
}

func (p *Player) GetAnimation() *anim.AnimatedSprite {
	return p.baseAnimation
}

type SkillFightOption rpg.Skill

func (o SkillFightOption) OptionName() string {
	return o.Name
}

func (o SkillFightOption) Trigger(s *State) {
	game.DebugNotificationf("Triggering skill: %s", o.Name)
}
