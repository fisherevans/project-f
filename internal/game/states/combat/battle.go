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

	PlayerSkill       *SkillInstance
	PlayerSkillEnding bool

	OpponentSkill       *SkillInstance
	OpponentSkillEnding bool
}

type BattleUpdateParams struct {
	PlayerNextSkill   func() *rpg.SkillId
	OpponentNextSkill func() *rpg.SkillId
}

func (b *Battle) Update(ctx *game.Context, s *State, timeDelta float64, params BattleUpdateParams) {
	ctx.DebugBR(fmt.Sprintf("PendingProgress: %f", b.PendingProgress))
	ctx.DebugBR(fmt.Sprintf("TickPlayerNext: %t", b.TickPlayerNext))
	ctx.DebugBR(fmt.Sprintf("PlayerSkill: %s", b.PlayerSkill))
	ctx.DebugBR(fmt.Sprintf("OpponentSkill: %s", b.OpponentSkill))

	if b.TickPlayerNext && b.PlayerSkill == nil {
		nextSkill := params.PlayerNextSkill()
		if nextSkill != nil {
			b.PlayerSkill = newInstance(*nextSkill)
			s.Player.Tempo.Increment()
		} else {
			s.Player.Tempo.Reset()
			return
		}
	}

	if !b.TickPlayerNext && b.OpponentSkill == nil {
		nextSkill := params.OpponentNextSkill()
		if nextSkill != nil {
			b.OpponentSkill = newInstance(*nextSkill)
		} else {
			return
		}
	}

	tps := ticksPerSecond
	if ctx.DebugToggles.F1().ToggleState() {
		tps *= 2
	}
	b.PendingProgress += timeDelta * tps
	for b.PendingProgress >= 1 {
		if b.TickPlayerNext {
			if b.OpponentSkillEnding {
				b.OpponentSkill = nil
				b.OpponentSkillEnding = false
			}
			dmg, over := b.PlayerSkill.Tick(ctx, s, s.Player, s.Opponent)
			if over {
				b.PlayerSkillEnding = true
			}
			s.emitDamageFx(dmg, false)
		} else {
			if b.PlayerSkillEnding {
				b.PlayerSkill = nil
				b.PlayerSkillEnding = false
			}
			dmg, over := b.OpponentSkill.Tick(ctx, s, s.Opponent, s.Player)
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
