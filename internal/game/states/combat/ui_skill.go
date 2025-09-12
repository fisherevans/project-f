package combat

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillFrame                  = frames.New("combat/menu/skill_frame", atlas)
	skillPendingFrame           = frames.New("combat/menu/skill_pending_frame", atlas)
	skillFrameWidth             = 84
	skillFrameHeight            = 13
	skillFrameHorizontalSpacing = 26

	skillText = textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6), tbcfg.NewConfig(skillFrameWidth, skillFrameHeight,
		tbcfg.Foreground(colors.Black.RGBA),
		tbcfg.HAligned(tbcfg.HAlignCenter),
		tbcfg.VAligned(tbcfg.VAlignMiddle),
	))

	typeOptionKey = map[int]input.Direction{
		0: input.Up,
		1: input.Right,
		2: input.Down,
		3: input.Left,
	}

	typeOptionKeyReverse = map[input.Direction]int{
		input.Up:    0,
		input.Right: 1,
		input.Down:  2,
		input.Left:  3,
	}

	skillPendingProgress = anim.SkillPendingProgress(atlas)

	skillStatsBadge           = badges.Using(atlas).ButtonAction("select", "stats")
	skillPendingCancelBadge   = badges.Using(atlas).ButtonAction("a", "commit")
	skillCommittedCancelBadge = badges.Using(atlas).ButtonAction("b", "cancel")
	skillMenuBadge            = badges.Using(atlas).ButtonAction("start", "menu")
)

func (s *State) renderSkills(ctx *game.Context, target pixel.Target, targetBounds pixel.Rect, timeDelta float64) {
	bottomLeft := pixel.V(float64((game.GameWidth-(skillFrameWidth*2+skillFrameHorizontalSpacing))/2), 3)

	s.skillFlashTimeElapsed += timeDelta
	sin := (math.Sin(s.skillFlashTimeElapsed*10) + 1.0) / 2.0 // [0-1]
	s.skillFlashAlpha = 0.5 + sin*0.5                         // [0.5-1]
	s.skillFlashAlphaInverse = 0.5 + (1.0-sin)*0.5

	if ctx.Controls.DPad().IsPressed() {
		dir := ctx.Controls.DPad().PressedDirection()
		switch dir {
		case input.Up:
			s.combatArrowAlpha, s.combatArrowColumn = 1, 2
		case input.Right:
			s.combatArrowAlpha, s.combatArrowColumn = 1, 3
		case input.Down:
			s.combatArrowAlpha, s.combatArrowColumn = 1, 4
		case input.Left:
			s.combatArrowAlpha, s.combatArrowColumn = 1, 5
		}
		if ctx.Controls.DPad().JustPressed() {
			selectSkill := s.Player.GetFightOption(typeOptionKeyReverse[dir])
			if selectSkill != s.Player.NextSkill {
				s.Player.NextSkill = s.Player.GetFightOption(typeOptionKeyReverse[dir])
				s.Player.NextSkillCommitted = false
			}
		}
	}

	for optionId := 0; optionId < 4; optionId++ {
		option := s.Player.GetFightOption(optionId)

		text := ""
		optionActive, optionPending := false, false
		if option != nil {
			opt := option.Get()
			text = opt.Name
			if s.Player.GetCurrentSkill() != nil && s.Player.GetCurrentSkill().Skill.Id == opt.Id {
				optionActive = true
			}
			if s.Player.NextSkill != nil && *s.Player.NextSkill == opt.Id {
				optionPending = true
			}
		}
		if optionActive {
			text = "{+u}" + text
		}
		content := s.simpleSkillContent(text)

		matrix := pixel.IM.Moved(bottomLeft)
		rightDx := skillFrameWidth + skillFrameHorizontalSpacing
		switch typeOptionKey[optionId] {
		case input.Up:
			matrix = matrix.Moved(pixel.V(float64(rightDx/2), float64(skillFrameHeight-1)*2))
		case input.Down:
			matrix = matrix.Moved(pixel.V(float64(rightDx/2), 0))
		case input.Right:
			matrix = matrix.Moved(pixel.V(float64(rightDx), float64(skillFrameHeight-1)))
		case input.Left:
			matrix = matrix.Moved(pixel.V(0, float64(skillFrameHeight-1)))
		}
		frame := skillFrame
		if optionActive {
			frame = skillPendingFrame
		}
		frameRect := pixel.R(0, 0, float64(skillFrameWidth), float64(skillFrameHeight))
		frame.Draw(target, frameRect, matrix)
		skillText.Render(ctx, target, matrix, content)
		if optionPending {
			skillPendingProgress.Update(timeDelta)
			s := skillPendingProgress.Sprite()
			padding := (skillFrameHeight - int(s.Bounds().H())) / 2
			s.Draw(target, matrix.
				Moved(gfx.IVec(skillFrameWidth-padding, padding)).
				Moved(gfx.BottomRight.Align(s)))
		}
	}
	centerMatrix := pixel.IM.Moved(bottomLeft).Moved(pixel.V(
		float64(skillFrameWidth+(skillFrameHorizontalSpacing/2)),
		math.Ceil(float64(skillFrameHeight-1)*1.5)))
	atlas.GetTilesheetSprite("combat/menu/skill_arrows", 1, 1).Draw(target, centerMatrix)
	s.combatArrowAlpha -= timeDelta * 0.75
	ctx.DebugTR("arrow: %.2f, %d", s.combatArrowAlpha, s.combatArrowColumn)
	if s.combatArrowAlpha > 0 {
		atlas.GetTilesheetSprite("combat/menu/skill_arrows", s.combatArrowColumn, 1).DrawColorMask(target, centerMatrix, colors.Alpha(s.combatArrowAlpha))
	}

	badgeBottomLeft := pixel.V(3, 3)
	badgeBottomRight := pixel.V(targetBounds.W()-3, 3)
	skillStatsBadge.Render(ctx, target, pixel.IM.Moved(badgeBottomLeft), gfx.BottomLeft)
	if s.Player.NextSkill != nil {
		if s.Player.IsNextSkillCommitted() {
			skillCommittedCancelBadge.Render(ctx, target, pixel.IM.Moved(badgeBottomRight), gfx.BottomRight)
		} else {
			skillPendingCancelBadge.Render(ctx, target, pixel.IM.Moved(badgeBottomRight), gfx.BottomRight)
		}
		if ctx.Controls.ButtonA().JustPressed() {
			s.Player.NextSkillCommitted = true
		} else if ctx.Controls.ButtonB().JustPressed() {
			s.Player.NextSkill = nil
			s.Player.NextSkillCommitted = false
		}
	} else {
		skillMenuBadge.Render(ctx, target, pixel.IM.Moved(badgeBottomRight), gfx.BottomRight)
	}

	if s.Player.NextSkill != nil && ctx.Controls.ButtonB().JustPressed() {
		s.Player.NextSkill = nil
	}

	if ctx.Controls.ButtonSelect().JustPressed() {
		ctx.Notify("TODO - add stats menu")
	}

	if ctx.Controls.ButtonStart().JustPressed() {
		ctx.Notify("TODO - add combat menu")
	}
}

func (s *State) simpleSkillContent(text string) *textbox.Content {
	c, exist := s.cachedContents[text]
	if !exist {
		c = skillText.NewComplexContent(text)
		s.cachedContents[text] = c
	}
	return c
}
