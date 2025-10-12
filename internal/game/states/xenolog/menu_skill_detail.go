package xenolog

import (
	"fmt"
	"strings"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type skillDetailMenu struct {
	screen    *Screen
	skillId   rpg.SkillId
	selection int
}

func newSkillDetailMenu(screen *Screen, skillId rpg.SkillId) *skillDetailMenu {
	return &skillDetailMenu{
		screen:  screen,
		skillId: skillId,
	}
}

func (*skillDetailMenu) Enter() {}

func (v *skillDetailMenu) OnTick(target pixel.Target, timeDelta float64) {
	v.handleInput()
	skill := v.skillId.Get()
	smallText := newTextRenderer(target, smallTextbox)
	lineHeight := 12
	y := screenHeight - lineHeight

	smallText.render(fmt.Sprintf("{+u}Skill: %s", skill.Name), 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(maskTextHighlight))
	y -= lineHeight

	smallText.render(skill.Description, 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(maskText))
	y -= lineHeight

	smallText.render("Ticks:", 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(maskText))
	y -= lineHeight

	for _, tick := range skill.Ticks {
		var effects []string
		for _, effect := range tick.Effects {
			if effect.Status != nil {
				effects = append(effects, fmt.Sprintf("apply %s (%.1f stacks)", effect.Status.Status, effect.Status.Stacks))
			}
			if effect.Damage != nil {
				effects = append(effects, fmt.Sprintf("%d damage", effect.Damage.Amount))
			}
		}
		tickText := "<idle>"
		if len(effects) > 0 {
			tickText = strings.Join(effects, ", ")
		}
		smallText.render("- "+tickText, 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(maskText))
		y -= lineHeight
	}
}

func (v *skillDetailMenu) handleInput() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		v.screen.PopMenu()
	}
}
