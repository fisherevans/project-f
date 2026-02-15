package xenolog

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
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
	screen        *screen.Instance[*State]
	nav           *navigtion.System[*animechMenu]
	statsActions  []*animechMenuAction
	skillsActions []*animechMenuAction
	skillItems    map[input.Direction]*navigtion.SimpleItem[*animechMenu]
}

func newAnimechMenu(s *screen.Instance[*State]) *animechMenu {
	menu := &animechMenu{
		screen: s,
		statsActions: []*animechMenuAction{
			{
				label: "re-spec",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					if game.CurrentSave().Animech.Upgrades.GetLevel() <= 1 {
						return
					}
					modal := newConfirmationModal(
						m.screen,
						"Re-spec Animech?",
						"This will reset all of your stat boosts. You get the experience back.",
						func() {
							respec()
							saveOrNotify()
							game.DebugNotificationf("Removed all level ups")
						})
					m.screen.PushMenu(modal, false)
				}),
				disabled: func(m *animechMenu) bool {
					return game.CurrentSave().Animech.Upgrades.GetLevel() <= 1
				},
			},
			{
				label: "level up",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					if !isAnimechUpgradeAvailable() {
						return
					}
					m.screen.PushMenuAnimated(newAnimechStatsMenu(m.screen))
				}),
				disabled: func(m *animechMenu) bool {
					return !isAnimechUpgradeAvailable()
				},
			},
		},
		skillsActions: []*animechMenuAction{
			{
				label: "edit",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					s.PushMenuAnimated(newSkillSetMenu(s))
				}),
			},
			{
				label: "store",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					game.DebugNotificationf("todo - implement storing load outs")
				}),
			},
			{
				label: "load",
				SimpleItem: navigtion.NewSimpleItem[*animechMenu]().WithButtonAHandler(func(m *animechMenu) {
					game.DebugNotificationf("todo - implement loading load outs")
				}),
			},
		},
		skillItems: map[input.Direction]*navigtion.SimpleItem[*animechMenu]{},
	}
	menu.nav = navigtion.NewSystem[*animechMenu](menu)
	for _, a := range menu.statsActions {
		menu.nav.AddNextToLast(input.Right, a)
	}
	for id, dir := range []input.Direction{input.Left, input.Down, input.Right, input.Up} {
		menu.skillItems[dir] = newSkillInput(dir)
		if id == 0 {
			menu.nav.AddItem(menu.skillItems[dir])
		} else {
			menu.nav.AddNextToLast(input.Up, menu.skillItems[dir])
		}
	}
	menu.nav.SetLastItem(menu.statsActions[len(menu.statsActions)-1])
	for _, a := range menu.skillsActions {
		menu.nav.AddNextToLast(input.Right, a).Below(menu.skillItems[input.Left])
	}
	menu.nav.SetLastItem(nil)
	for _, i := range append(menu.statsActions, menu.skillsActions...) {
		if i.disabled == nil || !i.disabled(menu) {
			menu.nav.Highlight(i)
			break
		}
	}
	return menu
}

func newSkillInput(dir input.Direction) *navigtion.SimpleItem[*animechMenu] {
	onPress := func(a *animechMenu) {
		skillId := *game.CurrentSave().Animech.SkillSet.DirectionalSkill(dir)
		if skillId == rpg.UnsetSkillId {
			return
		}
		a.screen.PushMenuAnimated(newSkillDetailMenu(a.screen, skillId))
	}
	return navigtion.NewSimpleItem[*animechMenu]().
		WithButtonAHandler(onPress).
		WithButtonSelectHandler(onPress)
}

type animechMenuAction struct {
	*navigtion.SimpleItem[*animechMenu]
	label    string
	disabled func(m *animechMenu) bool
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
		m.screen.PopMenuAnimated()
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
		disabled    bool
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
		if action.disabled != nil {
			b.disabled = action.disabled(m)
		}
		buttons = append(buttons, b)
	}
	topLeft := topCenter.Moved(gfx.IVec(-animechFrameWidth/2, -framePadding))
	actionBottomLeft := topLeft.Moved(gfx.IVec(int((float64(animechFrameWidth)-totalWidth)/2), -actionHeight))
	for _, b := range buttons {
		bgMask, borderMask, textMask := screenColors.Dark, screenColors.Clear, screenColors.Text
		if b.disabled {
			bgMask = screenColors.Clear
			borderMask, textMask = screenColors.Dark, screenColors.Dark
			if b.highlighted {
				borderMask = screenColors.Text
			}
		} else if b.highlighted {
			bgMask = flashingHighlight()
			borderMask, textMask = screenColors.Dark, screenColors.Dark
		}
		frame2px.Draw(target, b.rect, actionBottomLeft,
			frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(bgMask))
		frame2pxBorder.Draw(target, b.rect, actionBottomLeft,
			frames.WithRenderOrigin(gfx.BottomLeft), frames.WithColor(borderMask))
		b.label.Render(target, actionBottomLeft.Moved(b.rect.Bounds().Center()),
			tbcfg.RenderFrom(gfx.Centered), tbcfg.Foreground(textMask))
		actionBottomLeft = actionBottomLeft.Moved(pixel.V(b.rect.W()+actionPadding, 0))
	}
}

