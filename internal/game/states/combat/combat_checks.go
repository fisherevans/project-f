package combat

import (
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/rs/zerolog/log"
)

type CheckAgainst struct {
	Stance rpg.CombatStance
}

type TickListener func(t rpg.SkillTick, a CheckAgainst) bool

type SkillCondition struct {
	OnPlayerTick   TickListener
	OnOpponentTick TickListener
}

func NewSkillCondition() *SkillCondition {
	return &SkillCondition{
		OnPlayerTick: func(t rpg.SkillTick, a CheckAgainst) bool {
			return false
		},
		OnOpponentTick: func(t rpg.SkillTick, a CheckAgainst) bool {
			return false
		},
	}
}

func (c *SkillCondition) WithOnPlayerTick(tick TickListener) *SkillCondition {
	c.OnPlayerTick = tick
	return c
}

func (c *SkillCondition) WithOnOpponentTick(tick TickListener) *SkillCondition {
	c.OnOpponentTick = tick
	return c
}

func stanceAtTick(i *SkillInstance, tick int) rpg.CombatStance {
	if i == nil || tick <= 0 || tick >= len(i.Skill.Ticks) {
		return rpg.TickStanceNone
	}
	if i.InterruptedAt >= tick {
		return rpg.TickStanceNone
	}
	lastTickStance := i.Skill.Ticks[tick-1].StanceType
	currentTickStance := i.Skill.Ticks[tick].StanceType
	if lastTickStance != currentTickStance {
		return rpg.TickStanceNone
	}
	return lastTickStance
}

func (c *SkillCondition) Check(s *State) bool {
	playerSkillCurrent := s.Player.GetCurrentSkill()
	opponentSkillCurrent := s.Opponent.GetCurrentSkill()
	if playerSkillCurrent == nil || opponentSkillCurrent == nil {
		return false
	}

	var aSkill, bSkill *SkillInstance
	var onA, onB TickListener
	if s.Battle.TickPlayerNext {
		aSkill = playerSkillCurrent
		bSkill = opponentSkillCurrent
		onA = c.OnPlayerTick
		onB = c.OnOpponentTick
	} else {
		aSkill = opponentSkillCurrent
		bSkill = playerSkillCurrent
		onA = c.OnOpponentTick
		onB = c.OnPlayerTick
	}

	aTick := aSkill.NextTick
	bTick := bSkill.NextTick
	for aTick <= aSkill.Duration() && bTick <= bSkill.Duration() {
		aCurrent := aSkill.Skill.Ticks[aTick]
		bCurrent := bSkill.Skill.Ticks[bTick]
		aOpp := CheckAgainst{Stance: stanceAtTick(aSkill, aTick)}
		bOpp := CheckAgainst{Stance: stanceAtTick(bSkill, bTick)}
		if onA(aCurrent, bOpp) {
			return true
		}
		if onB(bCurrent, aOpp) {
			return true
		}
		aTick++
		bTick++
	}
	return false
}

type ConditionSleeper struct {
	start   int
	waitFor int
	wrapped func(state *State) bool
}

func WaitSomeTicks(waitFor int, andThen func(state *State) bool) func(state *State) bool {
	if waitFor <= 0 {
		log.Fatal().Msg("waitFor must be greater than 0")
	}
	if andThen == nil {
		log.Fatal().Msg("andThen must be provided")
	}
	start := -1
	return func(state *State) bool {
		if start <= 0 {
			start = state.Battle.TicksTriggered
		}
		if state.Battle.TicksTriggered < start+waitFor {
			return false
		}
		return andThen(state)
	}
}
