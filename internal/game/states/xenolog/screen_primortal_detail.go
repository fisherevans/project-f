package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type primortalDetailView struct {
	primortal rpg.PrimortalType
	selection int
}

func newPrimortalDetailView(primortal rpg.PrimortalType) *primortalDetailView {
	return &primortalDetailView{
		primortal: primortal,
	}
}

var (
	detailLineHeight   = 14
	detailMargin       = 8
	descriptionTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth-48-detailMargin*2, 8,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(uiMask),
			tbcfg.RenderFrom(gfx.TopCenter)))
)

func (v *primortalDetailView) RenderPrimortalView(ps *primortalsMenu, ctx *game.Context, target *pixel.Batch, timeDelta float64) {
	v.handleInput(ctx, ps)

	mediumText := newTextRenderer(ctx, target, mediumTextbox)
	smallText := newTextRenderer(ctx, target, smallTextbox)
	descriptionText := newTextRenderer(ctx, target, descriptionTextbox)
	p := v.primortal.Primortal()
	y := screenHeight - detailMargin

	rp := 0
	if savedPrimortal, exists := ctx.GameSave.Primortals[v.primortal]; exists {
		rp = savedPrimortal.ResearchPoints
	}

	mediumText.render(p.Name, 10, y, uiMask, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(colors.White.RGBA))
	y -= detailLineHeight

	_, descHeight := descriptionText.render(p.Description, 10, y, uiMask, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.HAligned(tbcfg.HAlignLeft))
	y -= descHeight + detailLineHeight

	mediumText.render(fmt.Sprintf("Research Points: %d", rp), 10, y, uiMask, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(colors.White.RGBA))
	y -= detailLineHeight

	mediumText.render("Skills:", 10, y, uiMask, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.Foreground(colors.White.RGBA))
	y -= detailLineHeight

	seenLocked := false
	for id, us := range p.UnlockableSkills {
		mask := uiMask
		cursor := "- "

		name := "???"
		if !seenLocked {
			name = us.SkillId.Get().Name
		}
		if id == v.selection {
			mask = colors.White.RGBA
			cursor = "> "
			name = "{+u}" + name + "{-u}"
		}

		suffix := ""
		if seenLocked {
			suffix = " (locked)"
		} else if ctx.GameSave.IsSkillUnlocked(us.SkillId) {
			suffix = " (unlocked)"
		} else {
			suffix = fmt.Sprintf(" {+c:white}>> %d RP <<{-c}", us.Cost)
			seenLocked = true
		}

		smallText.render(cursor+name+suffix, 10, y, mask, tbcfg.RenderFrom(gfx.TopLeft))
		y -= detailLineHeight
	}

	icon := anim.Load(atlas, "primortals/"+string(p.Type), "default").Sprite()
	icon.Draw(target, pixel.IM.Moved(gfx.TopRight.Align(icon)).Moved(gfx.IVec(screenWidth-detailMargin, screenHeight-detailMargin)))

	smallText.render("[F1] to reset", screenWidth-detailMargin, detailMargin, uiMask, tbcfg.RenderFrom(gfx.BottomRight))
}

func (v *primortalDetailView) handleInput(ctx *game.Context, ps *primortalsMenu) {
	if ctx.Controls.ButtonB().JustPressed() {
		ps.returnToList(ctx)
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Up) {
		v.selection--
		if v.selection < 0 {
			v.selection = 0
		}
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Down) {
		v.selection++
		maxSelection := len(v.primortal.Primortal().UnlockableSkills) - 1
		if v.selection > maxSelection {
			v.selection = maxSelection
		}
	}
	if ctx.Controls.ButtonA().JustPressed() {
		us := v.primortal.Primortal().UnlockableSkills[v.selection]
		if !ctx.GameSave.IsSkillUnlocked(us.SkillId) && us.Cost <= ctx.GameSave.Primortals[v.primortal].ResearchPoints {
			ctx.GameSave.Primortals[v.primortal].ResearchPoints -= us.Cost
			ctx.GameSave.UnlockedSkills[us.SkillId] = struct{}{}
		}
		saveOrNotify(ctx)
	}
	if ctx.DebugToggles.F1().JustPressed() {
		for _, us := range v.primortal.Primortal().UnlockableSkills {
			if ctx.GameSave.IsSkillUnlocked(us.SkillId) {
				delete(ctx.GameSave.UnlockedSkills, us.SkillId)
				if _, exists := ctx.GameSave.Primortals[v.primortal]; !exists {
					ctx.GameSave.Primortals[v.primortal] = &rpg.PrimortalProgress{}
				}
				ctx.GameSave.Primortals[v.primortal].ResearchPoints += us.Cost
			}
		}
		saveOrNotify(ctx)
	}
}

func saveOrNotify(ctx *game.Context) {
	if err := ctx.GameSave.Save(); err != nil {
		log.Err(err).Msg("failed to save game")
		ctx.Notify("Failed to save game!!!")
	}
}
