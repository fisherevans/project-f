package combat

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/shaders"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"

	"image/color"
	"math"
)

// Things to add
// - enhance damage FX for various conditions (i.e. effective, immune, etc.)
// - add skill log
//   - add data structure to hold it
//   - add rendering
// - add arrow + select popup to view skill details
// - add health delay (interp to target) - allow temp death
//   - render target in bar
// - state stuff
//   - add transition state into combat (i.e. fade)
//   - add battle intro that has creature enter
//   - add initiative roll (maybe add ticks of confusion when starting)
//   - add battle menu (fight, run, item, etc.)
//   - add battle end (win, lose)
// - generate experience at battle end (include in end summary)
// - combat abilities
//   - blocking mechanism
//   - slow down tempo (make it smooth)
// - add simple skill animations for effects to make it easier to understand what's happening
//   - make sprite hop, like pokemon
//   - add some sprite for fire, etc.
// - add damage animations (i.e. red flash)
// - sound
//   - add damage noise (FIRST ONE)
//   - add background music
//   - add win vs lose chime
// - add more interesting AI (more skills, show next skill after time)

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

	overlay        string
	overlayElapsed float64

	batch *pixel.Batch
}

func New(animech *rpg.DeployedAnimech, onComplete OnComplete) *State {
	return &State{
		Player: NewPlayer(animech),
		Opponent: &Wall{
			Health: NewHealthState(100),
			Tempo:  &Tempo{},
		},
		OnComplete: onComplete,
		Battle:     &Battle{},

		cachedContents: map[string]*textbox.Content{},

		batch: atlas.NewBatch(),
	}
}

type HealthState struct {
	Max        int
	Target     int
	Current    float64
	AdjustRate float64
}

func NewHealthState(max int) *HealthState {
	return &HealthState{
		Max:        max,
		Current:    float64(max),
		Target:     max,
		AdjustRate: 0.1,
	}
}

func (h *HealthState) GetCurrentInt() int {
	return int(math.Round(h.Current))
}

func (h *HealthState) AdjustTarget(amount int) int {
	h.Target += amount
	if h.Target < 0 {
		remainder := h.Target
		h.Target = 0
		return remainder
	}
	if h.Target > h.Max {
		remainder := h.Target - h.Max
		h.Target = h.Max
		return remainder
	}
	return 0
}

func (h *HealthState) Update(timeDelta float64) {
	if math.Round(h.Current) == float64(h.Target) {
		return
	}
	sign := 1.0
	if h.Current > float64(h.Target) {
		sign = -1
	}
	diff := h.AdjustRate * timeDelta * float64(h.Max)
	maxDiff := math.Abs(float64(h.Target) - h.Current)
	diff = math.Min(diff, maxDiff)
	h.Current += sign * diff
}

type CombatantStats struct {
	BodyType        rpg.BodyType
	Affinities      []rpg.SkillType
	PhysicalAttack  int
	PhysicalDefense int
	AetherAttack    int
	AetherDefense   int
}

type Combatant interface {
	Name() string
	ApplyDamage(damage rpg.DamageResult)
	GetStats() CombatantStats
	GetTempo() *Tempo
	PopNextSkill() *rpg.SkillId
	PeekNextSkill() *rpg.SkillId
}

func (s *State) ClearColor() color.Color {
	return color.Black
}

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

var robotAnim = anim.IdleRobot(atlas)
var plentAnim = anim.IdlePlent(atlas)

