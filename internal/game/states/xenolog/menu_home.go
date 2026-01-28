package xenolog

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

// Variables declared in init_vars.go
var (
	labelUnknown = "???????"
	
	selectBoxWidth  = 74
	selectBoxHeight = 82
	selectBoxLabelMargin = 3
	selectBoxSubLabelFlashSpeed = 1.0
	selectBoxSubLabelMargin = 3
	selectBoxArrowMargin = 2.0
)

type homeMenu struct {
	screen     *Screen
	selectLeft bool
	left       *selectBox
	right      *selectBox
}

func newHomeMenu(screen *Screen, primortalsEnabled bool) *homeMenu {
	home := &homeMenu{
		screen:     screen,
		selectLeft: true,
		left:       newSelectBox(spriteAnimech, "Animech", ""),
		right:      newSelectBox(spriteUnknown, labelUnknown, "[locked]"),
	}
	if primortalsEnabled {
		home.right = newSelectBox(spritePrimortal, "Primortals", "")
	}
	return home
}

func (s *homeMenu) Enter() {
	s.left.subLabel = ""
	if isAnimechUpgradeAvailable() {
		s.left.subLabel = "Upgrade Available"
	}

	s.right.subLabel = ""
	if s.right.label != labelUnknown && isPrimortalUpgradeAvailable() {
		s.right.subLabel = "Upgrade Available"
	}
}

func isAnimechUpgradeAvailable() bool {
	a := game.CurrentSave().Animech
	level := a.Upgrades.GetLevel()
	return a.AnimechExperience >= rpg.AnimechUpgradeExperienceRequiredToUpgrade(level+1)
}

func isPrimortalUpgradeAvailable() bool {
	for pType, save := range game.CurrentSave().Primortals {
		next := cheapestAvailableUnlock(pType.Primortal(), game.CurrentSave())
		if next != nil && save.ResearchPoints >= next.Cost {
			return true
		}
	}
	return false
}

func (s *homeMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().DPad().JustPressedDirection() == input.Left {
		s.selectLeft = true
	} else if game.Controls[*State]().DPad().JustPressedDirection() == input.Right {
		s.selectLeft = false
	}

	if game.Controls[*State]().ButtonA().JustPressed() {
		if s.selectLeft {
			s.screen.PushMenuAnimated(newAnimechMenu(s.screen))
		} else if s.right.label != labelUnknown {
			s.screen.PushMenuAnimated(newPrimortalsMenu(s.screen))
		} else {
			game.GetAudioSystem().PlaySFX("adventure/beeps/error", 1.0)
		}
	}

	if game.Controls[*State]().ButtonB().JustPressed() {
		s.screen.state.Close()
	}

	center := pixel.IM.Moved(gfx.IVec(screenWidth/2, screenHeight/2+3))
	dx := math.Floor(float64(selectBoxWidth) * 0.6)
	s.left.render(center.Moved(pixel.V(-dx, 0)), target, s.selectLeft, timeDelta)
	s.right.render(center.Moved(pixel.V(dx, 0)), target, !s.selectLeft, timeDelta)

	badgeStartClose.Render(target, pixel.IM.Moved(gfx.IVec(2, 2)), gfx.BottomLeft)
	badgeASelect.Render(target, pixel.IM.Moved(gfx.IVec(screenWidth-2, 2)), gfx.BottomRight)
}

var (
	selectBoxArrowBounceHeight = 3.0
	selectBoxArrowBounceSpeed  = 1.0
)

type selectBox struct {
	sprite   pixelutil.BoundedDrawable
	label    string
	subLabel string

	elapsed float64
}

func (b *selectBox) Bounds() pixel.Rect {
	return pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
}

func newSelectBox(sprite pixelutil.BoundedDrawable, label, subLabel string) *selectBox {
	s := &selectBox{
		sprite:   sprite,
		label:    label,
		subLabel: subLabel,
	}
	return s
}

func (b *selectBox) render(center pixel.Matrix, target pixel.Target, selected bool, timeDelta float64) {
	b.elapsed += timeDelta

	mask := colorText
	if selected {
		mask = colorHighlight
	}

	frameRect := pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
	frame4px.Draw(target, frameRect, center, frames.WithColor(colorDark))
	frame4pxBorder.Draw(target, frameRect, center, frames.WithColor(mask))

	bottomCenter := center.Moved(gfx.IVec(0, -selectBoxHeight/2))
	spriteCenter := bottomCenter.Moved(gfx.BottomCenter.Align(b.sprite))
	b.sprite.DrawColorMask(target, spriteCenter, mask)

	if b.label != "" {
		c := selectBoxLabelText.NewSimpleContent(b.label)
		c.Render(target, center.Moved(gfx.IVec(0, selectBoxHeight/2-selectBoxLabelMargin)), tbcfg.Foreground(mask))
	}

	if b.subLabel != "" {
		subLabelMask := colors.Lerp(colorText, colorHighlight, game.Utils().TimeCycleSin(selectBoxSubLabelFlashSpeed))
		c := selectBoxSubLabelText.NewSimpleContent(b.subLabel)
		c.Render(target,
			center.Moved(gfx.IVec(0, -selectBoxHeight/2-selectBoxSubLabelMargin)),
			tbcfg.Foreground(subLabelMask))
	}

	if selected {
		phase := b.elapsed * selectBoxArrowBounceSpeed * 2 * math.Pi
		dy := math.Abs(math.Sin(phase)) * selectBoxArrowBounceHeight
		arrowBottomCenter := center.Moved(gfx.IVec(0, selectBoxHeight/2)).
			Moved(pixel.V(0, selectBoxArrowMargin+dy))
		selectBoxArrow.Draw(target, arrowBottomCenter.Moved(gfx.BottomCenter.Align(selectBoxArrow)))
	}
}
