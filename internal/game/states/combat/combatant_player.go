package combat

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
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
	*rpg.DeployedAnimech
	*CurrentCombatantSkills
	Statuses           *AppliedStatuses
	CurrentPrimortal   int
	NextSkill          *rpg.SkillId
	NextSkillCommitted bool
	Tempo              *Tempo

	Shield          *HealthState
	Syncs           map[int]*HealthState
	DamageFlashMask *DamageFlashMask
}

func NewPlayer(deployed *rpg.DeployedAnimech) *Player {
	return &Player{
		DeployedAnimech:        deployed,
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
		Statuses:               NewAppliedStatuses(),
		Tempo:                  &Tempo{},
		Shield:                 NewHealthState(deployed.GetMaxShield()),
		Syncs:                  make(map[int]*HealthState),
		DamageFlashMask:        NewDamageFlashMask(),
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

func (p *Player) getDeployedPrimortal() *rpg.DeployedPrimortal {
	return p.DeployedPrimortals[p.CurrentPrimortal]
}

var _ PlayerCombatant = &Player{}

func (p *Player) GetStats() CombatantStats {
	//dp := p.getDeployedPrimortal()
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
	dp := p.getDeployedPrimortal()
	if slot >= len(dp.SelectedSkills) {
		return nil
	}
	skill := dp.SelectedSkills[slot]
	return &skill
}

func (p *Player) GetCurrentSync() *HealthState {
	sync, exists := p.Syncs[p.CurrentPrimortal]
	if !exists {
		dp := p.getDeployedPrimortal()
		sync = NewHealthState(dp.GetMaxSync())
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
	dp := p.getDeployedPrimortal()
	if dp.Nickname != "" {
		return dp.Nickname
	}
	return dp.Base().Name
}

type SkillFightOption rpg.Skill

func (o SkillFightOption) OptionName() string {
	return o.Name
}

func (o SkillFightOption) Trigger(ctx *game.Context, s *State) {
	ctx.Notify("Triggering skill: %s", o.Name)
}