func (s *State) OnTick(ctx *game.Context, target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
	s.batch.Clear()

	backgroundSprite.DrawColorMask(target, pixel.IM.Moved(targetBounds.Center()), colors.Grey4.RGBA)

	if s.overlay == "" {
		s.Opponent.GetHealth().Update(timeDelta)
		s.Player.GetCurrentShield().Update(timeDelta)
		s.Player.GetCurrentSync().Update(timeDelta)

		s.Battle.Update(ctx, s, timeDelta, BattleUpdateParams{
			PlayerNextSkill: func() *rpg.SkillId {
				return s.Player.PopNextSkill()
			},
			OpponentNextSkill: func() *rpg.SkillId {
				return s.Opponent.PopNextSkill()
			},
		})
	}

	if s.Player.GetCurrentSync().GetCurrentInt() <= 0 {
		s.overlay = "{+c:#e64565,+o}You died!"
		if s.overlayElapsed > 5 {
			// todo
			s.OnComplete(ctx, s)
			return
		}
	}

	if s.Opponent.GetHealth().GetCurrentInt() <= 0 {
		s.overlay = "{+c:#45e682,+o}You won!"
		if s.overlayElapsed > 5 {
			// todo
			s.OnComplete(ctx, s)
			return
		}
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

	s.renderSkills(ctx, s.batch, targetBounds, timeDelta)

	s.drawPlayerStats(ctx)
	s.drawOpponentStats(ctx)

	s.Player.Tempo.Render(ctx, s.batch, pixel.IM.Moved(pixel.V(8, game.GameHeight*0.675)))

	if s.overlay != "" {
		s.overlayElapsed += timeDelta
		content := combatantNameText.NewComplexContent(s.overlay)
		combatantNameText.Render(ctx, s.batch, pixel.IM.Moved(pixel.V(game.GameWidth/2, game.GameHeight*0.75)), content, tbcfg.RenderFrom(gfx.Centered))
	}

	s.batch.Draw(target)
}

var tickBubbleDisplayNone = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 1, 1)
var tickBubbleDisplayNormal = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 2, 1)
var tickBubbleDisplaySpecial = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 3, 1)
var tickBubbleDisplayVBar = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 4, 1)

var tickBarStanceBoxActive = atlas.GetSprite("combat/tick_bar/skill_active_stance_box")
var tickBarStanceBoxPending = atlas.GetSprite("combat/tick_bar/skill_pending_stance_box")

var stanceIcons = map[rpg.TickStanceType]pixelutil.BoundedDrawable{
	rpg.TickStanceDefending:  atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 3),
	rpg.TickStanceReflecting: atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 1),
	rpg.TickStanceReposing:   atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 5),
	rpg.TickStanceVulnerable: atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 5, 1),
	rpg.TickStanceExposed:    atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 7, 3),
}

var skillBarWidth = 8
var skillBarTickSpacing = 12
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
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Battle.PlayerSkill, s.Player.PeekNextSkill(), false)
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Battle.OpponentSkill, s.Opponent.PeekNextSkill(), true)
	skillEaterSprite.Draw(target, matrixTopMiddle)
}

var baseNextSkillMaskScale = 0.8
var nextSkillFlashRation = 0.2

func (s *State) drawCombatantSkills(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, currentTickProgress float64, currentSkill *SkillInstance, nextSkillId *rpg.SkillId, flip bool) {
	nextSkillMaskScale := baseNextSkillMaskScale*(1-nextSkillFlashRation) + baseNextSkillMaskScale*nextSkillFlashRation*s.skillFlashAlpha
	noNextSkillAlpha := 1.0
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (currentTickProgress-0.5)*float64(skillBarTickSpacing)))
	if currentSkill != nil {
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, (float64(currentSkill.NextTick))*float64(skillBarTickSpacing)))
	}
	nextSkillProgress := 0.0
	if currentSkill != nil {
		skillProgress := currentTickProgress + float64(currentSkill.NextTick)
		skillProgressRemaining := float64(currentSkill.Duration) + 1.5 - skillProgress // i honestly don't know why I need to add 1.5
		if skillProgressRemaining < 0.5 {
			nextSkillProgress = 0.5 - skillProgressRemaining
		}
		mask := colors.OfSkillType(currentSkill.Skill.Type).RGBA
		alpha := math.Min((skillProgress)/1, 1)*(1-nextSkillMaskScale) + nextSkillMaskScale
		mask = colors.ScaleColor(mask, alpha)
		s.drawSkill(ctx, target, matrixTopMiddle, currentSkill.Skill, mask, skillProgress > 0.5, 1.0, skillProgress, flip)
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -float64((currentSkill.Duration+1)*skillBarTickSpacing)))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick-1)+currentTickProgress)/float64(currentSkill.Duration)) // 100% by 1 tick away
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := colors.OfSkillType(nextSkill.Type).RGBA
		mask = colors.ScaleColor(mask, nextSkillMaskScale)
		s.drawSkill(ctx, target, matrixTopMiddle, &nextSkill, mask, false, 1.0, nextSkillProgress, flip)
	} else {
		y := noneSelectedSprite.Bounds().H() / 2
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		noneSelectedSprite.DrawColorMask(target, matrixTopMiddle.Moved(pixel.V(0, -y)), pixel.RGBA{noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha, noNextSkillAlpha})
	}
}

