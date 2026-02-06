package combat

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

var skillEaterSprite pixelutil.BoundedDrawable
var tickBarRenderer *tick_bar.Renderer

var (
	tickBarWidth = 8
	tickSpacing  = 12
)

var skillBarWidth = 8
var skillBarTickSpacing = 12
var skillBarSpacing = 2

func (s *State) drawActiveSkills(target pixel.Target, targetBounds pixel.Rect, matrixTopMiddle pixel.Matrix) {
	playerProgress := s.Battle.PendingProgress / 2.0
	opponentProgress := playerProgress
	if s.Battle.TickPlayerNext {
		playerProgress += 0.5
	} else {
		opponentProgress += 0.5
	}
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -skillEaterSprite.Bounds().H()/2))

	dy := s.visibilityActiveSkills.GetInvisibleAmount() * 10
	matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, dy))
	mask := colors.Alpha(s.visibilityActiveSkills.GetVisibleAmount())

	s.drawCombatantSkills(target, matrixTopMiddle.Moved(pixel.V(-float64(skillBarSpacing/2+skillBarWidth/2), 0)), playerProgress, s.Player, false, mask)
	s.drawCombatantSkills(target, matrixTopMiddle.Moved(pixel.V(float64(skillBarSpacing/2+skillBarWidth/2), 0)), opponentProgress, s.Opponent, true, mask)
	skillEaterSprite.DrawColorMask(target, matrixTopMiddle, mask)
}

var baseNextSkillMaskScale = 0.8
var nextSkillFlashRation = 0.2

func (s *State) drawCombatantSkills(target pixel.Target, matrixTopMiddle pixel.Matrix, currentTickProgress float64, combatant Combatant, flip bool, visMask pixel.RGBA) {
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
		skillProgressRemaining := float64(currentSkill.Duration()) + 1.5 - skillProgress // i honestly don't know why I need to add 1.5
		if skillProgressRemaining < 0.5 {
			nextSkillProgress = 0.5 - skillProgressRemaining
		}
		mask := visMask
		alpha := math.Min((skillProgress)/1, 1)*(1-nextSkillMaskScale) + nextSkillMaskScale
		mask = colors.ScaleColor(mask, alpha)
		tickBarOpts := tick_bar.NewDrawOptions(tickBarWidth, tickSpacing).
			InterruptedAt(currentSkill.InterruptedAt).
			Mask(mask).
			Active(skillProgress > 0.5).
			Alpha(alpha).
			SkillProgress(skillProgress).
			Flip(flip)
		tickBarRenderer.Draw(target, matrixTopMiddle, currentSkill.Skill, tickBarOpts)
		//drawSkill(target, matrixTopMiddle, currentSkill.Skill, currentSkill.InterruptedAt, mask, skillProgress > 0.5, 1.0, skillProgress, flip)
		matrixTopMiddle = matrixTopMiddle.Moved(pixel.V(0, -float64((currentSkill.Duration()+1)*skillBarTickSpacing)))
		noNextSkillAlpha = math.Min(1.0, (float64(currentSkill.NextTick-1)+currentTickProgress)/float64(currentSkill.Duration())) // 100% by 1 tick away
	}
	if nextSkillId != nil {
		nextSkill := nextSkillId.Get()
		mask := visMask
		mask = colors.ScaleColor(mask, 0.95)
		if !combatant.IsNextSkillCommitted() {
			mask = colors.ScaleColor(mask, 0.9)
		}
		mask = colors.ScaleColor(mask, nextSkillMaskScale)
		tickBarOpts := tick_bar.NewDrawOptions(tickBarWidth, tickSpacing).
			InterruptedAt(-1).
			Mask(mask).
			Active(combatant.IsNextSkillCommitted()).
			Alpha(1.0).
			SkillProgress(nextSkillProgress).
			Flip(flip)
		tickBarRenderer.Draw(target, matrixTopMiddle, &nextSkill, tickBarOpts)
		//drawSkill(target, matrixTopMiddle, &nextSkill, -1, mask, combatant.IsNextSkillCommitted(), 1.0, nextSkillProgress, flip)
	} else {
		y := noneSelectedSprite.Bounds().H() / 2
		noNextSkillAlpha *= s.skillFlashAlphaInverse
		mask := colors.LayerAlpha(visMask, noNextSkillAlpha)
		noneSelectedSprite.DrawColorMask(target, matrixTopMiddle.Moved(pixel.V(0, -y)), mask)
	}
}
