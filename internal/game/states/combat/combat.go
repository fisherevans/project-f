package combat

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/colors/typecolors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fmt"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"image/color"
	"math"
	"math/rand"
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

	overlay string

	batch *pixel.Batch
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

		batch: atlas.NewBatch(),
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
var stateText = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7), textbox.
	NewConfig((game.GameWidth-padding*3)/2).
	Foreground(colors.White.RGBA).
	RenderFrom(textbox.TopLeft).
	ExtraLineSpacing(2))

var atlas = resources.CreateAtlas(resources.AtlasFilter{
	FontNames: []string{
		resources.FontNameM5x7,
		resources.FontNameM3x6,
		resources.FontNameAddStandard,
		resources.FontNameFF,
		resources.FontName3x5,
	},
})

func init() {
	atlas.Dump("temp", "combat")
}

var backgroundSprite = resources.LoadSprite("combat/background_sample")

var imd = imdraw.New(nil)

var robotAnim = anim.IdleRobot(atlas)
var plentAnim = anim.IdlePlent(atlas)

func (s *State) OnTick(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

	backgroundSprite.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Grey4.RGBA)

	if s.overlay == "" {
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
	}

	if s.Player.GetCombatant().GetCurrentSync().Current <= 0 {
		s.overlay = "{+c:#e64565,+o}You died!"
		// todo
		//s.OnComplete(ctx, s)
		//return
	}

	if s.Opponent.GetHealth().Current <= 0 {
		s.overlay = "{+c:#45e682,+o}You won!"
		// todo
		//s.OnComplete(ctx, s)
		//return
	}

	var remainingFx []FX
	for _, fx := range s.fx {
		if !fx.Update(ctx, s, timeDelta) {
			remainingFx = append(remainingFx, fx)
		}
	}
	s.fx = remainingFx

	robotAnim.Update(timeDelta)
	robotAnim.Sprite().Draw(s.batch, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.15), math.Floor(game.GameHeight*0.566))))

	plentAnim.Update(timeDelta)
	plentAnim.Sprite().Draw(s.batch, pixel.IM.Moved(pixel.V(math.Floor(game.GameWidth*0.85), math.Floor(game.GameHeight*0.6667))))

	// while time left
	//  - progress time
	//  - if skill ends, pop next

	s.drawActiveSkills(ctx, s.batch, targetBounds, pixel.IM.Moved(pixel.V(targetBounds.Center().X, targetBounds.H())))

	for _, fx := range s.fx {
		fx.Render(ctx, s.batch)
	}

	s.renderSkills(ctx, s.batch, pixel.V(float64((game.GameWidth-(skillFrameWidth*2+skillFrameHorizontalSpacing))/2), float64(3)), timeDelta)

	s.drawPlayerStats(ctx)
	s.drawOpponentStats(ctx)

	if s.overlay != "" {
		content := combatantNameText.NewComplexContent(s.overlay)
		combatantNameText.Render(ctx, s.batch, pixel.IM.Moved(pixel.V(game.GameWidth/2, game.GameHeight*0.75)), content, gfx.Centered)
	}

	s.batch.Draw(target)
}

var tickBubbleDisplayNone = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 1, 1)
var tickBubbleDisplayNormal = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 2, 1)
var tickBubbleDisplaySpecial = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 3, 1)

var skillBarWidth = 8
var skillBarTickSpacing = 10
var skillBarSpacing = 2

var skillEaterSprite = atlas.GetSprite("combat/tick_bar/skill_eater")

func (s *State) drawActiveSkills(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, matrixTopMiddle pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -skillEaterSprite.Bounds().H()/2))
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Battle.PreviousPlayerSkill, s.Battle.PlayerSkill, s.Player.NextSkill)
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Battle.PreviousOpponentSkill, s.Battle.OpponentSkill, nil)
	skillEaterSprite.Draw(target, matrixTopMiddle)
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
		y := noneSelectedSprite.Bounds().H() / 2
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		noneSelectedSprite.DrawColorMask(target, matrixTopMiddle.Moved(pixel.V(0, -y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
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
	frames.New(frameName, atlas).Draw(target, rect, matrixBottomLeft, frames.WithColor(mask))

	for i := 0; i <= skill.Duration(); i++ {
		var sprite pixelutil.BoundedDrawable
		switch skill.Ticks[i].DisplayType {
		case rpg.TickDisplayNone:
			sprite = tickBubbleDisplayNone
		case rpg.TickDisplayNormal:
			sprite = tickBubbleDisplayNormal
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
