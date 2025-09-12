package combat

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var tickBubbleDisplayNone = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 1, 1)
var tickBubbleDisplayNormal = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 2, 1)
var tickBubbleDisplaySpecial = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 3, 1)
var tickBubbleDisplayVBar = atlas.GetTilesheetSprite("combat/tick_bar/bubbles", 4, 1)

var tickBarStanceBoxActive = atlas.GetSprite("combat/tick_bar/skill_active_stance_box")
var tickBarStanceBoxPending = atlas.GetSprite("combat/tick_bar/skill_pending_stance_box")

var stanceIcons = map[rpg.CombatStance]pixelutil.BoundedDrawable{
	rpg.TickStanceDefending:  atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 3),
	rpg.TickStanceReflecting: atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 1),
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
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Player, false)
	s.drawCombatantSkills(ctx, target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Opponent, true)
	skillEaterSprite.Draw(target, matrixTopMiddle)
}

var baseNextSkillMaskScale = 0.8
var nextSkillFlashRation = 0.2

func (s *State) drawCombatantSkills(ctx *game.Context, target pixel.Target, matrixTopMiddle pixel.Matrix, currentTickProgress float64, combatant Combatant, flip bool) {
	nextSkillMaskScale := baseNextSkillMaskScale*(1-nextSkillFlashRation) + baseNextSkillMaskScale*nextSkillFlashRation*s.skillFlashAlpha
	noNextSkillAlpha := 1.0
	currentSkill := combatant.GetCurrentSkill()
	nextSkillId := combatant.PeekNextSkill()
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
		mask := pixel.RGBA{1, 1, 1, 1}
		alpha := math.Min((skillProgress)/1, 1)*(1-nextSkillMaskScale) + nextSkillMaskScale
		mask = colors.ScaleColor(mask, alpha)
		s.drawSkill(ctx, target, matrixTopMiddle, currentSkill.Skill, mask, skillProgress > 0.5, 1.0, skillProgress, flip)
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -float64((currentSkill.Duration+1)*skillBarTickSpacing)))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick-1)+currentTickProgress)/float64(currentSkill.Duration)) // 100% by 1 tick away
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := pixel.RGBA{1, 1, 1, 1}
		if !combatant.IsNextSkillCommitted() {
			mask = colors.ScaleColor(mask, 0.5)
		}
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
			artificialSkillProgress := skillProgress + 0.25 // artificial progress to preempt overlay sprites above
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
