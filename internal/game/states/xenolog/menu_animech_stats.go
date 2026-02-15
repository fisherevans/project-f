package xenolog

import (
	"fmt"

	"fisherevans.com/project/f/internal/game/states/xenolog/screen"
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/navigtion"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

var (
	animechStatePanePadding = 6
	animechStatePaneWidth   = 62
	animechStatePaneHeight  = 100
)

type animechStatsMenu struct {
	screen   *screen.Instance[*State]
	upgrades []*upgradedState
	actions  []*action
	nav      *navigtion.System[*animechStatsMenu]

	availableExperience            int
	uncommitedLevel                int
	requiredExperienceForNextLevel int
}

type upgradedState struct {
	navigtion.BaseItem[*animechStatsMenu]
	label                   string
	uncommitedLevelIncrease int
	level                   *int
	base                    int
	boostAtLevel            func(level int) int
}

func (u *upgradedState) OnButtonAJustPressed(m *animechStatsMenu) {
	if m.availableExperience >= m.requiredExperienceForNextLevel {
		u.uncommitedLevelIncrease++
	}
	m.updateUncommitedUpgrades()
	if m.availableExperience < m.requiredExperienceForNextLevel {
		m.nav.Highlight(m.actions[0])
	}
}

type action struct {
	navigtion.BaseItem[*animechStatsMenu]
	label   string
	handler func(m *animechStatsMenu)
}

func (a *action) OnButtonAJustPressed(m *animechStatsMenu) {
	if a.handler == nil {
		return
	}
	a.handler(m)
}

func newAnimechStatsMenu(s *screen.Instance[*State]) *animechStatsMenu {
	menu := &animechStatsMenu{
		screen: s,
		upgrades: []*upgradedState{
			{
				label:        "Sync",
				level:        &game.CurrentSave().Animech.Upgrades.SyncLevel,
				base:         rpg.BaseAnimechSync,
				boostAtLevel: rpg.AnimechUpgradeAdditionalSync,
			},
			{
				label:        "Shield",
				level:        &game.CurrentSave().Animech.Upgrades.ShieldLevel,
				base:         rpg.BaseAnimechShield,
				boostAtLevel: rpg.AnimechUpgradeAdditionalShield,
			},
		},
		actions: []*action{
			{
				label: "commit",
				handler: func(m *animechStatsMenu) {
					from := game.CurrentSave().Animech.Upgrades.GetLevel()
					for _, upgrade := range m.upgrades {
						for l := 0; l < upgrade.uncommitedLevelIncrease; l++ {
							game.CurrentSave().Animech.Experience -= rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
							*upgrade.level++
						}
						upgrade.uncommitedLevelIncrease = 0
					}
					to := game.CurrentSave().Animech.Upgrades.GetLevel()
					saveOrNotify()
					m.screen.SwapActiveMenuAnimated(newAnimechLevelUpAnimation(m.screen, from, to))
				},
			},
			{
				label: "reset",
				handler: func(m *animechStatsMenu) {
					for _, upgrade := range m.upgrades {
						upgrade.uncommitedLevelIncrease = 0
					}
					m.nav.Highlight(m.upgrades[0])
				},
			},
		},
	}
	menu.nav = navigtion.NewSystem[*animechStatsMenu](menu)
	for _, upgrade := range menu.upgrades {
		menu.nav.AddNextToLast(input.Down, upgrade)
	}
	for _, action := range menu.actions {
		menu.nav.AddNextToLast(input.Down, action)
	}
	return menu
}

func (m *animechStatsMenu) Enter() {}

func (m *animechStatsMenu) updateUncommitedUpgrades() {
	m.availableExperience = game.CurrentSave().Animech.Experience
	m.uncommitedLevel = game.CurrentSave().Animech.Upgrades.GetLevel()
	for _, upgrade := range m.upgrades {
		for l := 0; l < upgrade.uncommitedLevelIncrease; l++ {
			m.availableExperience -= rpg.AnimechUpgradeExperienceRequiredToUpgrade(m.uncommitedLevel)
			m.uncommitedLevel++
		}
	}
	m.requiredExperienceForNextLevel = rpg.AnimechUpgradeExperienceRequiredToUpgrade(m.uncommitedLevel)
}

func (m *animechStatsMenu) OnTick(target pixel.Target, timeDelta float64) {
	if game.Controls[*State]().ButtonB().JustPressed() {
		m.screen.PopMenuAnimated()
	} else {
		m.nav.HandleInputs(game.Controls[*State](), m)
	}
	m.updateUncommitedUpgrades()

	smallText := newTextRenderer(target, smallTextbox)
	regularTxt := newTextRenderer(target, regularTextbox)
	titleText := newTextRenderer(target, titleTextbox)

	// stat panels
	y := screenHeight - animechStatePanePadding
	paceCenterX := animechStatePaneWidth/2 + animechStatePanePadding
	m.renderStats(target, pixel.IM.Moved(gfx.IVec(paceCenterX, y)), true)
	m.renderStats(target, pixel.IM.Moved(gfx.IVec(screenWidth-paceCenterX, y)), false)

	// middle
	x := screenWidth / 2
	titleText.render("Level Up", x, y, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))
	y -= 16

	for _, upgrade := range m.upgrades {
		uncommitedUpgradeLevel := upgrade.base + upgrade.uncommitedLevelIncrease
		boostNow := upgrade.boostAtLevel(uncommitedUpgradeLevel)
		boostNext := upgrade.boostAtLevel(uncommitedUpgradeLevel + 1)
		boostText := fmt.Sprintf("+%d", boostNext-boostNow)

		if m.availableExperience < m.requiredExperienceForNextLevel {
			boostText = "---"
		}

		textMask, frameBg, frameBorder := screenColors.Text, screenColors.Dark, screenColors.Clear
		if m.nav.IsHighlighted(upgrade) {
			textMask = screenColors.Dark
			frameBorder = screenColors.Dark
			frameBg = screenColors.Text
			if m.availableExperience >= m.requiredExperienceForNextLevel {
				frameBg = flashingHighlight()
			}
		}

		frameR := pixel.R(0, 0, 54, 25)
		frame2px.Draw(target, frameR, gfx.Moved(x, y), frames.WithColor(frameBg), frames.WithRenderOrigin(gfx.TopCenter))
		frame2pxBorder.Draw(target, frameR, gfx.Moved(x, y), frames.WithColor(frameBorder), frames.WithRenderOrigin(gfx.TopCenter))

		y -= 4

		smallText.render("boost "+upgrade.label, x, y, textMask, tbcfg.RenderFrom(gfx.TopCenter))
		y -= 7

		regularTxt.render(boostText, x, y, textMask, tbcfg.RenderFrom(gfx.TopCenter))
		y -= 15
	}

	for _, a := range m.actions {
		mask, frameBg, frameBorder := screenColors.Text, screenColors.Dark, screenColors.Clear
		if m.nav.IsHighlighted(a) {
			mask = screenColors.Dark
			frameBg, frameBorder = flashingHighlight(), screenColors.Dark
		}
		frameR := pixel.R(0, 0, 40, 13)
		frame2px.Draw(target, frameR, gfx.Moved(x, y), frames.WithColor(frameBg), frames.WithRenderOrigin(gfx.TopCenter))
		frame2pxBorder.Draw(target, frameR, gfx.Moved(x, y), frames.WithColor(frameBorder), frames.WithRenderOrigin(gfx.TopCenter))

		y -= 4

		smallText.render(a.label, x, y, mask, tbcfg.RenderFrom(gfx.TopCenter))
		y -= 10
	}

	y -= 3
	smallText.render("experience", x, y, screenColors.Highlight, tbcfg.RenderFrom(gfx.TopCenter))
	y -= 8
	expText := fmt.Sprintf("%d{+c:xenolog_text} / %d", m.availableExperience, m.requiredExperienceForNextLevel)
	expColor := screenColors.Highlight
	if m.availableExperience < m.requiredExperienceForNextLevel {
		expColor = screenColors.Dark
	}
	w, _ := regularTxt.render(expText, x, y, expColor, tbcfg.RenderFrom(gfx.TopCenter))

	dx := w/2 + 6
	y -= 3
	smallText.render("available", x-dx, y, screenColors.Dark, tbcfg.RenderFrom(gfx.TopRight))
	smallText.render("required", x+dx, y, screenColors.Dark, tbcfg.RenderFrom(gfx.TopLeft))
}

