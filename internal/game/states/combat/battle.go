package combat

import (
	"fmt"

	"fisherevans.com/project/f/internal/game"
)

type Battle struct {
	PendingProgress float64
	TickPlayerNext  bool

	PlayerSkillEnding   bool
	OpponentSkillEnding bool
}

func (b *Battle) GetPlayerCurrentTickProgress() float64 {
	playerProgress := b.PendingProgress / 2.0
	if b.TickPlayerNext {
		playerProgress += 0.5
	}
	return playerProgress
}

func (b *Battle) GetOpponentCurrentTickProgress() float64 {
	opponentProgress := b.PendingProgress / 2.0
	if !b.TickPlayerNext {
		opponentProgress += 0.5
	}
	return opponentProgress
}

func (b *Battle) Update(s *State, timeDelta float64) {
	game.DebugBRf(fmt.Sprintf("PendingProgress: %f", b.PendingProgress))
	game.DebugBRf(fmt.Sprintf("TickPlayerNext: %t", b.TickPlayerNext))
	game.DebugBRf(fmt.Sprintf("PlayerSkill: %s", s.Player.GetCurrentSkill()))
	game.DebugBRf(fmt.Sprintf("OpponentSkill: %s", s.Opponent.GetCurrentSkill()))

	if b.TickPlayerNext && s.Player.GetCurrentSkill() == nil {
		if s.Player.PeekNextSkill() == nil || !s.Player.IsNextSkillCommitted() {
			s.Player.Tempo.Reset()
			return
		}
		s.Player.SetCurrentSkill(newInstance(*s.Player.PopNextSkill()))
		s.Player.Tempo.Increment()
	}

	if !b.TickPlayerNext && s.Opponent.GetCurrentSkill() == nil {
		if s.Opponent.PeekNextSkill() == nil || !s.Opponent.IsNextSkillCommitted() {
			//s.Opponent.Tempo.Reset()
			return
		}
		s.Opponent.SetCurrentSkill(newInstance(*s.Opponent.PopNextSkill()))
		//s.Opponent.Tempo.Increment()
	}

	tps := ticksPerSecond
	if game.DebugToggles().F1().ToggleState() {
		tps *= 2
	}
	b.PendingProgress += timeDelta * tps
	for b.PendingProgress >= 1 {
		if b.TickPlayerNext {
			s.Opponent.GetStatuses().DecreaseStacks()
			if b.OpponentSkillEnding {
				s.Opponent.SetCurrentSkill(nil)
				b.OpponentSkillEnding = false
			}
			s.Player.GetStatuses().ApplyEffects(s, s.Player)
			over := s.Player.GetCurrentSkill().Tick(s, s.Player, s.Opponent)
			if over {
				b.PlayerSkillEnding = true
			}
		} else {
			s.Player.GetStatuses().DecreaseStacks()
			if b.PlayerSkillEnding {
				s.Player.SetCurrentSkill(nil)
				b.PlayerSkillEnding = false
			}
			s.Opponent.GetStatuses().ApplyEffects(s, s.Opponent)
			over := s.Opponent.GetCurrentSkill().Tick(s, s.Opponent, s.Player)
			if over {
				b.OpponentSkillEnding = true
			}
		}
		b.PendingProgress -= 1.0
		b.TickPlayerNext = !b.TickPlayerNext
	}
}

func (i *SkillInstance) Tick(s *State, source Combatant, target Combatant) bool {
	if i.NextTick <= i.Duration() {
		i.OnTick(s, i.NextTick, source, target)
		i.NextTick++
	}
	return i.NextTick > i.Duration()
}
