package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fmt"
)

type TickHandler func(ctx *game.Context, s *State, tick int, source Combatant, target Combatant) []rpg.DamageResult

type SkillInstance struct {
	Skill       *rpg.Skill
	NextTick    int
	Duration    int // 1+
	OnTick      TickHandler
	DisplayType func(int) rpg.TickDisplayType
}

func (i *SkillInstance) String() string {
	if i == nil {
		return "<none>"
	}
	return fmt.Sprintf("SkillInstance{NextTick: %d, Duration: %d}", i.NextTick, i.Duration)
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
				if effect.StaticDamage != nil {
					result := rpg.ComputeDamage(rpg.DamageSource{
						BaseDamage:     effect.StaticDamage.Amount,
						Affinities:     sourceStats.Affinities,
						DamageMedium:   effect.StaticDamage.Medium,
						SkillType:      skill.Type,
						PhysicalAttack: sourceStats.PhysicalAttack,
						AetherAttack:   sourceStats.AetherAttack,
					}, rpg.DamageTarget{
						TargetType:      targetStats.BodyType,
						Affinities:      targetStats.Affinities,
						PhysicalDefence: targetStats.PhysicalDefense,
						AetherDefence:   targetStats.AetherDefense,
					})
					allDamage = append(allDamage, result)
					outcome := target.ApplyDamage(result)
					if outcome.Perished {
						ctx.Notify("%s DIED!", target.Name())
					}
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
		DisplayType: func(tickId int) rpg.TickDisplayType {
			return skill.Ticks[tickId].DisplayType
		},
	}
}
