package xenolog

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	skillSetRowHeight = 14
)

type skillSetMenu struct {
	selection       int
	currentSkillSet *rpg.SkillSet
	actions         []skillSetAction
	elapsed         float64
}

type skillSetAction struct {
	label   string
	handler func(ctx Context, s *skillSetMenu)
}

func newSkillSetMenu(ctx Context) *skillSetMenu {
	return &skillSetMenu{
		selection:       0,
		currentSkillSet: &ctx.GameSave.Animech.SkillSet,
		actions: []skillSetAction{
			{
				label: "Save",
				handler: func(ctx Context, s *skillSetMenu) {
					ctx.Notify("todo - save")
				},
			},
			{
				label: "Load",
				handler: func(ctx Context, s *skillSetMenu) {
					ctx.Notify("todo - load")
				},
			},
		},
	}
}

func (*skillSetMenu) Enter(Context) {}

func (m *skillSetMenu) OnTick(ctx Context, target pixel.Target, timeDelta float64) {
	m.elapsed += timeDelta

	m.handleInput(ctx)
	m.drawCurrentSkillSet(ctx, target)
}

func (m *skillSetMenu) handleInput(ctx Context) {
	if ctx.Controls.ButtonB().JustPressed() {
		ctx.PopMenu()
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Up) {
		m.selection--
		if m.selection < 0 {
			m.selection = 0
		}
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Down) {
		m.selection++
		maxSelection := 4 + len(m.actions) - 1
		if m.selection > maxSelection {
			m.selection = maxSelection
		}
	}
	if ctx.Controls.ButtonA().JustPressed() {
		// todo
	}
}

func (m *skillSetMenu) drawCurrentSkillSet(ctx Context, target pixel.Target) {
	smallText := newTextRenderer(ctx, target, smallTextbox)
	y := screenHeight - skillSetRowHeight
	skills := []rpg.SkillId{
		m.currentSkillSet.Skill1,
		m.currentSkillSet.Skill2,
		m.currentSkillSet.Skill3,
		m.currentSkillSet.Skill4,
	}
	for id, skill := range skills {
		var name string
		if skill == rpg.UnsetSkillId {
			name = "<empty>"
		} else {
			name = skill.Get().Name
		}
		cursor := "- "
		mask := uiMask
		if id == m.selection {
			cursor = "> "
			name = "{+u}" + name + "{-cu}"
			mask = colors.White.RGBA
		}
		smallText.render(cursor+name, 10, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		y -= skillSetRowHeight
	}
	for id, action := range m.actions {
		mask := uiMask
		label := action.label
		if id == m.selection-len(skills) {
			mask = colors.White.RGBA
			label = ">> " + label + " <<"
		}
		smallText.render(label, 10, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		y -= skillSetRowHeight
	}
}
