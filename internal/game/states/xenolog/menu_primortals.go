package xenolog

import (
	"fmt"
	"time"

	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	tooltipMargin           = 2
	primortalListMargin     = 2
	primortalListLeftWidth  = 7
	primortalListRightWidth = 6
	primortalListRowPadding = 7
)

type primortalsMenu struct {
	screen                                 *screen.Instance
	list                                   *scrollList[*primortalsMenu]
	upgradeAbove, upgradeBelow, wasSkipped bool
	title                                  *primortalListTitle
	footer                                 *primortalListFooter
}

func newPrimortalsMenu(screen *screen.Instance) *primortalsMenu {
	menu := &primortalsMenu{
		screen: screen,
		title:  newPrimortalListTitle(),
		footer: newPrimortalListFooter(),
	}
	menu.list = newScrollList[*primortalsMenu](menu, scrollListOptions{
		targetHeight:              screenHeight,
		transitionTime:            0.15,
		verticalItemMargin:        1,
		verticalListPaddingTop:    8,
		verticalListPaddingBottom: 8,
	})
	menu.list.Add(menu.title)
	for i := 1; i <= rpg.MaxXenoLogEntryIndex; i++ {
		primortalId := rpg.XenoLogEntries[i]
		lastSeen := time.Time{}
		if progress, hasProgress := game.CurrentSave().Primortals[primortalId]; hasProgress {
			lastSeen = progress.LastSeen
		}
		menu.list.Add(&primortalListItem{
			xenologIndex: i,
			lastSeen:     lastSeen,
		})
	}
	menu.list.Add(menu.footer)
	menu.list.cursor = &primortalCursor{}
	return menu
}

func (*primortalsMenu) Enter() {}

type badgeAction struct {
	badge   badges.Instance
	handler func(m *primortalsMenu)
}

type horizontalActionItem struct {
	baseListItem[*primortalsMenu]
	selection int
	actions   []*badgeAction
}

func (p *horizontalActionItem) ButtonAJustPressed(m *primortalsMenu) {
	p.actions[p.selection].handler(m)
}

func (p *horizontalActionItem) DirectionJustPressed(m *primortalsMenu, dir input.Direction) {
	if dir == input.Left {
		p.selection--
		if p.selection < 0 {
			p.selection = len(p.actions) - 1
		}
	} else if dir == input.Right {
		p.selection++
		if p.selection >= len(p.actions) {
			p.selection = len(p.actions) - 1
		}
	}
}

func (p *horizontalActionItem) Render(m *primortalsMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	x := primortalListMargin + primortalListLeftWidth + 2

	badgeBg := screenColors.Dark
	badgeFg := screenColors.Text

	for id, action := range p.actions {
		bg, fg := badgeBg, badgeFg
		selected := id == p.selection
		if selected {
			bg = colors.Lerp(badgeBg, screenColors.Highlight, highlightProgress)
			fg = colors.Lerp(badgeFg, screenColors.Clear, highlightProgress)
		}
		action.badge.RenderMask(target, pixel.IM.Moved(gfx.IVec(x, topLeftY)), gfx.TopLeft, fg, bg)
		x += int(action.badge.Bounds().W()) + 5
	}
}

type primortalListTitle struct {
	*horizontalActionItem
}

func newPrimortalListTitle() *primortalListTitle {
	return &primortalListTitle{
		horizontalActionItem: &horizontalActionItem{
			actions: []*badgeAction{
				{
					badge: badgeBack,
					handler: func(m *primortalsMenu) {
						m.screen.PopMenuAnimated()
					},
				},
				{
					badge: badgeNextUnlock,
					handler: func(m *primortalsMenu) {
						m.ScrollToNextUpgrade()
					},
				},
			},
		},
	}
}

func (p *primortalListTitle) Height() int {
	return 33
}

func (p *primortalListTitle) Render(m *primortalsMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	titleText := newTextRenderer(target, titleTextbox)
	smallTxt := newTextRenderer(target, smallTextbox)

	x := primortalListMargin + primortalListLeftWidth
	_, dy := titleText.render("XenoLog", x, topLeftY, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopLeft))
	topLeftY -= dy + 3
	_, dy = smallTxt.render("A catalogue of your Primortal discoveries", x, topLeftY, screenColors.Text, tbcfg.RenderFrom(gfx.TopLeft))
	topLeftY -= dy + 3

	topLeftY -= 2
	p.horizontalActionItem.Render(m, topLeftY, target, highlightProgress)
}

type primortalListFooter struct {
	*horizontalActionItem
}

func newPrimortalListFooter() *primortalListFooter {
	return &primortalListFooter{
		horizontalActionItem: &horizontalActionItem{
			actions: []*badgeAction{
				{
					badge: badgeBackToTop,
					handler: func(m *primortalsMenu) {
						m.list.highlight(0)
					},
				},
			},
		},
	}
}

func (p *primortalListFooter) Height() int {
	return 11
}

