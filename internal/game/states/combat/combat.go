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
	Foreground(colors.White.RGBA).
	RenderFrom(textbox.TopLeft))

var backgroundSprite = resources.GetNonAtlasSprite("combat/background_sample").Sprite

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
			return &rpg.Skill_Crush.Id
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

	backgroundSprite.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Grey4.RGBA)

	robotAnim.Update(timeDelta)
	robotAnim.Sprite().Draw(target, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.15), math.Floor(game.GameHeight*0.566))))

	plentAnim.Update(timeDelta)
	plentAnim.Sprite().Draw(target, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.85), math.Floor(game.GameHeight*0.6667))))

	// while time left
	//  - progress time
	//  - if skill ends, pop next

	s.drawActiveSkills(ctx, target, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))

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

	statFrameW, statFrameH := 78, 31

	nameBoxSprite := resources.GetTilesheetSprite("combat/combatant_stats/background", 1, 1)
	rightSprite := resources.GetTilesheetSprite("combat/combatant_stats/background", 2, 1)
	bottomSprite := resources.GetTilesheetSprite("combat/combatant_stats/background", 3, 1)

	nameBoxHeight := combatantNameText.GetLetterHeight() + namePadding*2 + statFrameH - int(bottomSprite.Bounds.H())
	nameBoxWidth := nameContent.Width() + namePadding*2
	ctx.DebugBR("nameBoxHeight: %d, nameBoxWidth: %d", nameBoxHeight, nameBoxWidth)
	nameBoxSpriteScaleX := float64(nameBoxWidth) / nameBoxSprite.Bounds.W()
	nameBoxSpriteScaleY := float64(nameBoxHeight) / nameBoxSprite.Bounds.H()
	ctx.DebugBR("nameBoxSpriteScaleX: %f, nameBoxSpriteScaleY: %f", nameBoxSpriteScaleX, nameBoxSpriteScaleY)
	nameBoxSprite.Sprite.Draw(target, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(nameBoxSpriteScaleX, nameBoxSpriteScaleY)).
		Moved(pixel.V(float64(nameBoxWidth/2), float64(game.GameHeight-nameBoxHeight/2))))

	rightSprite.Sprite.Draw(target, pixel.IM.Moved(rightSprite.HalfDimensions()).Moved(pixel.V(float64(nameBoxWidth-1), game.GameHeight-rightSprite.Bounds.H())))

	bottomSprite.Sprite.Draw(target, pixel.IM.Moved(bottomSprite.HalfDimensions()).Moved(pixel.V(0, game.GameHeight-float64(nameBoxHeight)-bottomSprite.Bounds.H())))

	combatantNameText.Render(ctx, target, pixel.IM.Moved(pixel.V(float64(namePadding), float64(game.GameHeight-namePadding+combatantNameText.GetTailHeight()))), nameContent)

	statFrameBLX, statFrameBLY := namePadding, game.GameHeight-(namePadding*2+combatantNameText.GetLetterHeight()+statFrameH)
	frames.Draw(target, combatStatFrame, pixel.R(0, 0, float64(statFrameW), float64(statFrameH)),
		pixel.IM.Moved(pixel.V(float64(statFrameBLX), float64(statFrameBLY))))

	maxBarWidth := statFrameW - 2*statBarFrame.HorizontalPadding()
	var syncBarWidth, shldBarWidth int
	if s.Player.GetCombatant().GetCurrentSync().Max > s.Player.GetCombatant().GetCurrentShield().Max {
		syncBarWidth = maxBarWidth
		shldBarWidth = int(float64(maxBarWidth) * float64(s.Player.GetCombatant().GetCurrentShield().Max) / float64(s.Player.GetCombatant().GetCurrentSync().Max))
	} else {
		shldBarWidth = maxBarWidth
		syncBarWidth = int(float64(maxBarWidth) * float64(s.Player.GetCombatant().GetCurrentSync().Max) / float64(s.Player.GetCombatant().GetCurrentShield().Max))
	}

	syncBar := &StatBar{
		lines:       []StatBarLine{StatBarVisual, StatBarLabel},
		labelSprite: resources.GetTilesheetSprite("combat/combatant_stats/background", 6, 1),
		colorDark:   colors.HexColor("772712"),
		color:       colors.HexColor("d58d7a"),
		colorBright: colors.HexColor("ebc4bb"),
		current:     s.Player.GetCombatant().GetCurrentSync().Current,
		max:         s.Player.GetCombatant().GetCurrentSync().Max,
		width:       syncBarWidth,
	}

	shieldBar := &StatBar{
		lines:       []StatBarLine{StatBarLabel, StatBarVisual},
		labelSprite: resources.GetTilesheetSprite("combat/combatant_stats/background", 5, 1),
		colorDark:   colors.HexColor("126177"),
		color:       colors.HexColor("73bed3"),
		colorBright: colors.HexColor("bbe0eb"),
		current:     s.Player.GetCombatant().GetCurrentShield().Current,
		max:         s.Player.GetCombatant().GetCurrentShield().Max,
		width:       shldBarWidth,
	}

	statBox := &StatBox{
		bars:           []*StatBar{shieldBar, syncBar},
		originLocation: StatBoxOriginTopLeft,
	}
	statBox.Draw(ctx, target, pixel.IM.Moved(pixel.V(float64(namePadding), float64(game.GameHeight-(namePadding*2+combatantNameText.GetLetterHeight())))))

}

