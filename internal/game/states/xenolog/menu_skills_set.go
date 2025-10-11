package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/navigtion"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillSetRowHeight = 14
)

type skillSetInputMode int

const (
	SkillSetInputModeOver skillSetInputMode = iota
	SkillSetInputModeSelect
)

type skillSetMenu struct {
	screen               *Screen
	inputMode            skillSetInputMode
	currentSkillSelected input.Direction
	nav                  *navigtion.System[*skillSetMenu]
	currentSkillSetItem  *navigtion.SimpleItem[*skillSetMenu]
	actions              []*skillSetAction
}

func newSkillSetMenu(screen *Screen) *skillSetMenu {
	m := &skillSetMenu{
		screen: screen,
		nav:    navigtion.NewSystem[*skillSetMenu](),
		currentSkillSetItem: navigtion.NewSimpleItem[*skillSetMenu]().
			WithButtonAHandler(func(m *skillSetMenu) {
				m.inputMode = SkillSetInputModeSelect
			}),
		currentSkillSelected: input.Up,
		actions: []*skillSetAction{
			newSkillSetAction("Save Preset", func(s *skillSetMenu) {
				game.DebugNotification("todo - save")
			}),
			newSkillSetAction("Load Preset", func(s *skillSetMenu) {
				game.DebugNotification("todo - load")
			}),
		},
	}
	m.nav.AddItem(m.currentSkillSetItem, m)
	m.nav.SetLastItem(nil)
	for _, a := range m.actions {
		m.nav.AddNextToLast(input.Right, m, a).Below(m.currentSkillSetItem)
	}
	return m
}

type skillSetAction struct {
	navigtion.BaseItem[*skillSetMenu]
	label   string
	handler func(s *skillSetMenu)
}

func newSkillSetAction(label string, handler func(s *skillSetMenu)) *skillSetAction {
	return &skillSetAction{
		label:   label,
		handler: handler,
	}
}

func (a *skillSetAction) OnButtonAJustPressed(s *skillSetMenu) {
	if a.handler == nil {
		return
	}
	a.handler(s)
}

func (*skillSetMenu) Enter() {}

func (m *skillSetMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInput()
	m.render(target)
}

func (m *skillSetMenu) handleInput() {
	switch m.inputMode {
	case SkillSetInputModeOver:
		if game.Controls[*State]().ButtonB().JustPressed() {
			m.screen.PopMenu()
		}
		m.nav.HandleInputs(game.Controls[*State](), m)
	case SkillSetInputModeSelect:
		if game.Controls[*State]().ButtonB().JustPressed() {
			m.inputMode = SkillSetInputModeOver
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
}

func (m *skillSetMenu) render(target pixel.Target) {
	m.renderCurrentSkills(screenWidth/2, screenHeight/3*2, target)
	dx := screenWidth / (len(m.actions) + 1)
	x := dx
	for _, action := range m.actions {
		m.renderAction(action.label, m.nav.IsHighlighted(action), x, 10, target)
		x += dx
	}
}

func (m *skillSetMenu) renderAction(label string, highlighted bool, x, y int, target pixel.Target) {
	smallText := newTextRenderer(target, smallTextbox)
	mask := uiMask
	if highlighted {
		mask = colors.White.RGBA
		label = ">> " + label + " <<"
	}
	smallText.render(label, x, y, mask, tbcfg.RenderFrom(gfx.Centered))
}

func (m *skillSetMenu) renderCurrentSkills(x, y int, target pixel.Target) {
	dy := 16
	dx := screenWidth / 4

	mask := uiMask
	if m.inputMode == SkillSetInputModeOver && m.nav.IsHighlighted(m.currentSkillSetItem) {
		mask = uiMaskSelected
	}

	frameW := screenWidth - 20
	frameH := dy * 4
	frameRect := pixel.R(0, 0, float64(frameW), float64(frameH))
	selectBoxFrame.Draw(target, frameRect, pixel.IM.Moved(gfx.IVec(x, y)), frames.WithColor(mask))

	m.renderAction("Edit", m.nav.IsHighlighted(m.currentSkillSetItem), x, y-frameH/2-10, target)

	m.renderCurrentSkill(input.Up, x, y+dy, target)
	m.renderCurrentSkill(input.Right, x+dx, y, target)
	m.renderCurrentSkill(input.Down, x, y-dy, target)
	m.renderCurrentSkill(input.Left, x-dx, y, target)
}

func (m *skillSetMenu) renderCurrentSkill(direction input.Direction, x, y int, target pixel.Target) {
	skillId := game.CurrentSave().Animech.SkillSet.DirectionalSkill(direction)
	smallText := newTextRenderer(target, smallTextbox)
	mask := uiMask
	var label string
	if *skillId == rpg.UnsetSkillId {
		label = "---"
	} else {
		label = skillId.Get().Name
	}
	if m.inputMode == SkillSetInputModeSelect && direction == m.currentSkillSelected {
		label = "{+u}" + label + "{-u}"
		mask = uiMaskSelected
	}
	smallText.render(label, x, y, mask, tbcfg.RenderFrom(gfx.Centered))
}
