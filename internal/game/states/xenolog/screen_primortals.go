package xenolog

import (
	"fmt"
	"math"

	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	primortalRowHeight = 14
)

type primortalsMenu struct {
	selection    int
	dy           int
	scrollMargin int
	scrollSpeed  float64
	initialized  bool
	elapsed      float64

	detailView *primortalDetailView
}

func newPrimortalsMenu(ctx *game.Context) *primortalsMenu {
	return &primortalsMenu{
		selection:    1,
		dy:           0,
		scrollMargin: 20,
		scrollSpeed:  10.0,
	}
}

func (m *primortalsMenu) isUpgradeAvailable(ctx *game.Context) bool {
	for pType, save := range ctx.GameSave.Primortals {
		next := nextUnlock(pType.Primortal(), ctx.GameSave)
		if next != nil && save.ResearchPoints >= next.Cost {
			return true
		}
	}
	return false
}

func (m *primortalsMenu) RenderPrimortals(s *Screen, ctx *game.Context, target *pixel.Batch, timeDelta float64) {
	m.elapsed += timeDelta

	if !m.initialized {
		m.ensureVisibleNow()
		m.initialized = true
	}

	if m.detailView != nil {
		m.detailView.RenderPrimortalView(m, ctx, target, timeDelta)
		return
	}

	m.handleInput(ctx, s)
	m.updateScroll(screenHeight, timeDelta)
	m.drawList(ctx, target)
}

func (m *primortalsMenu) handleInput(ctx *game.Context, s *Screen) {
	if ctx.Controls.ButtonB().JustPressed() {
		s.returnToMenu(ctx)
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Up) {
		m.selection--
		if m.selection < 1 {
			m.selection = 1
		}
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Down) {
		m.selection++
		if m.selection > rpg.MaxXenoLogEntryIndex {
			m.selection = rpg.MaxXenoLogEntryIndex
		}
	}
	if ctx.Controls.ButtonA().JustPressed() {
		pType, exists := rpg.XenoLogEntries[m.selection]
		if exists {
			m.detailView = newPrimortalDetailView(pType)
		}
	}
}

func (m *primortalsMenu) calcTargetDy() float64 {
	// Keep selected row visible without unnecessary movement
	selY := screenHeight - primortalRowHeight - (m.selection-1)*primortalRowHeight
	selYWithOffset := selY + m.dy

	minY := m.scrollMargin
	maxY := screenHeight - m.scrollMargin

	targetDy := float64(m.dy)
	if selYWithOffset < minY {
		targetDy += float64(minY - selYWithOffset)
	} else if selYWithOffset > maxY {
		targetDy -= float64(selYWithOffset - maxY)
	}
	return targetDy
}

func (m *primortalsMenu) ensureVisibleNow() {
	// Set dy instantly so the initial render doesn't jiggle
	m.dy = int(math.Round(m.calcTargetDy()))
}

func (m *primortalsMenu) updateScroll(screenHeight int, timeDelta float64) {
	targetDy := m.calcTargetDy()
	// Adaptive ease toward target: faster when far, gentle when near
	dist := targetDy - float64(m.dy)
	absDist := math.Abs(dist)
	mult := 1.0 + math.Min(3.0, absDist/float64(primortalRowHeight*2))
	step := m.scrollSpeed * mult * timeDelta
	if step > 0.95 { // avoid overshoot/snapping on low FPS spikes
		step = 0.95
	}
	m.dy = int(math.Floor(float64(m.dy) + dist*step))
}

func (m *primortalsMenu) drawList(ctx *game.Context, target *pixel.Batch) {
	smallText := newTextRenderer(ctx, target, smallTextbox)

	baseY := screenHeight - primortalRowHeight + m.dy
	upgradeMask := flash(m.elapsed, selectBoxSubLabelFlashSpeed, uiMask, colors.White.RGBA)

	upgradeAbove := false
	upgradeBelow := false

	for index := 1; index <= rpg.MaxXenoLogEntryIndex; index++ {
		y := baseY - (index-1)*primortalRowHeight
		visibleTop := screenHeight + primortalRowHeight/2
		visibleBottom := -primortalRowHeight / 2

		// Determine if this row has an upgrade available
		rowHasUpgrade := false
		if pId, ok := rpg.XenoLogEntries[index]; ok {
			if save, exists := ctx.GameSave.Primortals[pId]; exists {
				if next := nextUnlock(pId.Primortal(), ctx.GameSave); next != nil && save.ResearchPoints >= next.Cost {
					rowHasUpgrade = true
				}
			}
		}

		// Off-screen handling: mark flags and skip draw
		if y > visibleTop {
			if rowHasUpgrade {
				upgradeAbove = true
			}
			continue
		}
		if y < visibleBottom {
			if rowHasUpgrade {
				upgradeBelow = true
			}
			continue
		}

		mask := uiMask
		if index == m.selection {
			mask = uiMaskSelected
			smallText.render(">", 10, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		}
		smallText.render(fmt.Sprintf("%03d", index), 20, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		pId, exists := rpg.XenoLogEntries[index]
		if !exists {
			smallText.render("???", 35, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
			continue
		}
		p := pId.Primortal()
		nameDx, _ := smallText.render(p.Name, 35, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		x := 35 + nameDx + 10
		nextSkill := nextUnlock(p, ctx.GameSave)
		if nextSkill == nil {
			smallText.render("[COMPLETE]", x, y, mask, tbcfg.RenderFrom(gfx.LeftCenter))
		} else if rowHasUpgrade {
			smallText.render("> UPGRADE AVAILABLE <", x, y, upgradeMask, tbcfg.RenderFrom(gfx.LeftCenter))
		}
	}

	toolTipMargin := 2
	if upgradeAbove {
		smallText.render("^ UPGRADE", screenWidth-toolTipMargin, screenHeight-toolTipMargin, upgradeMask, tbcfg.RenderFrom(gfx.TopRight))
	}
	if upgradeBelow {
		smallText.render("v UPGRADE", screenWidth-toolTipMargin, toolTipMargin, upgradeMask, tbcfg.RenderFrom(gfx.BottomRight))
	}
}

func (m *primortalsMenu) returnToList(ctx *game.Context) {
	m.detailView = nil
}

func nextUnlock(primortal rpg.Primortal, save *rpg.GameSave) *rpg.UnlockableSkill {
	for _, us := range primortal.UnlockableSkills {
		if !save.IsSkillUnlocked(us.SkillId) {
			return &us
		}
	}
	return nil
}
