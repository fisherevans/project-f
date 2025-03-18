package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fmt"
	"github.com/gopxl/pixel/v2"
)

type Battle struct {
	PendingProgress float64
	TickPlayerNext  bool
	PlayerSkill     *SkillInstance
	OpponentSkill   *SkillInstance

	PreviousPlayerSkill   *SkillInstance
	PreviousOpponentSkill *SkillInstance
}

type BattleUpdateParams struct {
	PlayerNextSkill   func() *rpg.SkillId
	OpponentNextSkill func() *rpg.SkillId
}

func (b *Battle) Update(ctx *game.Context, s *State, timeDelta float64, params BattleUpdateParams) {
	if b.PlayerSkill == nil || (b.PlayerSkill.NextTick == 0 && !b.TickPlayerNext) {
		nextSkill := params.PlayerNextSkill()
		if nextSkill != nil {
			if b.PlayerSkill == nil {
				s.Tempo.Increment()
			}
			b.PlayerSkill = newInstance(*nextSkill)
		}
	}
	if b.OpponentSkill == nil {
		nextSkill := params.OpponentNextSkill()
		if nextSkill != nil {
			b.OpponentSkill = newInstance(*nextSkill)
		}
	}

	ctx.DebugBR(fmt.Sprintf("PendingProgress: %f", b.PendingProgress))
	ctx.DebugBR(fmt.Sprintf("TickPlayerNext: %t", b.TickPlayerNext))
	ctx.DebugBR(fmt.Sprintf("PlayerSkill: %s", b.PlayerSkill))
	ctx.DebugBR(fmt.Sprintf("OpponentSkill: %s", b.OpponentSkill))

	if b.TickPlayerNext && b.PlayerSkill == nil {
		s.Tempo.Reset()
		return
	}
	if !b.TickPlayerNext && b.OpponentSkill == nil {
		return
	}

	tps := ticksPerSecond
	if ctx.DebugToggles.F1().ToggleState() {
		tps *= 2
	}
	b.PendingProgress += timeDelta * tps
	for b.PendingProgress >= 1 {
		if b.TickPlayerNext {
			dmg, over := b.PlayerSkill.Tick(ctx, s, s.Player.GetCombatant(), s.Opponent)
			if over {
				b.PreviousPlayerSkill = b.PlayerSkill
				b.PlayerSkill = nil
			}
			s.emitDamageFx(dmg, false)
		} else {
			dmg, over := b.OpponentSkill.Tick(ctx, s, s.Opponent, s.Player.GetCombatant())
			if over {
				b.PreviousOpponentSkill = b.OpponentSkill
				b.OpponentSkill = nil
			}
			s.emitDamageFx(dmg, true)
		}
		b.PendingProgress -= 1.0
		b.TickPlayerNext = !b.TickPlayerNext
	}
}

func (s *State) emitDamageFx(dmgs []rpg.DamageResult, damagingPlayer bool) {
	if len(dmgs) == 0 {
		return
	}
	pos := pixel.V(game.GameWidth*0.85, game.GameHeight*0.5)
	velocity := pixel.V(-20, rand.Float64()*50+50)
	if damagingPlayer {
		pos.X = game.GameWidth * 0.15
		velocity.X = velocity.X * -1
	}
	for _, dmg := range dmgs {
		s.fx = append(s.fx, &DamageFX{
			Damage:     dmg,
			Position:   pos,
			Velocity:   velocity,
			SpeedScale: 2,
		})
	}
}

func (i *SkillInstance) Tick(ctx *game.Context, s *State, source Combatant, target Combatant) ([]rpg.DamageResult, bool) {
	//log.Info().Msgf("%s: %s.Tick(%d:%d)", source.Name(), i, i.NextTick, i.Duration)
	var dmg []rpg.DamageResult
	if i.NextTick <= i.Duration {
		dmg = i.OnTick(ctx, s, i.NextTick, source, target)
		i.NextTick++
	}
	return dmg, i.NextTick > i.Duration
}
