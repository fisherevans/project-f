package xenolog

import (
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	spriteAnimech   = atlas.GetSprite("xenolog/select_animech")
	spritePrimortal = atlas.GetSprite("xenolog/select_primortal")
)

type homeMenu struct {
	screen     *Screen
	selectLeft bool
	left       *selectBox
	right      *selectBox
}

func newHomeMenu(screen *Screen) *homeMenu {
	home := &homeMenu{
		screen:     screen,
		selectLeft: true,
		left:       newSelectBox(spriteAnimech, "Animech", ""),
		right:      newSelectBox(spritePrimortal, "Primortals", ""),
	}
	return home
}

func (s *homeMenu) Enter() {
	s.left.subLabel = ""
	if isAnimechUpgradeAvailable() {
		s.left.subLabel = "Upgrade Available"
	}

	s.right.subLabel = ""
	if isPrimortalUpgradeAvailable() {
		s.right.subLabel = "Upgrade Available"
	}
}

func isAnimechUpgradeAvailable() bool {
	available := game.CurrentSave().Animech.AnimechExperience
	level := game.CurrentSave().Animech.Upgrades.GetLevel()
	if available >= rpg.AnimechUpgradeExperienceRequiredToUpgrade(level+1) {
		return true
	}
	return false
}

func isPrimortalUpgradeAvailable() bool {
	for pType, save := range game.CurrentSave().Primortals {
		next := nextUnlock(pType.Primortal(), game.CurrentSave())
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
			s.screen.PushMenu(newAnimechMenu(s.screen))
		} else {
			s.screen.PushMenu(newPrimortalsMenu(s.screen))
		}
	}

	center := pixel.IM.Moved(gfx.IVec(screenWidth/2, screenHeight/2+3))
	dx := math.Floor(float64(selectBoxWidth) * 0.6)
	s.left.render(center.Moved(pixel.V(-dx, 0)), target, s.selectLeft, timeDelta)
	s.right.render(center.Moved(pixel.V(dx, 0)), target, !s.selectLeft, timeDelta)

	badgeStartClose.Render(target, pixel.IM.Moved(gfx.IVec(2, 2)), gfx.BottomLeft)
	badgeASelect.Render(target, pixel.IM.Moved(gfx.IVec(screenWidth-2, 2)), gfx.BottomRight)
}

var (
	selectBoxHeight = 82
	selectBoxWidth  = 74

	selectBoxLabelMargin = 3
	selectBoxLabelText   = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(colors.XenoLogText.RGBA),
			tbcfg.RenderFrom(gfx.TopCenter)))

	selectBoxSubLabelFlashSpeed = 1.0
	selectBoxSubLabelMargin     = 3
	selectBoxSubLabelText       = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.RenderFrom(gfx.TopCenter)))

	selectBoxArrow             = atlas.GetSprite("xenolog/select_arrow_down")
	selectBoxArrowMargin       = 2.0
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

	mask := colors.XenoLogText.RGBA
	if selected {
		mask = colors.XenoLogHighlight.RGBA
	}

	frameRect := pixel.R(0, 0, float64(selectBoxWidth), float64(selectBoxHeight))
	frame5px.Draw(target, frameRect, center, frames.WithColor(colors.XenoLogDark.RGBA))
	frame5pxBorder.Draw(target, frameRect, center, frames.WithColor(mask))

	bottomCenter := center.Moved(gfx.IVec(0, -selectBoxHeight/2))
	spriteCenter := bottomCenter.Moved(gfx.BottomCenter.Align(b.sprite))
	b.sprite.DrawColorMask(target, spriteCenter, mask)

	if b.label != "" {
		c := selectBoxLabelText.NewSimpleContent(b.label)
		selectBoxLabelText.Render(target, center.Moved(gfx.IVec(0, selectBoxHeight/2-selectBoxLabelMargin)), c, tbcfg.Foreground(mask))
	}

	if b.subLabel != "" {
		subLabelMask := colors.Lerp(colors.XenoLogText.RGBA, colors.XenoLogHighlight.RGBA, game.Utils().TimeCycleSin(selectBoxSubLabelFlashSpeed))
		c := selectBoxSubLabelText.NewSimpleContent(b.subLabel)
		selectBoxSubLabelText.Render(target,
			center.Moved(gfx.IVec(0, -selectBoxHeight/2-selectBoxSubLabelMargin)),
			c,
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
