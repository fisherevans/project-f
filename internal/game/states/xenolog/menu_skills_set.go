package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillSetRowHeight = 14
	skillSetArrows    = map[input.Direction]pixelutil.BoundedDrawable{
		input.NotPressed: atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 1, 1),
		input.Up:         atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 2, 1),
		input.Right:      atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 3, 1),
		input.Down:       atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 4, 1),
		input.Left:       atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 5, 1),
	}
)

type skillSetMenu struct {
	screen               *Screen
	currentSkillSelected input.Direction
}

func newSkillSetMenu(screen *Screen) *skillSetMenu {
	m := &skillSetMenu{
		screen:               screen,
		currentSkillSelected: input.Up,
	}
	return m
}

func (*skillSetMenu) Enter() {}

func (m *skillSetMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInput()
	m.render(target)
}

func (m *skillSetMenu) handleInput() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu()
	}
	if game.Controls[*State]().DPad().JustPressed() {
		m.currentSkillSelected = game.Controls[*State]().DPad().GetDirection()
	}
	if game.Controls[*State]().ButtonA().JustPressed() {
		original := *game.CurrentSave().Animech.SkillSet.DirectionalSkill(m.currentSkillSelected)
		m.screen.PushMenu(newSkillsMenu(m.screen, original, func(newSkill rpg.SkillId) {
			if newSkill == original {
				return
			}
			swapDirection, doSwap := game.CurrentSave().Animech.SkillSet.DirectionOfSkill(newSkill)
			if doSwap {
				toSwap := game.CurrentSave().Animech.SkillSet.DirectionalSkill(swapDirection)
				*toSwap = original
			}
			toSet := game.CurrentSave().Animech.SkillSet.DirectionalSkill(m.currentSkillSelected)
			*toSet = newSkill
			saveOrNotify()
		}))
	}
}

func (m *skillSetMenu) render(target pixel.Target) {
	x := screenWidth / 2
	y := screenHeight / 3 * 2
	dy := 14
	dx := 55

	m.renderCurrentSkill(input.Up, x, y+dy, target)
	m.renderCurrentSkill(input.Right, x+dx, y, target)
	m.renderCurrentSkill(input.Down, x, y-dy, target)
	m.renderCurrentSkill(input.Left, x-dx, y, target)

	skillSetArrows[m.currentSkillSelected].Draw(target, pixel.IM.Moved(gfx.IVec(x, y)))
}

func (m *skillSetMenu) renderCurrentSkill(direction input.Direction, x, y int, target pixel.Target) {
	skillId := game.CurrentSave().Animech.SkillSet.DirectionalSkill(direction)
	regularTxt := newTextRenderer(target, regularTextbox)

	fgMask := colors.XenoLogText.RGBA
	bgMask := colors.XenoLogClear.RGBA
	borderMask := colors.XenoLogDark.RGBA
	if m.currentSkillSelected == direction {
		fgMask = colors.XenoLogDark.RGBA
		bgMask = colors.XenoLogHighlight.RGBA
		borderMask = colors.XenoLogDark.RGBA
	}

	var label string
	if *skillId == rpg.UnsetSkillId {
		label = "---"
	} else {
		label = skillId.Get().Name
	}
	frameR := pixel.R(0, 0, 86, 15)
	frame2px.Draw(target, frameR, pixel.IM.Moved(gfx.IVec(x, y)),
		frames.WithColor(bgMask), frames.WithRenderOrigin(gfx.Centered))
	frame2pxBorder.Draw(target, frameR, pixel.IM.Moved(gfx.IVec(x, y)),
		frames.WithColor(borderMask), frames.WithRenderOrigin(gfx.Centered))
	regularTxt.render(label, x, y, fgMask, tbcfg.RenderFrom(gfx.Centered))

}
