package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	tooltipMargin = 2
)

type primortalsMenu struct {
	screen                     *Screen
	list                       *scrollList[*primortalsMenu]
	upgradeAbove, upgradeBelow bool
}

func newPrimortalsMenu(screen *Screen) *primortalsMenu {
	menu := &primortalsMenu{
		screen: screen,
		list: newScrollList[*primortalsMenu](scrollListOptions{
			targetHeight:        screenHeight,
			transitionTime:      0.15,
			verticalItemMargin:  1,
			verticalListPadding: 8,
		}),
	}
	for i := 1; i <= rpg.MaxXenoLogEntryIndex; i++ {
		menu.list.Add(&primortalListItem{
			xenologIndex: i,
		})
	}
	menu.list.cursor = &primortalCursor{}
	return menu
}

func (*primortalsMenu) Enter() {}

type primortalListItem struct {
	xenologIndex int
}

func (p *primortalListItem) RenderWasSkipped(m *primortalsMenu, wasAbove bool) {
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
		m.screen.PushMenu(newPrimortalDetailMenu(m.screen, pType))
	}
}

func (p *primortalListItem) Height() int {
	return 12
}

func (p *primortalListItem) SkipHighlight(m *primortalsMenu) bool {
	_, exists := rpg.XenoLogEntries[p.xenologIndex]
	return !exists
}

type primortalCursor struct{}

func (c *primortalCursor) Render(m *primortalsMenu, centerLeft int, target pixel.Target, isMoving bool) {
	smallText := newTextRenderer(target, smallTextbox)
	smallText.render(">", 10, centerLeft, uiMaskSelected, tbcfg.RenderFrom(gfx.LeftCenter))
}

func (p *primortalListItem) Render(m *primortalsMenu, topLeftY int, target pixel.Target, isHighlighted bool) {
	centerY := topLeftY - p.Height()/2
	smallText := newTextRenderer(target, smallTextbox)

	mask := uiMask
	if isHighlighted {
		mask = uiMaskSelected
		//smallText.render(">", 10, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	}
	smallText.render(fmt.Sprintf("%03d", p.xenologIndex), 20, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	pId, exists := rpg.XenoLogEntries[p.xenologIndex]
	if !exists {
		smallText.render("???", 35, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		return
	}
	primortal := pId.Primortal()
	nameDx, _ := smallText.render(primortal.Name, 35, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	x := 35 + nameDx + 10
	nextSkill := nextUnlock(primortal, game.CurrentSave())
	if nextSkill == nil {
		smallText.render("[COMPLETE]", x, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
	} else if p.upgradeAvailable() {
		smallText.render("> UPGRADE AVAILABLE <", x, centerY, flashingHighlight(), tbcfg.RenderFrom(gfx.LeftCenter))
	}
}

func (p *primortalListItem) upgradeAvailable() bool {
	upgradeAvailable := false
	if pId, ok := rpg.XenoLogEntries[p.xenologIndex]; ok {
		if save, exists := game.CurrentSave().Primortals[pId]; exists {
			if next := nextUnlock(pId.Primortal(), game.CurrentSave()); next != nil && save.ResearchPoints >= next.Cost {
				upgradeAvailable = true
			}
		}
	}
	return upgradeAvailable
}

func (m *primortalsMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu()
	}

	m.upgradeAbove, m.upgradeBelow = false, false
	m.list.Render(m, game.Controls[*State](), target, timeDelta)

	smallText := newTextRenderer(target, smallTextbox)
	if m.upgradeAbove {
		smallText.render("^ UPGRADE", screenWidth-tooltipMargin, screenHeight-tooltipMargin, flashingHighlight(), tbcfg.RenderFrom(gfx.TopRight))
	}
	if m.upgradeBelow {
		smallText.render("v UPGRADE", screenWidth-tooltipMargin, tooltipMargin, flashingHighlight(), tbcfg.RenderFrom(gfx.BottomRight))
	}
}

func nextUnlock(primortal rpg.Primortal, save *rpg.GameSave) *rpg.UnlockableSkill {
	for _, us := range primortal.UnlockableSkills {
		if !save.IsSkillUnlocked(us.SkillId) {
			return &us
		}
	}
	return nil
}
