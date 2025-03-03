package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
)

var typeOptionKey = map[int]input.Direction{
	0: input.Up,
	1: input.Right,
	2: input.Down,
	3: input.Left,
}

type PlayerHealth struct {
	Integrity HealthState
	Sync      HealthState
	Shield    HealthState
}

type PlayerCombatant interface {
	Combatant
	GetFightOption(slot int) *rpg.SkillId
	GetCurrentShield() HealthState
	GetCurrentSync() HealthState
}

type Player struct {
	*rpg.DeployedAnimech
	CurrentPrimortal int
	NextSkill        *rpg.SkillId
}

func (p *Player) PopNextSkill() *rpg.SkillId {
	if p.NextSkill == nil {
		return nil
	}
	next := p.NextSkill
	p.NextSkill = nil
	return next
}

func (p *Player) GetCombatant() PlayerCombatant {
	if p.CurrentPrimortal == -1 {
		return nil // TODO void mode
	}
	return &PlayerPrimortal{
		Player:            p,
		DeployedPrimortal: p.DeployedPrimortals[p.CurrentPrimortal],
	}
}

type PlayerPrimortal struct {
	*rpg.DeployedPrimortal
	*Player
}

var _ PlayerCombatant = &PlayerPrimortal{}

func (p *PlayerPrimortal) GetStats() CombatantStats {
	return CombatantStats{
		BodyType:        p.Base().BodyType,
		Affinities:      []rpg.SkillType{p.Base().Affinity},
		PhysicalAttack:  p.Base().BasePhysicalAttack + p.AdditionalPhysicalAttack,
		PhysicalDefense: p.Base().BasePhysicalDefense + p.AdditionalPhysicalDefense,
		AetherAttack:    p.Base().BaseAetherAttack + p.AdditionalAetherAttack,
		AetherDefense:   p.Base().BaseAetherDefense + p.AdditionalAetherDefense,
	}
}

func (p *PlayerPrimortal) ApplyDamage(damage rpg.DamageResult) DamageOutcome {
	applyDamage := func(current, dmg int) (int, int) {
		current -= dmg
		if current >= 0 {
			return current, 0
		}
		return 0, -current
	}
	remainingDamge := damage.TotalDamage
	p.CurrentShield, remainingDamge = applyDamage(p.CurrentShield, remainingDamge)
	p.CurrentSync, remainingDamge = applyDamage(p.CurrentSync, remainingDamge)
	p.CurrentIntegrity, remainingDamge = applyDamage(p.CurrentIntegrity, remainingDamge)
	return DamageOutcome{
		Perished: p.CurrentIntegrity <= 0,
	}
}

func (p *PlayerPrimortal) GetFightOption(slot int) *rpg.SkillId {
	if slot >= len(p.SelectedSkills) {
		return nil
	}
	skill := p.SelectedSkills[slot]
	return &skill
}

func (p *PlayerPrimortal) GetCurrentSync() HealthState {
	return HealthState{
		Max:     p.Base().BaseSync + p.AdditionalSync,
		Current: p.CurrentSync,
	}
}

func (p *PlayerPrimortal) GetCurrentShield() HealthState {
	return HealthState{
		Max:     rpg.BaseAnimechShield + p.AdditionalShield,
		Current: p.CurrentShield,
	}
}

func (p *PlayerPrimortal) Name() string {
	if p.Nickname != "" {
		return p.Nickname
	}
	return p.Base().Name
}

type SkillFightOption rpg.Skill

func (o SkillFightOption) OptionName() string {
	return o.Name
}

func (o SkillFightOption) Trigger(ctx *game.Context, s *State) {
	ctx.Notify("Triggering skill: %s", o.Name)
}
