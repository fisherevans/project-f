package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type skillsMenu struct {
	screen     *Screen
	list       *scrollList[*skillsMenu]
	onComplete func(id rpg.SkillId)
}

func newSkillsMenu(screen *Screen, highlighted rpg.SkillId, onComplete func(id rpg.SkillId)) *skillsMenu {
	menu := &skillsMenu{
		screen: screen,
		list: newScrollList[*skillsMenu](scrollListOptions{
			targetHeight:        screenHeight,
			transitionTime:      0.15,
			verticalItemMargin:  1,
			verticalListPadding: 8,
		}),
		onComplete: onComplete,
	}
	for _, skillId := range game.CurrentSave().UnlockedSkillsSorted() {
		menu.list.Add(&skillListItem{
			skillId: skillId,
		})
		if skillId == highlighted {
			menu.list.highlightLast()
		}
	}
	menu.list.cursor = &skillsCursor{}
	return menu
}

func (*skillsMenu) Enter() {}

type skillListItem struct {
	baseListItem[*skillsMenu]
	skillId rpg.SkillId
}

func (p *skillListItem) ButtonAJustPressed(m *skillsMenu) {
	m.onComplete(p.skillId)
	m.screen.PopMenu()
}

func (p *skillListItem) ButtonStartJustPressed(m *skillsMenu) {
	m.screen.PushMenu(newSkillDetailMenu(m.screen, p.skillId))
}

func (p *skillListItem) Height() int {
	return 12
}

type skillsCursor struct{}

func (c *skillsCursor) Render(m *skillsMenu, centerLeft int, target pixel.Target, index int, movementProgress, highlightProgress float64) {
	smallText := newTextRenderer(target, smallTextbox)
	smallText.render(">", 10, centerLeft, maskTextHighlight, tbcfg.RenderFrom(gfx.LeftCenter))
}

func (p *skillListItem) Render(m *skillsMenu, topLeftY int, target pixel.Target, highlightProgress float64) {
	centerY := topLeftY - p.Height()/2
	smallText := newTextRenderer(target, smallTextbox)

	mask := maskText
	if highlightProgress > 0 {
		mask = colors.Lerp(mask, maskTextHighlight, highlightProgress)
	}
	smallText.render(p.skillId.Get().Name, 35, centerY, mask, tbcfg.RenderFrom(gfx.LeftCenter))
}

func (m *skillsMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu()
	}
	m.list.Render(m, game.Controls[*State](), target, timeDelta)
}
