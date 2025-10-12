package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type primortalDetailMenu struct {
	screen    *Screen
	primortal rpg.PrimortalType
	selection int
}

func newPrimortalDetailMenu(screen *Screen, primortal rpg.PrimortalType) *primortalDetailMenu {
	return &primortalDetailMenu{
		screen:    screen,
		primortal: primortal,
	}
}

var (
	detailLineHeight   = 14
	detailMargin       = 8
	descriptionTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth-48-detailMargin*2-primortalIconMargin*3, 8,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(maskText),
			tbcfg.RenderFrom(gfx.TopCenter)))
	primortalIconMargin = 4
)

func (*primortalDetailMenu) Enter() {}

func (v *primortalDetailMenu) OnTick(target pixel.Target, timeDelta float64) {
	v.handleInput()

	isComplete := nextUnlock(v.primortal.Primortal(), game.CurrentSave()) == nil

	titleTxt := newTextRenderer(target, titleTextbox)
	regularTxt := newTextRenderer(target, regularTextbox)
	smallTxt := newTextRenderer(target, smallTextbox)
	descriptionText := newTextRenderer(target, descriptionTextbox)
	p := v.primortal.Primortal()
	y := screenHeight - detailMargin

	rp := 0
	if savedPrimortal, exists := game.CurrentSave().Primortals[v.primortal]; exists {
		rp = savedPrimortal.ResearchPoints
	}

	dx, _ := titleTxt.render(p.Name, 10, y, maskTextHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	if isComplete {
		smallTxt.render("[complete]", 10+dx+5, y-4, maskDark, tbcfg.RenderFrom(gfx.TopLeft))
	}
	y -= detailLineHeight

	_, descHeight := descriptionText.render(p.Description, 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.HAligned(tbcfg.HAlignLeft))
	y -= descHeight + detailLineHeight

	titleTxt.render("Skills:", 10, y, maskTextHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	y -= detailLineHeight

	dx, _ = smallTxt.render("Research Points:", 10, y, maskText, tbcfg.RenderFrom(gfx.TopLeft))
	smallTxt.render(fmt.Sprintf("%d", rp), 10+dx+3, y, maskTextHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	y -= detailLineHeight

	y -= 2
	seenLocked := false
	bottomRightBadge := &badgeAViewDetails
	for id, us := range p.UnlockableSkills {
		selected := id == v.selection
		mask := maskText

		name := "???"
		if !seenLocked {
			name = us.SkillId.Get().Name
		}
		if selected {
			mask = maskTextHighlight
		}

		suffix := ""
		suffixMask := maskText
		if seenLocked {
			suffix = "[locked]"
			if selected {
				bottomRightBadge = nil
			}
		} else if game.CurrentSave().IsSkillUnlocked(us.SkillId) {
			suffix = "[unlocked]"
		} else {
			if selected {
				bottomRightBadge = &badgeAUnlock
			}
			suffixMask = flashingHighlight()
			suffix = fmt.Sprintf("> unlock for: %d RP <", us.Cost)
			seenLocked = true
		}

		if selected {
			scrollCursor.Draw(target, pixel.IM.Moved(gfx.IVec(10, y+1)).Moved(gfx.LeftCenter.Align(scrollCursor)))
		}

		x := 25
		dx, _ := regularTxt.render(name, x, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		x += dx + 9

		if suffix != "" {
			smallTxt.render(suffix, x, y, suffixMask, tbcfg.RenderFrom(gfx.LeftCenter))
		}

		y -= detailLineHeight
	}

	icon := anim.Load(atlas, "primortals/"+string(p.Type), "default").Sprite()
	iconMatrix := pixel.IM.Moved(gfx.TopRight.Align(icon)).Moved(gfx.IVec(screenWidth-detailMargin-primortalIconMargin, screenHeight-detailMargin-primortalIconMargin))
	w := int(icon.Bounds().W()) + primortalIconMargin*2
	h := int(icon.Bounds().H()) + primortalIconMargin*2
	frame4px.Draw(target, rect(w, h), iconMatrix, frames.WithColor(maskDark))
	frame4pxBorder.Draw(target, rect(w, h), iconMatrix, frames.WithColor(maskText))
	v.screen.spriteShader.DrawSprite(icon, iconMatrix)

	badgeMargin := 2
	badgeF1Reset.Render(target, pixel.IM.Moved(gfx.IVec(badgeMargin, badgeMargin)), gfx.BottomLeft)
	if bottomRightBadge != nil {
		(*bottomRightBadge).Render(target, pixel.IM.Moved(gfx.IVec(screenWidth-badgeMargin, badgeMargin)), gfx.BottomRight)
	}
}

func (v *primortalDetailMenu) handleInput() {
	if game.Controls[*State]().ButtonB().JustPressed() {
		v.screen.PopMenu()
	}
	if game.Controls[*State]().DPad().DirectionJustPressedOrRepeated(input.Up) {
		v.selection--
		if v.selection < 0 {
			v.selection = 0
		}
	}
	if game.Controls[*State]().DPad().DirectionJustPressedOrRepeated(input.Down) {
		v.selection++
		maxSelection := len(v.primortal.Primortal().UnlockableSkills) - 1
		if v.selection > maxSelection {
			v.selection = maxSelection
		}
	}
	if game.Controls[*State]().ButtonA().JustPressed() {
		us := v.primortal.Primortal().UnlockableSkills[v.selection]
		savedPrimortal, savedPrimortalExists := game.CurrentSave().Primortals[v.primortal]
		if game.CurrentSave().IsSkillUnlocked(us.SkillId) {
			v.screen.PushMenu(newSkillDetailMenu(v.screen, us.SkillId))
		} else if savedPrimortalExists && us.Cost <= savedPrimortal.ResearchPoints {
			game.CurrentSave().Primortals[v.primortal].ResearchPoints -= us.Cost
			game.CurrentSave().UnlockSkill(us.SkillId)
			saveOrNotify()
		}
	}
	if game.DebugToggles().F1().JustPressed() {
		for _, us := range v.primortal.Primortal().UnlockableSkills {
			if game.CurrentSave().IsSkillUnlocked(us.SkillId) {
				game.CurrentSave().RemoveUnlockedSkill(us.SkillId)
				if _, exists := game.CurrentSave().Primortals[v.primortal]; !exists {
					game.CurrentSave().Primortals[v.primortal] = &rpg.PrimortalProgress{}
				}
				game.CurrentSave().Primortals[v.primortal].ResearchPoints += us.Cost
			}
		}
		saveOrNotify()
	}
}

func rect(w, h int) pixel.Rect {
	return pixel.R(0, 0, float64(w), float64(h))
}
