package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/navigtion"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	animechFrameWidth      = 94
	animechFrameHeight     = 100
	animechFrameTopPadding = 15
)

type animechMenu struct {
	screen        *Screen
	nav           *navigtion.System[*animechMenu]
	statsActions  []*animechMenuAction
	skillsActions []*animechMenuAction
}

func newAnimechMenu(s *Screen) *animechMenu {
	menu := &animechMenu{
		screen: s,
		statsActions: []*animechMenuAction{
			{
				label: "re-spec",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					respec()
					saveOrNotify()
					game.DebugNotification("Removed all level ups")
				}),
			},
			{
				label: "level up",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					m.screen.PushMenu(newAnimechStatsMenu(m.screen))
				}),
			},
		},
		skillsActions: []*animechMenuAction{
			{
				label: "edit",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					s.PushMenu(newSkillSetMenu(s))
				}),
			},
			{
				label: "store",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					game.DebugNotification("todo - implement storing load outs")
				}),
			},
			{
				label: "load",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					game.DebugNotification("todo - implement loading load outs")
				}),
			},
		},
		nav: navigtion.NewSystem[*animechMenu](),
	}
	for _, a := range menu.statsActions {
		menu.nav.AddNextToLast(input.Right, menu, a)
	}
	for _, a := range menu.skillsActions {
		menu.nav.AddNextToLast(input.Right, menu, a)
	}
	menu.nav.SetLastItem(nil)
	return menu
}

type animechMenuAction struct {
	*navigtion.SimpleItem[*animechMenu]
	label string
}

func (m *animechMenu) Enter() {}

func (m *animechMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInputs()
	topFramePadding := 7
	dx := (animechFrameWidth + (screenWidth-animechFrameWidth*2)/3) / 2
	m.renderStats(target, pixel.IM.Moved(gfx.IVec(screenWidth/2-dx, screenHeight-topFramePadding)))
	m.renderSkills(target, pixel.IM.Moved(gfx.IVec(screenWidth/2+dx, screenHeight-topFramePadding)))
}

func (m *animechMenu) handleInputs() {
	controls := game.Controls[*State]()
	if controls.ButtonB().JustPressed() {
		m.screen.PopMenu()
		return
	}
	m.nav.HandleInputs(controls, m)
}

func (m *animechMenu) drawFrame(target pixel.Target, topCenter pixel.Matrix, title string, titleMask, frameMask pixel.RGBA) {
	frame4pxBorder.Draw(target, pixel.R(0, 0,
		float64(animechFrameWidth), float64(animechFrameHeight)),
		topCenter,
		frames.WithColor(frameMask),
		frames.WithRenderOrigin(gfx.TopCenter))
	titleTxt := newTextRenderer(target, titleTextbox).withMatrix(topCenter)
	titleTxt.render(title, 0, -4, titleMask, tbcfg.RenderFrom(gfx.TopCenter))

}

func (m *animechMenu) drawActions(target pixel.Target, topCenter pixel.Matrix, actions []*animechMenuAction) {
	type button struct {
		label       *textbox.Content
		rect        pixel.Rect
		highlighted bool
	}
	framePadding := 7
	actionPadding := 5.0
	textPadding := 5
	actionHeight := 13
	var buttons []button
	totalWidth := 0.0
	for id, action := range actions {
		if id != 0 {
			totalWidth += actionPadding
		}
		b := button{
			label:       smallTextbox.NewSimpleContent(action.label),
			highlighted: m.nav.IsHighlighted(action),
		}
		b.rect = pixel.R(0, 0, float64(b.label.Width()+textPadding*2), float64(actionHeight))
		totalWidth += b.rect.W()
		buttons = append(buttons, b)
	}
	topLeft := topCenter.Moved(gfx.IVec(-animechFrameWidth/2, -framePadding))
	actionBottomLeft := topLeft.Moved(gfx.IVec(int((float64(animechFrameWidth)-totalWidth)/2), -actionHeight))
	for _, b := range buttons {
		bgMask, borderMask, textMask := colors.XenoLogDark.RGBA, colors.XenoLogText.RGBA, colors.XenoLogText.RGBA
		if b.highlighted {
			bgMask = flashingHighlight()
			borderMask, textMask = colors.XenoLogHighlight.RGBA, colors.XenoLogDark.RGBA
		}
		frame2px.Draw(target, b.rect, actionBottomLeft,
			frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(bgMask))
		frame2pxBorder.Draw(target, b.rect, actionBottomLeft,
			frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(borderMask))
		smallTextbox.Render(target, actionBottomLeft.Moved(b.rect.Bounds().Center()), b.label,
			tbcfg.RenderFrom(gfx.Centered), tbcfg.Foreground(textMask))
		actionBottomLeft = actionBottomLeft.Moved(pixel.V(b.rect.W()+actionPadding, 0))
	}
}

