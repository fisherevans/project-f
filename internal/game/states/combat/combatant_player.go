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
	CurrentPrimortal int
	NextSkill        *rpg.SkillId
	Tempo            *Tempo

	Shield          *HealthState
	Syncs           map[int]*HealthState
	DamageFlashMask *DamageFlashMask
}

func NewPlayer(deployed *rpg.DeployedAnimech) *Player {
	return &Player{
		DeployedAnimech:        deployed,
		CurrentCombatantSkills: NewCurrentCombatantSkills(),
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

func (p *Player) PopNextSkill() *rpg.SkillId {
	if p.NextSkill == nil {
		return nil
	}
	next := p.NextSkill
	p.NextSkill = nil
	return next
}

func (p *Player) GetTempo() *Tempo {
	return p.Tempo
}

func (p *Player) getDeployedPrimortal() *rpg.DeployedPrimortal {
	return p.DeployedPrimortals[p.CurrentPrimortal]
}

var _ PlayerCombatant = &Player{}

func (p *Player) GetStats() CombatantStats {
	dp := p.getDeployedPrimortal()
	return CombatantStats{
		BodyType:        dp.Base().BodyType,
		Affinities:      []rpg.SkillType{dp.Base().Affinity},
		PhysicalAttack:  dp.Base().BasePhysicalAttack + dp.AdditionalPhysicalAttack,
		PhysicalDefense: dp.Base().BasePhysicalDefense + dp.AdditionalPhysicalDefense,
		AetherAttack:    dp.Base().BaseAetherAttack + dp.AdditionalAetherAttack,
		AetherDefense:   dp.Base().BaseAetherDefense + dp.AdditionalAetherDefense,
		Stance:          p.GetCurrentSkill().GetCurrentStance(),
	}
}

func (p *Player) ApplyDamage(damage rpg.DamageResult) {
	adjustment := -damage.TotalDamage
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
