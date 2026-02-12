package xenolog

import (
	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	paneW       = 64
	onDarkStyle = badges.ButtonColorStyle{
		Action:    screenColors.Text,
		Button:    screenColors.Clear,
		Highlight: screenColors.Highlight,
	}
)

type skillSwapMenu struct {
	screen     *screen.Instance
	original   rpg.SkillId
	swapWith   rpg.SkillId
	list       *scrollList[*skillSwapMenu]
	onComplete func(id rpg.SkillId)

	cursorStartIndex int
}

func newSkillSwapMenu(screen *screen.Instance, original rpg.SkillId, onComplete func(id rpg.SkillId)) *skillSwapMenu {
	menu := &skillSwapMenu{
		screen:     screen,
		original:   original,
		swapWith:   original,
		onComplete: onComplete,
	}
	menu.list = newScrollList[*skillSwapMenu](menu, scrollListOptions{
		targetHeight:              screenHeight,
		transitionTime:            0.15,
		verticalItemMargin:        1,
		verticalListPaddingTop:    10,
		verticalListPaddingBottom: 4,
	})
	menu.list.Add(&skillListItemTitle{})
	menu.list.Add(&skillListItemAction{
		label: "filter: none",
	})
	menu.list.Add(&skillListItemAction{
		label: "sort: abc",
	})
	menu.list.Add(&skillListItem{
		skillId: original,
	})
	menu.list.highlightLast()
	menu.cursorStartIndex = menu.list.highlightedIndex
	menu.list.anchorThreshold = menu.cursorStartIndex
	menu.list.Add(&skillListItemSpacer{})
	for _, skillId := range game.CurrentSave().UnlockedSkillsSorted() {
		if skillId == original {
			continue
		}
		menu.list.Add(&skillListItem{
			skillId: skillId,
		})
	}
	menu.list.Add(&skillListItemSpacer{})
	menu.list.Add(&skillListItemAction{
		label: "back to top",
		handler: func(m *skillSwapMenu) {
			m.list.highlight(m.cursorStartIndex)
		},
	})
	menu.list.cursor = &skillsCursor{}
	return menu
}

func (*skillSwapMenu) Enter() {}

type skillListItem struct {
	baseListItem[*skillSwapMenu]
	skillId rpg.SkillId
}

func (p *skillListItem) ButtonAJustPressed(m *skillSwapMenu) {
	m.onComplete(p.skillId)
	m.screen.PopMenuAnimated()
}

func (p *skillListItem) ButtonSelectJustPressed(m *skillSwapMenu) {
	m.screen.PushMenuAnimated(newSkillDetailMenu(m.screen, p.skillId))
}

func (p *skillListItem) OnHighlight(m *skillSwapMenu) {
	m.swapWith = p.skillId
}

func (p *skillListItem) OnUnhighlight(m *skillSwapMenu) {
	m.swapWith = rpg.UnsetSkillId
}

func (p *skillListItem) Height() int {
	return 12
}

func (p *skillListItem) Render(m *skillSwapMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	centerY := topLeftY - p.Height()/2
	smallText := newTextRenderer(target, smallTextbox)

	mask := screenColors.Text
	if highlightProgress > 0 {
		mask = colors.Lerp(mask, screenColors.Highlight, highlightProgress)
	}
	name := "???"
	if p.skillId != "" {
		name = p.skillId.Get().Name
	}
	smallText.render(name, screenWidth/2, centerY, mask, tbcfg.RenderFrom(gfx.Centered))
}

type skillListItemSpacer struct {
	baseListItem[*skillSwapMenu]
}

func (p *skillListItemSpacer) Height() int {
	return 6
}

func (p *skillListItemSpacer) Render(m *skillSwapMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	w := 35
	gfx.DrawRect(atlas, target, pixel.IM.Moved(gfx.IVec(screenWidth/2, topLeftY-2)), gfx.TopCenter, w, 1, screenColors.Dark)
}

func (p *skillListItemSpacer) DoSkipHighlight(m *skillSwapMenu) bool {
	return true
}

type skillListItemTitle struct {
	baseListItem[*skillSwapMenu]
}

func (p *skillListItemTitle) Height() int {
	return 19
}

func (p *skillListItemTitle) DoSkipHighlight(m *skillSwapMenu) bool {
	return true
}

func (p *skillListItemTitle) Render(m *skillSwapMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	titleText := newTextRenderer(target, titleTextbox)
	titleText.render("Swap Skill", screenWidth/2, topLeftY-2, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))
}

type skillListItemAction struct {
	baseListItem[*skillSwapMenu]
	label   string
	handler func(p *skillSwapMenu)
}

func (p *skillListItemAction) Height() int {
	return 11
}

func (p *skillListItemAction) ButtonAJustPressed(m *skillSwapMenu) {
	if p.handler == nil {
		game.DebugNotificationf("todo - implement sort and filter")
		return
	}
	p.handler(m)
}