var tickBubbleDisplayNone = resources.GetTilesheetSprite("combat/tick_bar/bubbles", 1, 1)
var tickBubbleDisplayNormal = resources.GetTilesheetSprite("combat/tick_bar/bubbles", 2, 1)
var tickBubbleDisplaySpecial = resources.GetTilesheetSprite("combat/tick_bar/bubbles", 3, 1)

var skillBarWidth = 8
var skillBarTickSpacing = 10
var skillBarSpacing = 2

var skillEaterSprite = resources.GetSprite("combat/tick_bar/skill_eater")

func (s *State) drawActiveSkills(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, matrixTopMiddle pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -skillEaterSprite.Bounds.H()/2))
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Battle.PreviousPlayerSkill, s.Battle.PlayerSkill, s.Player.NextSkill)
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Battle.PreviousOpponentSkill, s.Battle.OpponentSkill, nil)
	skillEaterSprite.Sprite.Draw(target, matrixTopMiddle)
}

var baseNextSkillMaskScale = 0.8
var nextSkillFlashRation = 0.2

func (s *State) drawCombatantSkills(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, currentTickProgress float64, previousSkill *SkillInstance, currentSkill *SkillInstance, nextSkillId *rpg.SkillId) {
	nextSkillMaskScale := baseNextSkillMaskScale*(1-nextSkillFlashRation) + baseNextSkillMaskScale*nextSkillFlashRation*s.skillFlashAlpha
	noNextSkillAlpha := 1.0
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (currentTickProgress-0.5)*float64(skillBarTickSpacing)))
	if currentSkill != nil {
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (float64(currentSkill.NextTick))*float64(skillBarTickSpacing)))
	}
	if previousSkill != nil {
		mask := typecolors.SkillTypeColor(previousSkill.Skill.Type).RGBA
		matrixPreviousTopMiddle := matrixTopMiddle.Moved(pixel.V(0, float64((previousSkill.Duration+1)*skillBarTickSpacing)))
		s.drawSkill(ctx, target, matrixPreviousTopMiddle, previousSkill.Skill, mask, true, 1.0)
		ctx.DebugBR("printing previous")
	}
	if currentSkill != nil {
		skillProgress := currentTickProgress + float64(currentSkill.NextTick)
		mask := typecolors.SkillTypeColor(currentSkill.Skill.Type).RGBA
		alpha := math.Min((skillProgress)/1, 1)*(1-nextSkillMaskScale) + nextSkillMaskScale
		ctx.DebugBL("alpha: %f", alpha)
		mask = colors.ScaleColor(mask, alpha)
		s.drawSkill(ctx, target, matrixTopMiddle, currentSkill.Skill, mask, skillProgress > 0.5, 1.0)
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -float64((currentSkill.Duration+1)*skillBarTickSpacing)))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick-1)+currentTickProgress)/float64(currentSkill.Duration)) // 100% by 1 tick away
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := typecolors.SkillTypeColor(nextSkill.Type).RGBA
		mask = colors.ScaleColor(mask, nextSkillMaskScale)
		s.drawSkill(ctx, target, matrixTopMiddle, &nextSkill, mask, false, 1.0)
	} else {
		ctx.DebugBR("no skill!")
		y := noneSelectedSprite.Bounds.H() / 2
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		noneSelectedSprite.Sprite.DrawColorMask(target, matrixTopMiddle.Moved(pixel.V(0, -y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
	}
}

func (s *State) drawSkill(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, skill *rpg.Skill, mask pixel.RGBA, active bool, alpha float64) {
	if skill == nil {
		return
	}
	mask = colors.WithAlpha(mask, alpha)
	rect := pixel.R(0, 0,
		float64(skillBarWidth),
		float64(skillBarTickSpacing*(skill.Duration()+1)))
	matrixBottomLeft := matrixTopMiddle.Moved(pixel.V(-rect.W()/2, -rect.H()))
	frameName := "combat/tick_bar/skill_pending_frame"
	if active {
		frameName = "combat/tick_bar/skill_active_frame"
	}
	frame := resources.GetFrame(frameName)
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

	tps := ticksPerSecond
	if ctx.Toggles.F1().ToggleState() {
		tps *= 4
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
