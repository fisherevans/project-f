package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/colors/typecolors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/rs/zerolog/log"
	"image/color"
	"math"
	"math/rand"
	"strings"
)

type OnComplete func(ctx *game.Context, s *State)

type State struct {
	Player     *Player
	Opponent   Opponent
	OnComplete OnComplete
	Battle     *Battle

	fx []FX
}

func New(animech *rpg.DeployedAnimech, onComplete OnComplete) *State {
	return &State{
		Player: &Player{
			DeployedAnimech: animech,
		},
		Opponent: &Wall{
			MaxHealth:     100,
			CurrentHealth: 100,
		},
		OnComplete: onComplete,
		Battle:     &Battle{},
	}
}

type HealthState struct {
	Max     int
	Current int
}

type CombatantStats struct {
	BodyType        rpg.BodyType
	Affinities      []rpg.SkillType
	PhysicalAttack  int
	PhysicalDefense int
	AetherAttack    int
	AetherDefense   int
}

type DamageOutcome struct {
	Perished bool
}

type Combatant interface {
	Name() string
	ApplyDamage(damage rpg.DamageResult) DamageOutcome
	GetStats() CombatantStats
}

func (s *State) ClearColor() color.Color {
	return color.Black
}

var padding = 5
var stateText = textbox.NewInstance(textbox.FontLargeSpaced, textbox.
	NewConfig((game.GameWidth-padding*3)/2).
	Foreground(colors.White).
	RenderFrom(textbox.TopLeft))

var backgroundSprite = resources.NonAtlasSprites["background_combat_sample"].Sprite

var imd = imdraw.New(nil)

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {

	// pick next skill for player
	for optionId, direction := range typeOptionKey {
		if ctx.Controls.DPad().DirectionJustPressed(direction) {
			option := s.Player.GetCombatant().GetFightOption(optionId)
			if option != nil {
				s.Player.NextSkill = option
			}
		}
	}

	s.Battle.Update(ctx, s, timeDelta, BattleUpdateParams{
		PlayerNextSkill: func() *SkillInstance {
			next := s.Player.PopNextSkill()
			if next == nil {
				return nil
			}
			return newInstance(*next)

		},
		OpponentNextSkill: func() *SkillInstance {
			return newInstance(rpg.Skill_Tackle.Id)
		},
	})

	var remainingFx []FX
	for _, fx := range s.fx {
		if !fx.Update(ctx, s, timeDelta) {
			remainingFx = append(remainingFx, fx)
		}
	}
	s.fx = remainingFx

	backgroundSprite.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Grey4)

	// while time left
	//  - progress time
	//  - if skill ends, pop next

	s.drawActiveSkills(ctx, target, targetBounds, pixel.IM.Moved(targetBounds.Center()))

	playerText := []string{
		"Player",
		s.Player.GetCombatant().Name(),
		fmt.Sprintf("Health: %d/%d", s.Player.GetCombatant().GetCurrentSync().Current, s.Player.GetCombatant().GetCurrentSync().Max),
	}
	for optionId := 0; optionId < 4; optionId++ {
		direction := typeOptionKey[optionId]
		option := s.Player.GetCombatant().GetFightOption(optionId)
		text := fmt.Sprintf("%s: ", direction.String())
		if option != nil {
			name := option.Get().Name
			if s.Player.NextSkill != nil && *s.Player.NextSkill == *option {
				name = "{+c:warm_5}" + name + "{-*}"
			}
			text += name
		} else {
			text += "-"
		}
		playerText = append(playerText, text)
	}
	playerContent := stateText.NewComplexContent(strings.Join(playerText, "\n"))
	playerContent.Update(ctx, timeDelta)
	stateText.Render(ctx, target, pixel.IM.Moved(pixel.V(float64(padding), float64(game.GameHeight-padding))), playerContent)

	opponentText := []string{
		"Opponent",
		s.Opponent.Name(),
		fmt.Sprintf("Health: %d/%d", s.Opponent.GetHealth().Current, s.Opponent.GetHealth().Max),
	}
	opponentContent := stateText.NewComplexContent(strings.Join(opponentText, "\n"), textbox.WithAlignment(textbox.AlignRight))
	opponentContent.Update(ctx, timeDelta)
	stateText.Render(ctx, target, pixel.IM.Moved(pixel.V(float64(padding*2+(game.GameWidth-padding*3)/2), float64(game.GameHeight-padding))), opponentContent)

	for _, fx := range s.fx {
		fx.Render(ctx, target)
	}
}

var noneSelectedSprite = resources.Sprites["combat_tick_none_selected"]

var tickBubbleDisplayNone = resources.TilesheetSprites[resources.TilesheetSpriteId{
	Tilesheet: "combat_tick_bubbles",
	Column:    1,
	Row:       1,
}]

var tickBubbleDisplayNormal = resources.TilesheetSprites[resources.TilesheetSpriteId{
	Tilesheet: "combat_tick_bubbles",
	Column:    2,
	Row:       1,
}]

var tickBubbleDisplaySpecial = resources.TilesheetSprites[resources.TilesheetSpriteId{
	Tilesheet: "combat_tick_bubbles",
	Column:    3,
	Row:       1,
}]

var tickHeight = 8
var tickLength = 10
var tickBarPadding = 2

func (s *State) drawActiveSkills(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, matrix pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	s.drawCombatantSkills(ctx, target, matrix.Moved(pixel.V(0, float64(tickBarPadding/2))), opponentProgress, s.Battle.OpponentSkill, nil)
	s.drawCombatantSkills(ctx, target, matrix.Moved(pixel.V(0, float64(-tickHeight-tickBarPadding/2))), playerProgress, s.Battle.PlayerSkill, s.Player.NextSkill)
	//matrix = matrix.Moved(pixel.V(float64(-tickLength/2), 0))
	resources.Sprites["combat_tick_eater"].Sprite.Draw(target, matrix)
}

