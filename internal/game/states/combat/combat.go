package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/colors/typecolors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/textbox"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"image/color"
	"math"
	"math/rand"
	"strings"
)

var ticksPerSecond = 1.5

type OnComplete func(ctx *game.Context, s *State)

type State struct {
	Player     *Player
	Opponent   Opponent
	OnComplete OnComplete
	Battle     *Battle

	fx []FX

	combatArrowAlpha       float64
	combatArrowColumn      int
	cachedContents         map[string]*textbox.Content
	skillFlashTimeElapsed  float64
	skillFlashAlpha        float64
	skillFlashAlphaInverse float64
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

		cachedContents: map[string]*textbox.Content{},
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

var robotAnim = anim.IdleRobot()
var plentAnim = anim.IdlePlent()

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
		PlayerNextSkill: func() *rpg.SkillId {
			return s.Player.PopNextSkill()

		},
		OpponentNextSkill: func() *rpg.SkillId {
			return &rpg.Skill_Tackle.Id
		},
	})

	if s.Player.GetCombatant().GetCurrentSync().Current <= 0 {
		ctx.Notify("You died!")
		s.OnComplete(ctx, s)
		return
	}

	if s.Opponent.GetHealth().Current <= 0 {
		ctx.Notify("You won!")
		s.OnComplete(ctx, s)
		return
	}

	var remainingFx []FX
	for _, fx := range s.fx {
		if !fx.Update(ctx, s, timeDelta) {
			remainingFx = append(remainingFx, fx)
		}
	}
	s.fx = remainingFx

	backgroundSprite.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Grey4)

	robotAnim.Update(timeDelta)
	robotAnim.Sprite().Draw(target, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.15), math.Floor(game.GameHeight*0.6667))))

	plentAnim.Update(timeDelta)
	plentAnim.Sprite().Draw(target, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.85), math.Floor(game.GameHeight*0.6667))))

	// while time left
	//  - progress time
	//  - if skill ends, pop next

	s.drawActiveSkillsV(ctx, target, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))

	playerText := []string{
		"Player",
		s.Player.GetCombatant().Name(),
		fmt.Sprintf("Health: %d/%d", s.Player.GetCombatant().GetCurrentSync().Current, s.Player.GetCombatant().GetCurrentSync().Max),
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

	s.renderSkills(ctx, target, pixel.V(float64(3), float64(3)), timeDelta)
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

var skillBarWidth = 8
var skillBarTickSpacing = 10
var skillBarSpacing = 2

func (s *State) drawActiveSkills(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, matrix pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	s.drawCombatantSkills(ctx, target, matrix.Moved(pixel.V(0, float64(skillBarSpacing/2))), opponentProgress, s.Battle.OpponentSkill, nil)
	s.drawCombatantSkills(ctx, target, matrix.Moved(pixel.V(0, float64(-skillBarWidth-skillBarSpacing/2))), playerProgress, s.Battle.PlayerSkill, s.Player.NextSkill)
	matrix = matrix.Moved(pixel.V(float64(-skillBarTickSpacing/2), 0))
	resources.Sprites["combat_tick_eater"].Sprite.Draw(target, matrix)
}

func (s *State) drawCombatantSkills(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, currentTickProgress float64, currentSkill *SkillInstance, nextSkillId *rpg.SkillId) {
	noNextSkillAlpha := 1.0
	matrix = matrix.Moved(pixel.V(-currentTickProgress*float64(skillBarTickSpacing), 0))
	if currentSkill != nil {
		matrix = matrix.Moved(pixel.V(-(float64(currentSkill.NextTick))*float64(skillBarTickSpacing), 0))
		mask := typecolors.SkillTypeColor(currentSkill.Skill.Type)
		s.drawSkill(ctx, target, matrix, currentSkill.Skill, mask, 1.0)
		matrix = matrix.Moved(pixel.V(float64((currentSkill.Duration+1)*skillBarTickSpacing), 0))
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
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		noneSelectedSprite.Sprite.DrawColorMask(target, matrix.Moved(pixel.V(x, y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
	}
}

func (s *State) drawSkill(ctx *game.Context, target pixel.Target, matrix pixel.Matrix, skill *rpg.Skill, mask pixel.RGBA, f float64) {
	if skill == nil {
		return
	}
	mask = colors.WithAlpha(mask, f)
	rect := pixel.R(0, 0,
		float64(skillBarTickSpacing+skillBarTickSpacing*(skill.Duration())),
		float64(skillBarWidth))
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
		sprite.DrawColorMask(target, matrix.Moved(pixel.V(float64(skillBarTickSpacing/2+skillBarTickSpacing*i), float64(skillBarWidth/2))), mask)
	}
}

var skillEaterSprite = resources.Sprites["combat_skill_eater_vert"]

func (s *State) drawActiveSkillsV(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, matrixTopMiddle pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -skillEaterSprite.Bounds.H()/2))
	s.drawCombatantSkillsV(ctx, target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Battle.PreviousPlayerSkill, s.Battle.PlayerSkill, s.Player.NextSkill)
	s.drawCombatantSkillsV(ctx, target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Battle.PreviousOpponentSkill, s.Battle.OpponentSkill, nil)
	skillEaterSprite.Sprite.Draw(target, matrixTopMiddle)
}

var baseNextSkillMaskScale = 0.8
var nextSkillFlashRation = 0.2

func (s *State) drawCombatantSkillsV(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, currentTickProgress float64, previousSkill *SkillInstance, currentSkill *SkillInstance, nextSkillId *rpg.SkillId) {
	nextSkillMaskScale := baseNextSkillMaskScale*(1-nextSkillFlashRation) + baseNextSkillMaskScale*nextSkillFlashRation*s.skillFlashAlpha
	noNextSkillAlpha := 1.0
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (currentTickProgress-0.5)*float64(skillBarTickSpacing)))
	if currentSkill != nil {
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (float64(currentSkill.NextTick))*float64(skillBarTickSpacing)))
	}
	if previousSkill != nil {
		mask := typecolors.SkillTypeColor(previousSkill.Skill.Type)
		matrixPreviousTopMiddle := matrixTopMiddle.Moved(pixel.V(0, float64((previousSkill.Duration+1)*skillBarTickSpacing)))
		s.drawSkillV(ctx, target, matrixPreviousTopMiddle, previousSkill.Skill, mask, true, 1.0)
		ctx.DebugBR("printing previous")
	}
	if currentSkill != nil {
		skillProgress := currentTickProgress + float64(currentSkill.NextTick)
		mask := typecolors.SkillTypeColor(currentSkill.Skill.Type)
		alpha := math.Min((skillProgress)/1, 1)*(1-nextSkillMaskScale) + nextSkillMaskScale
		ctx.DebugBL("alpha: %f", alpha)
		mask = colors.ScaleColor(mask, alpha)
		s.drawSkillV(ctx, target, matrixTopMiddle, currentSkill.Skill, mask, skillProgress > 0.5, 1.0)
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -float64((currentSkill.Duration+1)*skillBarTickSpacing)))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick-1)+currentTickProgress)/float64(currentSkill.Duration)) // 100% by 1 tick away
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := typecolors.SkillTypeColor(nextSkill.Type)
		mask = colors.ScaleColor(mask, nextSkillMaskScale)
		s.drawSkillV(ctx, target, matrixTopMiddle, &nextSkill, mask, false, 1.0)
	} else {
		ctx.DebugBR("no skill!")
		y := noneSelectedSprite.Bounds.H() / 2
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		noneSelectedSprite.Sprite.DrawColorMask(target, matrixTopMiddle.Moved(pixel.V(0, -y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
	}
}

func (s *State) drawSkillV(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, skill *rpg.Skill, mask pixel.RGBA, active bool, alpha float64) {
	if skill == nil {
		return
	}
	mask = colors.WithAlpha(mask, alpha)
	rect := pixel.R(0, 0,
		float64(skillBarWidth),
		float64(skillBarTickSpacing*(skill.Duration()+1)))
	matrixBottomLeft := matrixTopMiddle.Moved(pixel.V(-rect.W()/2, -rect.H()))
	frameName := "combat_tick_frame"
	if active {
		frameName = "combat_tick_active_frame"
	}
	frame := resources.Frames[frameName]
	frames.Draw(target, frame, rect, matrixBottomLeft, frames.WithColor(mask))

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
		sprite.DrawColorMask(target, matrixBottomLeft.Moved(pixel.V(float64(skillBarWidth/2), float64(skillBarTickSpacing/2+skillBarTickSpacing*(skill.Duration()-i)))), mask)
	}
}

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
		return
	}
	if !b.TickPlayerNext && b.OpponentSkill == nil {
		return
	}

	b.PendingProgress += timeDelta * ticksPerSecond
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
