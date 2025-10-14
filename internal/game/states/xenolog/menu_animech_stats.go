package xenolog

import (
	"fmt"

	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/navigtion"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type animechStatsMenu struct {
	screen   *Screen
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
	boostAtLevel            func(level int) int
}

func (u *upgradedState) OnDirectionJustPressed(dir input.Direction, m *animechStatsMenu) {
	if dir == input.Left {
		u.uncommitedLevelIncrease--
		if u.uncommitedLevelIncrease < 0 {
			u.uncommitedLevelIncrease = 0
		}
		m.updateUncommitedUpgrades()
	}
	if dir == input.Right {
		if m.availableExperience >= m.requiredExperienceForNextLevel {
			u.uncommitedLevelIncrease++
		}
		m.updateUncommitedUpgrades()
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

func newAnimechStatsMenu(s *Screen) *animechStatsMenu {
	menu := &animechStatsMenu{
		screen: s,
		nav:    navigtion.NewSystem[*animechStatsMenu](),
		upgrades: []*upgradedState{
			{
				label:        "Sync",
				level:        &game.CurrentSave().Animech.Upgrades.SyncLevel,
				boostAtLevel: rpg.AnimechUpgradeAdditionalSync,
			},
			{
				label:        "Shield",
				level:        &game.CurrentSave().Animech.Upgrades.ShieldLevel,
				boostAtLevel: rpg.AnimechUpgradeAdditionalShield,
			},
		},
		actions: []*action{
			{
				label: "[COMMIT]",
				handler: func(m *animechStatsMenu) {
					for _, upgrade := range m.upgrades {
						for l := 0; l < upgrade.uncommitedLevelIncrease; l++ {
							game.CurrentSave().Animech.AnimechExperience -= rpg.AnimechUpgradeExperienceRequiredToUpgrade(game.CurrentSave().Animech.Upgrades.GetLevel())
							*upgrade.level++
						}
						upgrade.uncommitedLevelIncrease = 0
					}
					m.nav.Highlight(m.upgrades[0], m)
					if err := game.CurrentSave().Save(); err != nil {
						log.Error().Err(err).Msg("failed to save game")
						game.DebugNotification("Failed to save game")
					}
				},
			},
			{
				label: "[RESET]",
				handler: func(m *animechStatsMenu) {
					for _, upgrade := range m.upgrades {
						upgrade.uncommitedLevelIncrease = 0
					}
					m.nav.Highlight(m.upgrades[0], m)
				},
			},
		},
	}
	for _, upgrade := range menu.upgrades {
		menu.nav.AddNextToLast(input.Down, menu, upgrade)
	}
	for _, action := range menu.actions {
		menu.nav.AddNextToLast(input.Down, menu, action)
	}
	return menu
}

func (m *animechStatsMenu) Enter() {}

func (m *animechStatsMenu) updateUncommitedUpgrades() {
	m.availableExperience = game.CurrentSave().Animech.AnimechExperience
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
		m.screen.PopMenu()
	} else {
		m.nav.HandleInputs(game.Controls[*State](), m)
	}
	m.updateUncommitedUpgrades()

	smallText := newTextRenderer(target, smallTextbox)
	titleText := newTextRenderer(target, titleTextbox)

	lineHeight := 14

	statX := 20
	boostX := 45

	y := screenHeight - lineHeight

	{
		smallText.render("LEVEL", statX, y, colors.XenoLogText.RGBA)
		smallText.render("BOOST", boostX, y, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopLeft))
		y -= lineHeight
	}

	for _, upgrade := range m.upgrades {
		mask := colors.XenoLogText.RGBA
		if m.nav.IsHighlighted(upgrade) {
			mask = colors.White.RGBA
		}
		level := *upgrade.level + upgrade.uncommitedLevelIncrease
		levelFormat := ""
		if upgrade.uncommitedLevelIncrease > 0 {
			levelFormat = "{+u:white}"
		}
		if upgrade.uncommitedLevelIncrease > 0 {
			smallText.render("<", statX-10, y, mask)
		}
		smallText.render(fmt.Sprintf("%s%d", levelFormat, level), statX, y, mask)
		if m.availableExperience >= m.requiredExperienceForNextLevel {
			smallText.render(">", statX+10, y, mask)
		}
		smallText.render(fmt.Sprintf("+%d %s", upgrade.boostAtLevel(level), upgrade.label), boostX, y, mask, tbcfg.RenderFrom(gfx.TopLeft))
		y -= lineHeight
	}

	for _, a := range m.actions {
		mask := colors.XenoLogText.RGBA
		label := a.label
		if m.nav.IsHighlighted(a) {
			mask = colors.White.RGBA
			label = ">> " + label + " <<"
		}
		smallText.render(label, 50, y, mask)
		y -= lineHeight
	}

	expText := fmt.Sprintf("Experience required: %d/%d", m.availableExperience, m.requiredExperienceForNextLevel)
	smallText.render(expText, 10, 10, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.BottomLeft))

	currentLevel := game.CurrentSave().Animech.Upgrades.GetLevel()
	levelText := fmt.Sprintf("%d", currentLevel)
	if currentLevel != m.uncommitedLevel {
		levelText += fmt.Sprintf("{+c:white} > %d", m.uncommitedLevel)
	}
	titleText.render("Level: "+levelText, screenWidth/2+20, screenHeight-10, colors.XenoLogText.RGBA, tbcfg.RenderFrom(gfx.TopLeft))
}
