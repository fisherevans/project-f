package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type primortalDetailMenu struct {
	screen    *Screen
	primortal rpg.PrimortalType
	tree      *skillTree
}

func newPrimortalDetailMenu(screen *Screen, primortal rpg.PrimortalType) *primortalDetailMenu {
	m := &primortalDetailMenu{
		screen:    screen,
		primortal: primortal,
	}
	m.tree = newSkillTree(primortal, m)
	return m
}

// Variables declared in init_vars.go
var (
	detailLineHeight               = 14
	detailMargin                   = 8
	primortalIconSize              = 48
	primortalIconMargin            = 2
	primortalSkillDetailWidth      = 64
	primortalSkillDetailMargin     = 2
	primortalSkillDetailTextMargin = 5
)

func (*primortalDetailMenu) Enter() {}

func (v *primortalDetailMenu) OnTick(target pixel.Target, timeDelta float64) {
	v.handleInput()

	isComplete := cheapestAvailableUnlock(v.primortal.Primortal(), game.CurrentSave()) == nil

	titleTxt := newTextRenderer(target, titleTextbox)
	smallTxt := newTextRenderer(target, smallTextbox)
	descriptionText := newTextRenderer(target, primortalDescriptionTextbox)
	p := v.primortal.Primortal()
	y := screenHeight - detailMargin

	rp := 0
	if savedPrimortal, exists := game.CurrentSave().Primortals[v.primortal]; exists {
		rp = savedPrimortal.ResearchPoints
	}

	// top right icon
	icon := anim.LoadTilesheetAnimation(atlas, "primortals/"+string(p.Type), "default").Sprite()
	iconMatrix := pixel.IM.Moved(gfx.TopRight.Align(icon)).Moved(gfx.IVec(screenWidth-detailMargin-primortalIconMargin, screenHeight-detailMargin-primortalIconMargin))
	w := int(icon.Bounds().W()) + primortalIconMargin*2
	h := int(icon.Bounds().H()) + primortalIconMargin*2
	frame4px.Draw(target, rect(w, h), iconMatrix, frames.WithColor(colorDark))
	frame4pxBorder.Draw(target, rect(w, h), iconMatrix, frames.WithColor(colorText))
	icon.Draw(v.screen.spriteFilterBuffer.Target(), iconMatrix)

	// header
	dx, _ := titleTxt.render(p.Name, 10, y, colorHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	if isComplete {
		smallTxt.render("[complete]", 10+dx+5, y-4, colorDark, tbcfg.RenderFrom(gfx.TopLeft))
	}
	y -= detailLineHeight

	_, descHeight := descriptionText.render(p.Description, 10, y, colorText, tbcfg.RenderFrom(gfx.TopLeft), tbcfg.HAligned(tbcfg.HAlignLeft))
	y -= descHeight + detailLineHeight/2

	titleTxt.render("Skills:", 10, y, colorHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	y -= detailLineHeight

	dx, _ = smallTxt.render("Research Points:", 10, y, colorText, tbcfg.RenderFrom(gfx.TopLeft))
	smallTxt.render(fmt.Sprintf("%d", rp), 10+dx+3, y, colorHighlight, tbcfg.RenderFrom(gfx.TopLeft))
	y -= detailLineHeight

	// tree
	treeW := screenWidth - detailMargin*3 - primortalSkillDetailWidth
	treeH := y - detailMargin
	treeDx := (treeW - int(v.tree.frame.W())) / 2
	treeDy := (treeH - int(v.tree.frame.H())) / 2
	costYDeltaCompensation := -3
	treeBottomLeft := pixel.IM.Moved(gfx.IVec(detailMargin+treeDx, detailMargin+treeDy+costYDeltaCompensation))
	v.tree.renderTree(target, treeBottomLeft, timeDelta)

	// skill details
	skillDetailsTopLeftY := screenHeight - detailMargin*2 - primortalIconSize - primortalIconMargin*2
	skillDetailsTopLeft := pixel.IM.Moved(gfx.IVec(
		screenWidth-detailMargin-primortalSkillDetailWidth,
		skillDetailsTopLeftY))
	primortalSkillDetailHeight := skillDetailsTopLeftY - detailMargin
	frame2pxBorder.Draw(target, rect(primortalSkillDetailWidth, primortalSkillDetailHeight), skillDetailsTopLeft,
		frames.WithRenderOrigin(gfx.TopLeft), frames.WithColor(colorDark))
	v.renderSkillDetailsText(target, skillDetailsTopLeft, primortalSkillDetailWidth, primortalSkillDetailHeight)

	// badges
	badgeMargin := 2
	badgeF1Reset.Render(target, pixel.IM.Moved(gfx.IVec(screenWidth/2, screenHeight-badgeMargin)), gfx.TopCenter)
}

func (v *primortalDetailMenu) renderSkillDetailsText(target pixel.Target, topLeft pixel.Matrix, width int, height int) {
	topCenter := topLeft.Moved(gfx.IVec(width/2, 0))
	bottomCenter := topLeft.Moved(gfx.IVec(width/2, -height))

	txt := newTextRenderer(target, primortalSkillDescriptionTextbox)
	txt.matrix = topCenter

	skill := v.tree.selectedSkill
	state := v.tree.nodes[skill].getState()
	y := -primortalSkillDetailTextMargin

	titleLabel := "?????"
	desc := "unlock prior skills to unlock this one"
	if state != skillNodeStateHidden {
		titleLabel = skill.Get().Name
		if state == skillNodeStateUnlockable {
			desc = "unlock this skill to see details"
		} else {
			desc = skill.Get().Description
		}
	}

	_, dy := txt.render("{+s:xenolog_dark,+u:xenolog_text}"+titleLabel, 0, y, colorHighlight)
	y -= primortalSkillDetailTextMargin + dy + 2 // extra for underline

	_, dy = txt.render(desc, 0, y, colorText)
	y -= primortalSkillDetailTextMargin + dy

	var aBadge *badges.ButtonAction
	if state == skillNodeStateUnlockable {
		progress, hasProgress := game.CurrentSave().Primortals[v.primortal]
		unlockableSkill := v.primortal.Primortal().UnlockableSkills[skill]
		if hasProgress && unlockableSkill.Cost <= progress.ResearchPoints {
			label := fmt.Sprintf("cost: %d rp", unlockableSkill.Cost)
			txt.render(label, 0, y, flashingHighlight())
			aBadge = badgeAUnlock
		}
	} else if state == skillNodeStateUnlocked {
		aBadge = badgeADetails
	}
	if aBadge != nil {
		aBadge.Render(target, bottomCenter.Moved(gfx.IVec(0, 1)), gfx.Centered)
	}
}

func (v *primortalDetailMenu) handleInput() {
	c := game.Controls[*State]()
	if c.ButtonB().JustPressed() {
		v.screen.PopMenuAnimated()
	}
	if c.DPad().JustPressed() {
		v.tree.handleInput(c.DPad().JustPressedDirection())
	}
	if c.ButtonA().JustPressed() {
		skillId := v.tree.selectedSkill
		state := v.tree.nodes[skillId].getState()
		if state == skillNodeStateUnlocked {
			v.screen.PushMenuAnimated(newSkillDetailMenu(v.screen, skillId))
		} else if state == skillNodeStateUnlockable {
			progress, hasProgress := game.CurrentSave().Primortals[v.primortal]
			unlockableSkill := v.primortal.Primortal().UnlockableSkills[skillId]
			if hasProgress && unlockableSkill.Cost <= progress.ResearchPoints {
				game.CurrentSave().Primortals[v.primortal].ResearchPoints -= unlockableSkill.Cost
				game.CurrentSave().UnlockSkill(skillId)
				saveOrNotify()
				v.tree.nodes[skillId].generateParticles()
				//modal := newMenuModal(v.screen,
				//	skillId.Get().Name+" Unlocked!",
				//	fmt.Sprintf("You spent %d research points to unlock %s.", unlockableSkill.Cost, skillId.Get().Name),
				//	modalOption{
				//		label: "View Skill",
				//		onSelect: func() {
				//			v.screen.PushMenuAnimated(newSkillDetailMenu(v.screen, skillId))
				//		},
				//	},
				//	modalOption{
				//		label: "Okay",
				//	},
				//)
				//v.screen.PushMenu(modal, false)
			} else {
				// not enough rp
			}
		} else {
			// pre reqs not met
		}
	}
	if game.DebugToggles().F1().JustPressed() {
		for skillId, us := range v.primortal.Primortal().UnlockableSkills {
			if game.CurrentSave().IsSkillUnlocked(skillId) {
				game.CurrentSave().RemoveUnlockedSkill(skillId)
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
