package combat

import (
	"fmt"
	"math/rand"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
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

func (b *Battle) Update(ctx *game.Context, s *State, timeDelta float64) {
	ctx.DebugBR(fmt.Sprintf("PendingProgress: %f", b.PendingProgress))
	ctx.DebugBR(fmt.Sprintf("TickPlayerNext: %t", b.TickPlayerNext))
	ctx.DebugBR(fmt.Sprintf("PlayerSkill: %s", s.Player.GetCurrentSkill()))
	ctx.DebugBR(fmt.Sprintf("OpponentSkill: %s", s.Opponent.GetCurrentSkill()))

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
	if ctx.DebugToggles.F1().ToggleState() {
		tps *= 2
	}
	b.PendingProgress += timeDelta * tps
	for b.PendingProgress >= 1 {
		if b.TickPlayerNext {
			if b.OpponentSkillEnding {
				s.Opponent.SetCurrentSkill(nil)
				b.OpponentSkillEnding = false
			}
			dmg, over := s.Player.GetCurrentSkill().Tick(ctx, s, s.Player, s.Opponent)
			if over {
				b.PlayerSkillEnding = true
			}
			s.emitDamageFx(dmg, false)
		} else {
			if b.PlayerSkillEnding {
				s.Player.SetCurrentSkill(nil)
				b.PlayerSkillEnding = false
			}
			dmg, over := s.Opponent.GetCurrentSkill().Tick(ctx, s, s.Opponent, s.Player)
			if over {
				b.OpponentSkillEnding = true
			}
			s.emitDamageFx(dmg, true)
		}
		b.PendingProgress -= 1.0
		b.TickPlayerNext = !b.TickPlayerNext
	}
}

func (s *State) emitDamageFx(dmgs []rpg.DamageResult, damagingPlayer bool) {
	for _, dmg := range dmgs {
		if dmg.TargetDamage > 0 {
			if damagingPlayer {
				s.emitPlayerDamageFx(dmg.TargetDamage)
			} else {
				s.emitOpponentDamageFx(dmg.TargetDamage)
			}
		}
		if dmg.SourceDamage > 0 {
			if damagingPlayer {
				s.emitOpponentDamageFx(dmg.SourceDamage)
			} else {
				s.emitPlayerDamageFx(dmg.SourceDamage)
			}
		}
	}
}

func (s *State) emitPlayerDamageFx(damage int) {
	if damage == 0 {
		return
	}
	playerPosition := pixel.V(game.GameWidth*0.15, game.GameHeight*0.5)
	playerVelocity := pixel.V(20, rand.Float64()*50+50)
	s.fx = append(s.fx, &DamageFX{
		Damage:     damage,
		Position:   playerPosition,
		Velocity:   playerVelocity,
		SpeedScale: 2,
	})
}

func (s *State) emitOpponentDamageFx(damage int) {
	if damage == 0 {
		return
	}
	opponentPosition := pixel.V(game.GameWidth*0.85, game.GameHeight*0.5)
	opponentVelocity := pixel.V(-20, rand.Float64()*50+50)
	s.fx = append(s.fx, &DamageFX{
		Damage:     damage,
		Position:   opponentPosition,
		Velocity:   opponentVelocity,
		SpeedScale: 2,
	})
}

func (i *SkillInstance) Tick(ctx *game.Context, s *State, source Combatant, target Combatant) ([]rpg.DamageResult, bool) {
	var dmg []rpg.DamageResult
	if i.NextTick <= i.Duration {
		dmg = i.OnTick(ctx, s, i.NextTick, source, target)
		i.NextTick++
	}
	return dmg, i.NextTick > i.Duration
}
