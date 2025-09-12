package combat

import (
	"fmt"
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
)

type TickHandler func(ctx *game.Context, s *State, tick int, source Combatant, target Combatant) []rpg.DamageResult

type SkillInstance struct {
	Skill    *rpg.Skill
	NextTick int
	Duration int // 1 to len(skill.Ticks)-1
	OnTick   TickHandler
}

func (i *SkillInstance) String() string {
	if i == nil {
		return "<none>"
	}
	return fmt.Sprintf("SkillInstance{NextTick: %d, Duration: %d}", i.NextTick, i.Duration)
}

func (i *SkillInstance) GetCurrentStance() rpg.CombatStance {
	if i == nil || i.NextTick == 0 || i.NextTick >= len(i.Skill.Ticks) {
		return rpg.TickStanceNone
	}
	lastTickStance := i.Skill.Ticks[i.NextTick-1].StanceType
	nextTickStance := i.Skill.Ticks[i.NextTick].StanceType
	if lastTickStance != nextTickStance {
		return rpg.TickStanceNone
	}
	return lastTickStance
}

func newInstance(skillId rpg.SkillId) *SkillInstance {
	skill := skillId.Get()
	var handlers []TickHandler
	for _, tick := range skill.Ticks {
		handlers = append(handlers, func(ctx *game.Context, s *State, tickId int, source Combatant, target Combatant) []rpg.DamageResult {
			sourceStats := source.GetStats()
			targetStats := target.GetStats()
			var allDamage []rpg.DamageResult
			for _, effect := range tick.Effects {
				if effect.Damage != nil {
					damage := effect.Damage.Amount
					if effect.Damage.RandomVariance > 0 {
						damage += rand.Intn(effect.Damage.RandomVariance*2+1) - effect.Damage.RandomVariance
					}
					result := rpg.ComputeDamage(rpg.DamageSource{
						BaseDamage:     damage,
						Affinities:     sourceStats.Affinities,
						DamageMedium:   effect.Damage.Medium,
						SkillType:      skill.Type,
						PhysicalAttack: sourceStats.PhysicalAttack,
						AetherAttack:   sourceStats.AetherAttack,
						Tempo:          source.GetTempo().GetCurrent(),
						Stance:         sourceStats.Stance,
					}, rpg.DamageTarget{
						TargetType:      targetStats.BodyType,
						Affinities:      targetStats.Affinities,
						PhysicalDefence: targetStats.PhysicalDefense,
						AetherDefence:   targetStats.AetherDefense,
						Stance:          sourceStats.Stance,
					})
					allDamage = append(allDamage, result)
					ctx.Notify("damage from %d to %d", damage, result.TotalDamage)
					target.ApplyDamage(result)
				}
			}
			return allDamage
		})
	}
	return &SkillInstance{
		Skill:    &skill,
		Duration: len(handlers) - 1,
		OnTick: func(ctx *game.Context, s *State, tickId int, source Combatant, target Combatant) []rpg.DamageResult {
			return handlers[tickId](ctx, s, tickId, source, target)
		},
	}
}
