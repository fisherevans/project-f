package combat

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
)

type TickHandler func(s *State, tick int, source Combatant, target Combatant)

type SkillInstance struct {
	Skill         *rpg.Skill
	NextTick      int
	InterruptedAt int
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
	if i.InterruptedAt >= i.NextTick {
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
	return &SkillInstance{
		Skill:         &skill,
		InterruptedAt: -1,
	}
}

func (si *SkillInstance) Duration() int {
	return si.Skill.Duration()
}

func (si *SkillInstance) OnTick(s *State, tickId int, source Combatant, target Combatant) {
	if si.InterruptedAt >= tickId {
		return
	}
	tick := si.Skill.Ticks[tickId]
	sourceStats := source.GetStats()
	targetStats := target.GetStats()
	for _, effect := range tick.Effects {
		if effect.Status != nil {
			applyTo := target
			if effect.Self {
				applyTo = source
			}
			doApply := true
			switch effect.Status.RequireExistingStacks {
			case rpg.SkillTickStatusRequireExistingStacks:
				doApply = applyTo.GetStatuses().HasStatus(effect.Status.Status)
			case rpg.SkillTickStatusRequireNoStacks:
				doApply = !applyTo.GetStatuses().HasStatus(effect.Status.Status)
			}
			if doApply {
				applyTo.GetStatuses().Add(effect.Status.Status, effect.Status.Stacks)
				word := effect.Status.Status.PastTense()
				color, colorExists := colors.StatusColors[effect.Status.Status]
				if word != "" && colorExists {
					s.AddFX(NewWordFX(word, color, applyTo))
				}
			}
		}
		if effect.Damage != nil {
			result := rpg.ComputeDamage(*effect.Damage, sourceStats, targetStats)
			game.DebugNotificationf("damage from %d to %d", effect.Damage.Amount, result.TargetDamage)
			s.AdjustHealth(-result.TargetDamage, target, nil)
			s.AdjustHealth(-result.SourceDamage, source, nil)
			target.GetStatuses().ReduceResult(result.TargetStatusStackReductions)
			source.GetStatuses().ReduceResult(result.SourceStatusStackReductions)
			if result.InterruptSkill {
				if target.GetCurrentSkill() != nil {
					target.GetCurrentSkill().Interrupt()
					target.GetTempo().Reset()
					s.fx = append(s.fx, NewWordFX("Interrupt!", colors.HexString("#daff4b"), target))
				}
			}
			if result.Missed {
				s.fx = append(s.fx, NewWordFX("Missed!", colors.HexString("#b9b9b9"), source))
			}
		}
	}
}

func (i *SkillInstance) Interrupt() {
	if i.NextTick > i.InterruptedAt {
		i.InterruptedAt = i.NextTick
	}

}