func (m *animechMenu) renderStats(target pixel.Target, topCenter pixel.Matrix) {
	a := game.CurrentSave().Animech
	level := a.Upgrades.GetLevel()
	m.drawFrame(target, topCenter, fmt.Sprintf("Level %d", level), colors.XenoLogHighlight.RGBA, colors.XenoLogText.RGBA)
	m.drawActions(target, topCenter.Moved(gfx.IVec(0, -animechFrameHeight+3)), m.statsActions)

	smallTxt := newTextRenderer(target, smallTextbox).withMatrix(topCenter)
	regularTxt := newTextRenderer(target, regularTextbox).withMatrix(topCenter)

	y := -animechFrameTopPadding

	y -= 7
	smallTxt.render("Stats", 0, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopCenter))

	renderStat := func(name string, value int) {
		arrowDx := 0
		arrowRight.DrawColorMask(target, topCenter.Moved(gfx.IVec(arrowDx, y)), colors.XenoLogDark.RGBA)
		smallTxt.render(name, arrowDx-7, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.RightCenter))
		regularTxt.render(fmt.Sprintf("%d", value), arrowDx+6, y, colors.XenoLogHighlight.RGBA, tbcfg.RenderFrom(gfx.LeftCenter))
	}

	y -= 14
	renderStat("shield", a.GetMaxShield())
	y -= 13
	renderStat("sync", a.GetMaxSync())

	y -= 13
	smallTxt.render("Experience Points:", 0, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopCenter))
	y -= 7
	regularTxt.render(comma(a.AnimechExperience), 0, y, colors.XenoLogHighlight.RGBA, tbcfg.RenderFrom(gfx.TopCenter))

	y -= 19
	xpNeeded := rpg.AnimechUpgradeExperienceRequiredToUpgrade(level+1) - a.AnimechExperience
	if xpNeeded <= 0 {
		w, _ := smallTxt.render("upgrade available", 0, y, flashingHighlight(), tbcfg.RenderFrom(gfx.Centered))
		arrowMask := colors.Lerp(flashingHighlight(), colors.XenoLogClear.RGBA, 0.5)
		arrowX := w/2 + 5
		// +1's due to odd sprite sizes + origin rendering
		arrowRight.DrawColorMask(target, topCenter.Moved(gfx.IVec(-arrowX, y+1)), arrowMask)
		arrowLeft.DrawColorMask(target, topCenter.Moved(gfx.IVec(arrowX+1, y+1)), arrowMask)
	} else {
		label := fmt.Sprintf("{+c:xenolog_highlight,+u}%d{+c:xenolog_text,-u} XP 'til level up", xpNeeded)
		smallTxt.render(label, 0, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.Centered))
	}
}

func (m *animechMenu) renderSkills(target pixel.Target, topCenter pixel.Matrix) {
	m.drawFrame(target, topCenter, "Skill Set", colors.XenoLogHighlight.RGBA, colors.XenoLogText.RGBA)
	m.drawActions(target, topCenter.Moved(gfx.IVec(0, -animechFrameHeight+3)), m.skillsActions)
	s := game.CurrentSave().Animech.SkillSet

	//smallTxt := newTextRenderer(target, smallTextbox).withMatrix(topCenter)
	regularTxt := newTextRenderer(target, regularTextbox).withMatrix(topCenter)

	y := -animechFrameTopPadding

	skillLabel := func(s rpg.SkillId) string {
		if s == rpg.UnsetSkillId {
			return "---"
		}
		return s.Get().Name
	}

	skillHeight := 15
	renderSkill := func(dir input.Direction, name string) {
		borderMask := colors.XenoLogDark.RGBA
		bgMask := colors.XenoLogClear.RGBA
		fgMask := colors.XenoLogText.RGBA
		arrowMask := colors.XenoLogHighlight.RGBA
		skillWidth := animechFrameWidth - 8
		frameR := pixel.R(0, 0, float64(skillWidth), float64(skillHeight))
		frame2px.Draw(target, frameR, topCenter.Moved(gfx.IVec(0, y)), frames.WithColor(bgMask), frames.WithRenderOrigin(gfx.TopCenter))
		frame2pxBorder.Draw(target, frameR, topCenter.Moved(gfx.IVec(0, y)), frames.WithColor(borderMask), frames.WithRenderOrigin(gfx.TopCenter))
		regularTxt.render(name, -skillWidth/2+4, y-skillHeight/2-1, fgMask, tbcfg.RenderFrom(gfx.LeftCenter)) // -1 die to odd ints
		arrow := arrowSprite(dir)
		arrow.DrawColorMask(target, topCenter.Moved(gfx.IVec(skillWidth/2-4, y-skillHeight/2)).Moved(gfx.RightCenter.Align(arrow)), arrowMask)
	}

	y -= 3
	skillPadding := 5
	renderSkill(input.Up, skillLabel(s.Skill1))
	y -= skillHeight + skillPadding
	renderSkill(input.Right, skillLabel(s.Skill2))
	y -= skillHeight + skillPadding
	renderSkill(input.Down, skillLabel(s.Skill3))
	y -= skillHeight + skillPadding
	renderSkill(input.Left, skillLabel(s.Skill4))
	y -= skillHeight + skillPadding
}

func respec() {
	for game.CurrentSave().Animech.Upgrades.SyncLevel > 0 {
		game.CurrentSave().Animech.Upgrades.SyncLevel--
		game.CurrentSave().Animech.AnimechExperience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
	}
	for game.CurrentSave().Animech.Upgrades.ShieldLevel > 0 {
		game.CurrentSave().Animech.Upgrades.ShieldLevel--
		game.CurrentSave().Animech.AnimechExperience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
	}
}

func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