func (s *State) drawCombatantSkills(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, currentTickProgress float64, currentSkill *SkillInstance, nextSkillId *rpg.SkillId) {
	noNextSkillAlpha := 1.0
	if currentSkill != nil {
		dx := -(currentTickProgress + float64(currentSkill.NextTick)) * float64(tickLength)
		matrix = matrix.Moved(pixel.V(dx, 0))
		mask := typecolors.SkillTypeColor(currentSkill.Skill.Type)
		s.drawSkill(ctx, target, matrix, currentSkill.Skill, mask, 1.0)
		matrix = matrix.Moved(pixel.V(float64((currentSkill.Duration+1)*tickLength), 0))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick)+currentTickProgress)/float64(currentSkill.Duration))
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := typecolors.SkillTypeColor(nextSkill.Type)
		mask = colors.ScaleColor(mask, 0.5)
		s.drawSkill(ctx, target, matrix, &nextSkill, mask, 1.0)
	} else {
		ctx.DebugBR("no skill!")
		x := noneSelectedSprite.Bounds.W() / 2
		y := noneSelectedSprite.Bounds.H() / 2
		noneSelectedSprite.Sprite.DrawColorMask(target, matrix.Moved(pixel.V(x, y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
	}
}

func (s *State) drawSkill(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, skill *rpg.Skill, mask pixel.RGBA, f float64) {
	if skill == nil {
		return
	}
	mask = colors.WithAlpha(mask, f)
	rect := pixel.R(0, 0,
		float64(tickLength+tickLength*(skill.Duration())),
		float64(tickHeight))
	frame := resources.Frames["combat_tick_frame"]
	frames.Draw(target, frame, rect, matrix, frames.WithColor(mask))

	for i := 0; i <= skill.Duration(); i++ {
		var sprite *pixel.Sprite
		switch skill.Ticks[i].DisplayType {
		case rpg.TickDisplayNone:
			sprite = tickBubbleDisplayNone.Sprite
		case rpg.TickDisplayNormal:
			sprite = tickBubbleDisplayNormal.Sprite
		}
		if sprite == nil {
			continue
		}
		sprite.DrawColorMask(target, matrix.Moved(pixel.V(float64(tickLength/2+tickLength*i), float64(tickHeight/2))), mask)
	}
}

type Battle struct {
	PendingProgress float64
	TickPlayerNext  bool
	PlayerSkill     *SkillInstance
	OpponentSkill   *SkillInstance
}

type BattleUpdateParams struct {
	PlayerNextSkill   func() *SkillInstance
	OpponentNextSkill func() *SkillInstance
}

func (b *Battle) Update(ctx *game.Context, s *State, timeDelta float64, params BattleUpdateParams) {
	ctx.DebugBR(fmt.Sprintf("PendingProgress: %f", b.PendingProgress))
	ctx.DebugBR(fmt.Sprintf("TickPlayerNext: %t", b.TickPlayerNext))
	ctx.DebugBR(fmt.Sprintf("PlayerSkill: %s", b.PlayerSkill))
	ctx.DebugBR(fmt.Sprintf("OpponentSkill: %s", b.OpponentSkill))

	if b.PlayerSkill == nil {
		if b.PlayerSkill = params.PlayerNextSkill(); b.PlayerSkill == nil {
			return
		}
	}
	if b.OpponentSkill == nil {
		if b.OpponentSkill = params.OpponentNextSkill(); b.OpponentSkill == nil {
			return
		}
	}

	for b.PendingProgress >= 1 {
		if b.TickPlayerNext {
			dmg, over := b.PlayerSkill.Tick(ctx, s, s.Player.GetCombatant(), s.Opponent)
			if over {
				b.PlayerSkill = nil
			}
			s.emitDamageFx(dmg, false)
		} else {
			dmg, over := b.OpponentSkill.Tick(ctx, s, s.Opponent, s.Player.GetCombatant())
			if over {
				b.OpponentSkill = nil
			}
			s.emitDamageFx(dmg, true)
		}
		b.PendingProgress -= 1.0
		b.TickPlayerNext = !b.TickPlayerNext
	}
	ticksPerSecond := 2.0
	b.PendingProgress += timeDelta * ticksPerSecond
}

func (s *State) emitDamageFx(dmgs []rpg.DamageResult, damagingPlayer bool) {
	if len(dmgs) == 0 {
		return
	}
	pos := pixel.V(game.GameWidth*0.9, game.GameHeight*0.75)
	velocity := pixel.V(-20, rand.Float64()*50+50)
	if damagingPlayer {
		pos.X = game.GameWidth * 0.1
		velocity.X = velocity.X * -1
	}
	for _, dmg := range dmgs {
		s.fx = append(s.fx, &DamageFX{
			Damage:   dmg,
			Position: pos,
			Velocity: velocity,
		})
	}
}

func (i *SkillInstance) Tick(ctx *game.Context, s *State, source Combatant, target Combatant) ([]rpg.DamageResult, bool) {
	log.Info().Msgf("%s: %s.Tick(%d:%d)", source.Name(), i, i.NextTick, i.Duration)
	var dmg []rpg.DamageResult
	if i.NextTick <= i.Duration {
		dmg = i.OnTick(ctx, s, i.NextTick, source, target)
		i.NextTick++
	}
	return dmg, i.NextTick > i.Duration
}