func (s *State) drawSkill(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, skill *rpg.Skill, mask pixel.RGBA, active bool, alpha float64, skillProgress float64, flip bool) {
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
		lastStance := rpg.TickStanceNone
		if i > 0 {
			lastStance = skill.Ticks[i-1].StanceType
		}
		tick := skill.Ticks[i]
		inStance := tick.StanceType != rpg.TickStanceNone

		tickSpriteCenterMatrix := matrixBottomLeft.Moved(pixel.V(float64(skillBarWidth/2), float64(skillBarTickSpacing/2+skillBarTickSpacing*(skill.Duration()-i))))

		// if we're entering a stance, draw the full VBar, box, and icon
		if tick.StanceType != lastStance && tick.StanceType != rpg.TickStanceNone {
			// draw the VBar
			stanceDuration := 0
			for j := i + 1; j <= skill.Duration(); j++ {
				if skill.Ticks[j].StanceType != tick.StanceType {
					break
				}
				stanceDuration++
			}
			vbarSprite := tickBubbleDisplayVBar
			vbarVerticalMargin := 1
			height := (skillBarTickSpacing * stanceDuration) + int(vbarSprite.Bounds().H()) - vbarVerticalMargin*2
			scale := float64(height) / vbarSprite.Bounds().H()
			vbarSprite.DrawColorMask(target, pixel.IM.ScaledXY(pixel.V(0, vbarSprite.Bounds().H()/2), pixel.V(1, scale)).Moved(pixel.V(0, float64(-vbarVerticalMargin))).Chained(tickSpriteCenterMatrix), mask)

			// get box sprite
			boxSprite := tickBarStanceBoxPending
			if active {
				boxSprite = tickBarStanceBoxActive
			}

			stanceDelta := pixel.ZV
			artificialSkillProgress := skillProgress + 0.5 // artificial progress to preempt overlay sprites above
			if artificialSkillProgress > float64(i) {
				stanceProgress := math.Min(artificialSkillProgress-float64(i), float64(stanceDuration))
				dy := float64(skillBarTickSpacing) * math.Max(0, stanceProgress-1)
				dy = math.Min(dy, float64(height))
				stanceDelta = pixel.V(0, -dy)
			}

			// draw box sprite
			stanceBoxMatrix := pixel.IM
			if flip {
				stanceBoxMatrix = stanceBoxMatrix.ScaledXY(pixel.V(0, 0), pixel.V(-1, 1))
			}
			stanceBoxMatrix = stanceBoxMatrix.Moved(pixel.V(0, -5.5))
			stanceBoxMatrix = stanceBoxMatrix.Moved(stanceDelta)
			stanceBoxMatrix = stanceBoxMatrix.Chained(tickSpriteCenterMatrix)
			boxSprite.DrawColorMask(target, stanceBoxMatrix, mask)

			// draw the icon
			if stanceIconSprite, exists := stanceIcons[tick.StanceType]; exists {
				iconMatrix := pixel.IM
				if flip {
					iconMatrix = iconMatrix.ScaledXY(pixel.V(0, 0), pixel.V(-1, 1))
					iconMatrix = iconMatrix.Moved(pixel.V(1.5, -5.5))
				} else {
					iconMatrix = iconMatrix.Moved(pixel.V(-1.5, -5.5))
				}
				iconMatrix = iconMatrix.Moved(stanceDelta)
				stanceIconSprite.DrawColorMask(target, iconMatrix.Chained(tickSpriteCenterMatrix), mask)
			}
		}

		// draw dots for each tick
		var sprite pixelutil.BoundedDrawable
		switch tick.DisplayType {
		case rpg.TickDisplayDamage:
			sprite = tickBubbleDisplayNormal
		case rpg.TickDisplayNone:
			// only render "none" dots if we're not in a stance
			if !inStance {
				sprite = tickBubbleDisplayNone
			}
		}
		if sprite != nil {
			sprite.DrawColorMask(target, tickSpriteCenterMatrix, mask)
		}
	}
}
