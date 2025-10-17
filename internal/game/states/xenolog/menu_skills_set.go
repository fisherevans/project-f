package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
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
		m.screen.PushMenu(newSkillSwapMenu(m.screen, original, func(newSkill rpg.SkillId) {
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
	if game.Controls[*State]().ButtonSelect().JustPressed() {
		skillId := *game.CurrentSave().Animech.SkillSet.DirectionalSkill(m.currentSkillSelected)
		m.screen.PushMenu(newSkillDetailMenu(m.screen, skillId))
	}
}

func (m *skillSetMenu) render(target pixel.Target) {
	x := screenWidth / 2
	y := 123
	titleTxt := newTextRenderer(target, titleTextbox)
	titleTxt.render("Animech Skill Set", x, y, colors.XenoLogHighlight.RGBA, tbcfg.RenderFrom(gfx.TopCenter))

	// render arrow skills delta'ed from center
	y -= 36
	arrowDy := 14
	arrowDx := 55
	arrowSkillWidth := 90
	renderDir := func(dir input.Direction) {
		switch dir {
		case input.Up:
			m.renderSkill(dir, arrowSkillWidth, x, y+arrowDy, target)
		case input.Right:
			m.renderSkill(dir, arrowSkillWidth, x+arrowDx, y, target)
		case input.Down:
			m.renderSkill(dir, arrowSkillWidth, x, y-arrowDy, target)
		case input.Left:
			m.renderSkill(dir, arrowSkillWidth, x-arrowDx, y, target)
		}
	}
	// render the selected direction last
	selectedDir := input.NotPressed
	for _, dir := range input.Directions {
		if dir == m.currentSkillSelected {
			selectedDir = dir
			continue
		}
		renderDir(dir)
	}
	renderDir(selectedDir)
	skillSetArrows[m.currentSkillSelected].Draw(target, pixel.IM.Moved(gfx.IVec(x, y)))

	y -= arrowDy + 12
	m.renderSkillDetail(target, pixel.IM.Moved(gfx.IVec(x, y)))
}

func (m *skillSetMenu) renderSkill(direction input.Direction, width, x, y int, target pixel.Target) {
	skillId := game.CurrentSave().Animech.SkillSet.DirectionalSkill(direction)
	regularTxt := newTextRenderer(target, regularTextbox)

	fgMask := colors.XenoLogText.RGBA
	bgMask := colors.XenoLogDark.RGBA
	borderMask := colors.XenoLogClear.RGBA
	boldBorder := false
	if m.currentSkillSelected == direction {
		fgMask = colors.XenoLogDark.RGBA
		bgMask = flashingHighlight()
		borderMask = colors.XenoLogDark.RGBA
		//boldBorder = true
	}

	var label string
	if *skillId == rpg.UnsetSkillId {
		label = "---"
	} else {
		label = skillId.Get().Name
	}
	frameR := pixel.R(0, 0, float64(width), 15)
	if boldBorder {
		boldFrameR := pixel.R(0, 0, frameR.W()+2, frameR.H()+2)
		frame2px.Draw(target, boldFrameR, pixel.IM.Moved(gfx.IVec(x, y)),
			frames.WithColor(borderMask), frames.WithRenderOrigin(gfx.Centered))
	}
	frame2px.Draw(target, frameR, pixel.IM.Moved(gfx.IVec(x, y)),
		frames.WithColor(bgMask), frames.WithRenderOrigin(gfx.Centered))
	frame2pxBorder.Draw(target, frameR, pixel.IM.Moved(gfx.IVec(x, y)),
		frames.WithColor(borderMask), frames.WithRenderOrigin(gfx.Centered))
	regularTxt.render(label, x, y, fgMask, tbcfg.RenderFrom(gfx.Centered))
}

var (
	skillSetBadgeStyle = badges.ButtonColorStyle{
		Action:    colors.XenoLogHighlight.RGBA,
		Button:    colors.XenoLogDark.RGBA,
		Highlight: colors.XenoLogHighlight.RGBA,
	}
	badgeASwap         = badges.Using(atlas).ButtonAction("A", "swap skill", skillSetBadgeStyle).Flipped()
	badgeSelectDetails = badges.Using(atlas).ButtonAction("select", "view skill details", skillSetBadgeStyle)

	skillSetCurrentSkillFrameWidth  = screenWidth - 20
	skillSetCurrentSkillFrameHeight = 40

	skillSetDetailSmallTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(skillSetCurrentSkillFrameWidth-10, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(colors.XenoLogText.RGBA),
			tbcfg.RenderFrom(gfx.TopCenter)))
)

func (m *skillSetMenu) renderSkillDetail(target pixel.Target, topCenter pixel.Matrix) {
	regularTxt := newTextRenderer(target, regularTextbox).withMatrix(topCenter).withOpts(tbcfg.RenderFrom(gfx.TopCenter))
	smallTxt := newTextRenderer(target, skillSetDetailSmallTextbox).withMatrix(topCenter)

	frameR := pixel.R(0, 0, float64(skillSetCurrentSkillFrameWidth), float64(skillSetCurrentSkillFrameHeight))
	frame4px.Draw(target, frameR, topCenter, frames.WithRenderOrigin(gfx.TopCenter), frames.WithColor(colors.XenoLogDark.RGBA))
	frame4pxBorder.Draw(target, frameR, topCenter, frames.WithRenderOrigin(gfx.TopCenter), frames.WithColor(colors.XenoLogText.RGBA))

	dy := -3

	skillId := game.CurrentSave().Animech.SkillSet.DirectionalSkill(m.currentSkillSelected)
	if *skillId == rpg.UnsetSkillId {
		regularTxt.render("Select a skill to edit it", 0, dy, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopCenter))
		return
	}

	regularTxt.render(skillId.Get().Name, 0, dy, colors.XenoLogHighlight.RGBA, tbcfg.RenderFrom(gfx.TopCenter))
	dy -= 15
	smallTxt.render(skillId.Get().Description, 0, dy, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopCenter))

	dy -= 27
	badgePadding := 11
	badgeDx := skillSetCurrentSkillFrameWidth/2 - badgePadding
	badgeSelectDetails.Render(target, topCenter.Moved(gfx.IVec(-badgeDx, dy)), gfx.TopLeft)
	badgeASwap.Render(target, topCenter.Moved(gfx.IVec(badgeDx, dy)), gfx.TopRight)
}
