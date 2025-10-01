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
	*rpg.Run
	*rpg.Animech
	*CurrentCombatantSkills
	Statuses           *AppliedStatuses
	CurrentPrimortal   int
	NextSkill          *rpg.SkillId
	NextSkillCommitted bool
	Tempo              *Tempo

	Shield          *HealthState
	Syncs           map[int]*HealthState
	DamageFlashMask *DamageFlashMask

	baseAnimation *anim.AnimatedSprite
}

func NewPlayer(animech *rpg.Animech, run *rpg.Run) *Player {
	return &Player{
		Animech:                animech,
		Run:                    run,
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Statuses:               NewAppliedStatuses(),
		Tempo:                  &Tempo{},
		Shield:                 NewHealthState(animech.GetMaxShield()),
		Syncs:                  make(map[int]*HealthState),
		DamageFlashMask:        NewDamageFlashMask(),
		baseAnimation:          anim.IdleRobot(atlas),
	}
}

func (p *Player) Update(timeDelta float64) {
	p.DamageFlashMask.Update(timeDelta)
}

func (p *Player) GetColorMask() pixel.RGBA {
	return p.DamageFlashMask.getMask()
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

func (p *Player) GetStats() CombatantStats {
	//animech := p.getAnimech()
	return CombatantStats{
		Stance: p.GetCurrentSkill().GetCurrentStance(),
	}
}

func (p *Player) ApplyDamage(damage int) {
	adjustment := -damage
	adjustment = p.Shield.AdjustTarget(adjustment)
	if adjustment != 0 {
		p.GetCurrentSync().AdjustTarget(adjustment)
	}
	p.DamageFlashMask.damaged()
}

func (p *Player) GetFightOption(slot int) *rpg.SkillId {
	animech := p.getAnimech()
	var id rpg.SkillId
	switch slot {
	case 1:
		id = animech.SkillSet.Skill1
	case 2:
		id = animech.SkillSet.Skill2
	case 3:
		id = animech.SkillSet.Skill3
	case 4:
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
	return "Player" // todo plumb player name
}

func (p *Player) GetAnimation() *anim.AnimatedSprite {
	return p.baseAnimation
}

type SkillFightOption rpg.Skill

func (o SkillFightOption) OptionName() string {
	return o.Name
}

func (o SkillFightOption) Trigger(ctx *game.Context, s *State) {
	ctx.Notify("Triggering skill: %s", o.Name)
}