func (m *animechStatsMenu) renderStats(target pixel.Target, topCenter pixel.Matrix, base bool) {
	a := game.CurrentSave().Animech
	level := a.Upgrades.GetLevel()
	titleMask, borderMask := screenColors.Text, screenColors.Text
	if !base {
		if level != m.uncommitedLevel {
			titleMask = flashingHighlight()
			borderMask = screenColors.Highlight
		}
		level = m.uncommitedLevel
	}
	m.drawFrame(target, topCenter, fmt.Sprintf("Level %d", level), titleMask, borderMask)

	smallTxt := newTextRenderer(target, smallTextbox).withMatrix(topCenter)
	regularTxt := newTextRenderer(target, regularTextbox).withMatrix(topCenter)

	y := -animechFrameTopPadding

	y -= 7
	smallTxt.render("Stats", 0, y, screenColors.Text, tbcfg.RenderFrom(gfx.TopCenter))

	renderStat := func(name string, value int, upgraded bool) {
		nameMask, arrowMask, statMask := screenColors.Text, screenColors.Clear, screenColors.Text
		if upgraded {
			nameMask, arrowMask, statMask = screenColors.Highlight, screenColors.Highlight, titleMask
		}
		arrowDx := 1
		smallTxt.render(name, arrowDx-7, y, nameMask, tbcfg.RenderFrom(gfx.RightCenter))
		arrowRight.DrawColorMask(target, topCenter.Moved(gfx.IVec(arrowDx, y+1)), arrowMask)
		regularTxt.render(fmt.Sprintf("%d", value), arrowDx+6, y, statMask, tbcfg.RenderFrom(gfx.LeftCenter))
	}

	y -= 16

	for _, upgrade := range m.upgrades {
		statLevel := *upgrade.level
		upgraded := false
		if !base {
			statLevel += upgrade.uncommitedLevelIncrease
			upgraded = upgrade.uncommitedLevelIncrease > 0
		}
		renderStat(upgrade.label, upgrade.base+upgrade.boostAtLevel(statLevel), upgraded)
		y -= 15
	}
}

func (m *animechStatsMenu) drawFrame(target pixel.Target, topCenter pixel.Matrix, title string, titleMask, frameBorderMask pixel.RGBA) {
	frame4px.Draw(target, pixel.R(0, 0,
		float64(animechStatePaneWidth), float64(animechStatePaneHeight)),
		topCenter,
		frames.WithColor(screenColors.Dark),
		frames.WithRenderOrigin(gfx.TopCenter))
	frame4pxBorder.Draw(target, pixel.R(0, 0,
		float64(animechStatePaneWidth), float64(animechStatePaneHeight)),
		topCenter,
		frames.WithColor(frameBorderMask),
		frames.WithRenderOrigin(gfx.TopCenter))
	titleTxt := newTextRenderer(target, titleTextbox).withMatrix(topCenter)
	titleTxt.render(title, 0, -4, titleMask, tbcfg.RenderFrom(gfx.TopCenter))

}