func (p *skillListItemAction) Render(m *skillSwapMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	box := smallTextbox.NewSimpleContent(p.label)
	topCenter := pixel.IM.Moved(gfx.IVec(screenWidth/2, topLeftY-2))

	bgMask, fgMask := screenColors.Dark, screenColors.Text
	if m.list.highlightedItem() == p {
		bgMask, fgMask = flashingHighlight(), screenColors.Dark
	}
	padding := 3
	frameR := pixel.R(0, 0, float64(box.Width()+padding*2), 7)
	frame1px.Draw(target, frameR, topCenter, frames.WithColor(bgMask), frames.WithRenderOrigin(gfx.TopCenter))
	box.Render(target, topCenter.Moved(gfx.IVec(0, -1)), tbcfg.RenderFrom(gfx.TopCenter), tbcfg.Foreground(fgMask))
}

type skillsCursor struct{}

func (c *skillsCursor) Render(m *skillSwapMenu, centerLeft int, target pixel.Target, index int, movementProgress float64, highlightProgress float64, movingDown bool) {
	mask := screenColors.Highlight

	if index < m.cursorStartIndex {
		// Above threshold - hidden, unless transitioning while moving up
		if index == m.cursorStartIndex-1 && !movingDown {
			mask = colors.Lerp(mask, screenColors.Clear, highlightProgress)
		} else {
			mask = screenColors.Clear
		}
	} else if index == m.cursorStartIndex && movingDown {
		// Just crossed threshold going down - fade in
		mask = colors.Lerp(screenColors.Clear, mask, highlightProgress)
	}

	smallText := newTextRenderer(target, smallTextbox)
	dx := screenWidth / 3 / 2
	smallText.render(">", screenWidth/2-dx, centerLeft, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	smallText.render("<", screenWidth/2+dx, centerLeft, mask, tbcfg.RenderFrom(gfx.RightCenter))
}

func (m *skillSwapMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenuAnimated()
	}
	m.renderSkillPane(target, true, "original", m.original)
	newSkill := m.swapWith
	if newSkill == m.original {
		newSkill = rpg.UnsetSkillId
	}
	m.renderSkillPane(target, false, "new skill", newSkill)

	// list
	m.list.Render(game.Controls[*State](), target, timeDelta)
}

func (m *skillSwapMenu) renderSkillPane(target pixel.Target, isLeft bool, label string, skillId rpg.SkillId) {
	smallText := newTextRenderer(target, smallTextbox)

	var centerX, paneX, borderX int
	var origin gfx.OriginLocation

	tickBarOpt := tick_bar.NewDrawOptions(8, 13)
	if isLeft {
		centerX = paneW / 2
		paneX = 0
		borderX = paneW
		origin = gfx.BottomLeft
	} else {
		centerX = screenWidth - paneW/2
		paneX = screenWidth
		borderX = screenWidth - paneW
		origin = gfx.BottomRight
		tickBarOpt = tickBarOpt.Flip(true)
	}

	// Background pane
	gfx.DrawRect(atlas, target, pixel.IM.Moved(gfx.IVec(paneX, 0)), origin, paneW, screenHeight, screenColors.Dark)

	// Border line
	gfx.DrawRect(atlas, target, pixel.IM.Moved(gfx.IVec(borderX, 0)), gfx.BottomLeft, 1, screenHeight, screenColors.Highlight)

	// arrow
	arrowBoxSize := 11.0
	arrowBoxR := pixel.R(0, 0, arrowBoxSize, arrowBoxSize)
	arrowBoxCenter := pixel.IM.Moved(gfx.IVec(borderX, screenHeight-18))
	frame2px.Draw(target, arrowBoxR, arrowBoxCenter, frames.WithColor(screenColors.Dark), frames.WithRenderOrigin(gfx.Centered))
	frame2pxBorder.Draw(target, arrowBoxR, arrowBoxCenter, frames.WithColor(screenColors.Clear), frames.WithRenderOrigin(gfx.Centered))
	arrowRight.DrawColorMask(target, arrowBoxCenter.Moved(gfx.IVec(1, 1)), screenColors.Text)

	// Skill info
	smallText.render(label, centerX, screenHeight-4, screenColors.Text, tbcfg.RenderFrom(gfx.TopCenter))

	skillNameY := screenHeight - 15
	if skillId == rpg.UnsetSkillId {
		smallText.render("-------", centerX, skillNameY, screenColors.Clear, tbcfg.RenderFrom(gfx.TopCenter))
		return
	}

	skill := skillId.Get()
	smallText.render(skill.Name, centerX, skillNameY, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))
	tickBarRenderer.Draw(m.screen.SpriteFilterBuffer().Target(), pixel.IM.Moved(gfx.IVec(centerX, screenHeight-26)), &skill, tickBarOpt)

	if isLeft {
		badgeBCancelOnDark.Render(target, gfx.Moved(3, 2), gfx.BottomLeft)
	} else {
		badgeSelectDetailsOnDark.Render(target, gfx.Moved(screenWidth-3, 2), gfx.BottomRight)
	}
}
