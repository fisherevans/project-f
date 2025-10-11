package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/navigtion"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillSetRowHeight = 14
)

type skillSetAction struct {
	navigtion.BaseItem[*skillSetMenu]
	label   string
	handler func(s *skillSetMenu)
}

func (a *skillSetAction) OnButtonAJustPressed(s *skillSetMenu) {
	if a.handler == nil {
		return
	}
	a.handler(s)
}

type skillSetSkill struct {
	navigtion.BaseItem[*skillSetMenu]
	skillId rpg.SkillId
}

type skillSetMenu struct {
	screen  *Screen
	nav     *navigtion.System[*skillSetMenu]
	actions []*skillSetAction
	skills  []*skillSetSkill
}

func newSkillSetMenu(screen *Screen) *skillSetMenu {
	m := &skillSetMenu{
		screen: screen,
		nav:    navigtion.NewSystem[*skillSetMenu](),
		actions: []*skillSetAction{
			{
				label: "Save",
				handler: func(s *skillSetMenu) {
					game.DebugNotification("todo - save")
				},
			},
			{
				label: "Load",
				handler: func(s *skillSetMenu) {
					game.DebugNotification("todo - load")
				},
			},
		},
		skills: []*skillSetSkill{
			{
				skillId: game.CurrentSave().Animech.SkillSet.Skill1,
			},
			{
				skillId: game.CurrentSave().Animech.SkillSet.Skill2,
			},
			{
				skillId: game.CurrentSave().Animech.SkillSet.Skill3,
			},
			{
				skillId: game.CurrentSave().Animech.SkillSet.Skill4,
			},
		},
	}
	for _, skill := range m.skills {
		m.nav.AddNextToLast(input.Down, m, skill)
	}
	for _, action := range m.actions {
		m.nav.AddNextToLast(input.Down, m, action)
	}
	return m
}

func (*skillSetMenu) Enter() {}

func (m *skillSetMenu) OnTick(target pixel.Target, timeDelta float64) {
	m.handleInput()
	m.drawCurrentSkillSet(target)
}

func (m *skillSetMenu) handleInput() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenu()
	} else {
		m.nav.HandleInputs(game.Controls[*State](), m)
	}
}

func (m *skillSetMenu) drawCurrentSkillSet(target pixel.Target) {
	smallText := newTextRenderer(target, smallTextbox)
	y := screenHeight - skillSetRowHeight
	for _, skill := range m.skills {
		var name string
		if skill.skillId == rpg.UnsetSkillId {
			name = "<empty>"
		} else {
			name = skill.skillId.Get().Name
		}
		cursor := "- "
		mask := uiMask
		if m.nav.IsHighlighted(skill) {
			cursor = "> "
			name = "{+u}" + name + "{-cu}"
			mask = colors.White.RGBA
		}
		smallText.render(cursor+name, 10, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		y -= skillSetRowHeight
	}
	for _, action := range m.actions {
		mask := uiMask
		label := action.label
		if m.nav.IsHighlighted(action) {
			mask = colors.White.RGBA
			label = ">> " + label + " <<"
		}
		smallText.render(label, 10, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		y -= skillSetRowHeight
	}
}