func (m *animechMenu) renderStats(target pixel.Target, topCenter pixel.Matrix) {
	a := game.CurrentSave().Animech
	level := a.Upgrades.GetLevel()
	m.drawFrame(target, topCenter, fmt.Sprintf("Level %d", level), screenColors.Highlight, screenColors.Text)
	m.drawActions(target, topCenter.Moved(gfx.IVec(0, -animechFrameHeight+3)), m.statsActions)

	smallTxt := newTextRenderer(target, smallTextbox).withMatrix(topCenter)
	regularTxt := newTextRenderer(target, regularTextbox).withMatrix(topCenter)

	y := -animechFrameTopPadding

	y -= 7
	smallTxt.render("Stats", 0, y, screenColors.Text, tbcfg.RenderFrom(gfx.TopCenter))

	renderStat := func(name string, value int) {
		arrowDx := 0
		arrowRight.DrawColorMask(target, topCenter.Moved(gfx.IVec(arrowDx, y+1)), screenColors.Dark)
		smallTxt.render(name, arrowDx-7, y, screenColors.Text, tbcfg.RenderFrom(gfx.RightCenter))
		regularTxt.render(fmt.Sprintf("%d", value), arrowDx+6, y, screenColors.Highlight, tbcfg.RenderFrom(gfx.LeftCenter))
	}

	y -= 14
	renderStat("shield", a.GetMaxShield())
	y -= 13
	renderStat("sync", a.GetMaxSync())

	y -= 13
	smallTxt.render("Experience Points:", 0, y, screenColors.Text, tbcfg.RenderFrom(gfx.TopCenter))
	y -= 7
	regularTxt.render(comma(a.Experience), 0, y, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))

	y -= 19
	xpNeeded := rpg.AnimechUpgradeExperienceRequiredToUpgrade(level+1) - a.Experience
	if xpNeeded <= 0 {
		w, _ := smallTxt.render("upgrade available", 0, y, flashingHighlight(), tbcfg.RenderFrom(gfx.Centered))
		arrowMask := colors.Lerp(flashingHighlight(), screenColors.Clear, 0.5)
		arrowX := w/2 + 5
		// +1's due to odd sprite sizes + origin rendering
		arrowRight.DrawColorMask(target, topCenter.Moved(gfx.IVec(-arrowX, y+1)), arrowMask)
		arrowLeft.DrawColorMask(target, topCenter.Moved(gfx.IVec(arrowX+1, y+1)), arrowMask)
	} else {
		label := fmt.Sprintf("{+c:xenolog_highlight,+u}%d{+c:xenolog_text,-u} XP 'til level up", xpNeeded)
		smallTxt.render(label, 0, y, screenColors.Text, tbcfg.RenderFrom(gfx.Centered))
	}
}

func (m *animechMenu) renderSkills(target pixel.Target, topCenter pixel.Matrix) {
	m.drawFrame(target, topCenter, "Skill Set", screenColors.Highlight, screenColors.Text)
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
		borderMask := screenColors.Clear
		bgMask := screenColors.Dark
		fgMask := screenColors.Text
		arrowMask := screenColors.Highlight
		if m.nav.IsHighlighted(m.skillItems[dir]) {
			borderMask = screenColors.Dark
			bgMask = flashingHighlight()
			fgMask = screenColors.Dark
			arrowMask = screenColors.Clear
		}
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
		game.CurrentSave().Animech.Experience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
	}
	for game.CurrentSave().Animech.Upgrades.ShieldLevel > 0 {
		game.CurrentSave().Animech.Upgrades.ShieldLevel--
		game.CurrentSave().Animech.Experience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
	}
}

func comma(n int) string {
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