func (p *primortalListFooter) Render(m *primortalsMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	p.horizontalActionItem.Render(m, topLeftY-2, target, highlightProgress)
}

type primortalListItem struct {
	baseListItem[*primortalsMenu]
	xenologIndex int
	lastSeen     time.Time
}

func (p *primortalListItem) RenderWasSkipped(m *primortalsMenu, wasAbove bool) {
	if p.xenologIndex == 9 {
		game.DebugBLf("skipped")
	}
	m.wasSkipped = true
	if p.upgradeAvailable() {
		if wasAbove {
			m.upgradeAbove = true
		} else {
			m.upgradeBelow = true
		}
	}
}

func (p *primortalListItem) ButtonAJustPressed(m *primortalsMenu) {
	pType, exists := rpg.XenoLogEntries[p.xenologIndex]
	if exists {
		m.screen.PushMenuAnimated(newPrimortalDetailMenu(m.screen, pType))
	}
}

func (p *primortalListItem) Height() int {
	return 18
}

func (p *primortalListItem) DoSkipHighlight(m *primortalsMenu) bool {
	_, exists := rpg.XenoLogEntries[p.xenologIndex]
	return !exists
}

func (p *primortalListItem) Render(m *primortalsMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	centerY := topLeftY - p.Height()/2
	txt := newTextRenderer(target, regularTextbox)
	smallTxt := newTextRenderer(target, smallTextbox)

	mask := screenColors.Text
	borderMask := screenColors.Dark
	if highlightProgress > 0 {
		mask = colors.Lerp(mask, screenColors.Highlight, highlightProgress)
		borderMask = colors.Lerp(borderMask, screenColors.Text, highlightProgress)
	}

	fr := rect(screenWidth-primortalListMargin*2-primortalListLeftWidth-primortalListRightWidth, p.Height()-1-primortalListMargin)
	frame2pxBorder.Draw(target, fr,
		pixel.IM.Moved(gfx.IVec(primortalListMargin+primortalListLeftWidth, centerY)),
		frames.WithColor(borderMask),
		frames.WithRenderOrigin(gfx.LeftCenter))

	leftX := primortalListMargin + primortalListLeftWidth + primortalListRowPadding
	dx, _ := smallTxt.render(fmt.Sprintf("%03d", p.xenologIndex), leftX, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	leftX += dx + primortalListRowPadding
	dot.DrawColorMask(target, pixel.IM.Moved(gfx.IVec(leftX, centerY+1)).Moved(gfx.LeftCenter.Align(dot)), mask)
	leftX += int(dot.Bounds().W()) + primortalListRowPadding

	pId, exists := rpg.XenoLogEntries[p.xenologIndex]
	if !exists {
		dx, _ = smallTxt.render("?????", leftX, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		return
	}

	primortal := pId.Primortal()
	name := primortal.Name
	isComplete := cheapestAvailableUnlock(primortal, game.CurrentSave()) == nil
	upgradeAvailable := p.upgradeAvailableForPrimortal(pId)

	dx, _ = txt.render(name, leftX, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))

	rightX := screenWidth - primortalListMargin - primortalListRowPadding - primortalListRightWidth
	if isComplete {
		smallTxt.render("[COMPLETE]", rightX, centerY, screenColors.Text, tbcfg.RenderFrom(gfx.RightCenter))
	} else if upgradeAvailable {
		upgradeRightX := rightX
		arrowPadding := 3
		arrowLeft.DrawColorMask(target, pixel.IM.Moved(gfx.IVec(upgradeRightX, centerY+1)).Moved(gfx.RightCenter.Align(arrowLeft)), flashingHighlight())
		upgradeRightX -= int(arrowLeft.Bounds().W()) + arrowPadding
		dx, _ := smallTxt.render("UPGRADE AVAILABLE", upgradeRightX, centerY, flashingHighlight(), tbcfg.RenderFrom(gfx.RightCenter))
		upgradeRightX -= dx + arrowPadding
		arrowRight.DrawColorMask(target, pixel.IM.Moved(gfx.IVec(upgradeRightX, centerY+1)).Moved(gfx.RightCenter.Align(arrowLeft)), flashingHighlight())
	}
}

func (p *primortalListItem) upgradeAvailable() bool {
	if pId, ok := rpg.XenoLogEntries[p.xenologIndex]; ok {
		return p.upgradeAvailableForPrimortal(pId)
	}
	return false
}

func (p *primortalListItem) upgradeAvailableForPrimortal(pId rpg.PrimortalType) bool {
	if save, exists := game.CurrentSave().Primortals[pId]; exists {
		if next := cheapestAvailableUnlock(pId.Primortal(), game.CurrentSave()); next != nil && save.ResearchPoints >= next.Cost {
			return true
		}
	}
	return false
}

type primortalCursor struct{}

func (c *primortalCursor) Render(m *primortalsMenu, centerLeftY int, target pixel.Target, index int, movementProgress float64, highlightProgress float64, movingDown bool) {
	x := primortalListMargin + primortalListLeftWidth/2
	mask := screenColors.Highlight
	if movementProgress < 0 {
		mask = colors.Lerp(screenColors.Dark, screenColors.Highlight, movementProgress)
	}
	if index == 0 || index == len(m.list.items)-1 {
		mask = colors.Lerp(mask, screenColors.Clear, movementProgress)
	}
	scrollCursor.DrawColorMask(target, pixel.IM.Moved(gfx.IVec(x, centerLeftY+1)), mask)
}

func (m *primortalsMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenuAnimated()
	}

	m.upgradeAbove, m.upgradeBelow, m.wasSkipped = false, false, false
	m.list.Render(game.Controls[*State](), target, timeDelta)

	scrollVPadding := 12
	scrollProgression, scrollRatio := m.list.ScrollPosition()
	scrollWidth := primortalListRightWidth - primortalListMargin
	scrollFrameHeight := screenHeight - primortalListMargin*2 - scrollVPadding*2
	scrollIndicatorHeight := int(float64(scrollFrameHeight) * scrollRatio)
	indicatorDy := int(scrollProgression * float64(scrollFrameHeight-scrollIndicatorHeight))
	scrollTopRight := pixel.IM.Moved(gfx.IVec(screenWidth-primortalListMargin, screenHeight-primortalListMargin-scrollVPadding))
	frame1px.Draw(target, rect(scrollWidth, scrollFrameHeight),
		scrollTopRight,
		frames.WithColor(screenColors.Dark), frames.WithRenderOrigin(gfx.TopRight))
	frame2px.Draw(target, rect(scrollWidth, scrollIndicatorHeight),
		scrollTopRight.Moved(gfx.IVec(0, -indicatorDy)),
		frames.WithColor(screenColors.Text), frames.WithRenderOrigin(gfx.TopRight))

	smallText := newTextRenderer(target, smallTextbox)
	frb := rect(62+2, tooltipMargin+10)
	fr := rect(int(frb.W()+1), int(frb.H()+1))
	if m.upgradeAbove {
		fm := pixel.IM.Moved(gfx.IVec(screenWidth+2, screenHeight+2))
		frame2px.Draw(target, fr, fm, frames.WithColor(screenColors.Dark), frames.WithRenderOrigin(gfx.TopRight))
		frame2pxBorder.Draw(target, frb, fm, frames.WithColor(screenColors.Text), frames.WithRenderOrigin(gfx.TopRight))
		dx, _ := smallText.render("UPGRADE Above", screenWidth-tooltipMargin, screenHeight-tooltipMargin, flashingHighlight(), tbcfg.RenderFrom(gfx.TopRight))
		arrowUp.Draw(target, pixel.IM.Moved(gfx.IVec(screenWidth-dx-tooltipMargin*2, screenHeight-tooltipMargin)).Moved(gfx.TopRight.Align(arrowUp)))
	}
	if m.upgradeBelow {
		fm := pixel.IM.Moved(gfx.IVec(screenWidth+2, -2))
		frame2px.Draw(target, fr, fm, frames.WithColor(screenColors.Dark), frames.WithRenderOrigin(gfx.BottomRight))
		frame2pxBorder.Draw(target, frb, fm, frames.WithColor(screenColors.Text), frames.WithRenderOrigin(gfx.BottomRight))
		dx, _ := smallText.render("UPGRADE Below", screenWidth-tooltipMargin, tooltipMargin, flashingHighlight(), tbcfg.RenderFrom(gfx.BottomRight))
		arrowDown.Draw(target, pixel.IM.Moved(gfx.IVec(screenWidth-dx-tooltipMargin*2, tooltipMargin)).Moved(gfx.BottomRight.Align(arrowDown)))
	}
	game.DebugBLf("upgrade above: %v, upgrade below: %v, skipped :%v", m.upgradeAbove, m.upgradeBelow, m.wasSkipped)
}

func (m *primortalsMenu) ScrollToNextUpgrade() {
	for id, item := range m.list.items {
		if p, ok := item.(*primortalListItem); ok && p.upgradeAvailable() {
			m.list.highlight(id)
			return
		}
	}
}

func cheapestAvailableUnlock(primortal rpg.Primortal, save *rpg.GameSave) *rpg.UnlockableSkill {
	var cheapestUnlock *rpg.UnlockableSkill
	cheapestCost := int(^uint(0) >> 1) // max int

	for skillId, us := range primortal.UnlockableSkills {
		if save.IsSkillUnlocked(skillId) {
			continue
		}

		// Check if prerequisites are met
		prerequisitesMet := true
		for _, prereqId := range us.Prerequisites {
			if !save.IsSkillUnlocked(prereqId) {
				prerequisitesMet = false
				break
			}
		}

		if prerequisitesMet && us.Cost < cheapestCost {
			cheapestCost = us.Cost
			usCopy := us
			cheapestUnlock = &usCopy
		}
	}

	return cheapestUnlock
}
