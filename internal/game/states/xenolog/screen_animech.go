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
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

type animechMenu struct {
	upgrades  []*upgradedState
	actions   []*action
	selection int
}

type upgradedState struct {
	label                   string
	uncommitedLevelIncrease int
	level                   *int
	boostAtLevel            func(level int) int
}

type action struct {
	label   string
	handler func(ctx *game.Context, m *animechMenu)
}

func newAnimechMenu(ctx *game.Context) *animechMenu {
	return &animechMenu{
		upgrades: []*upgradedState{
			{
				label:        "Sync",
				level:        &ctx.GameSave.Animech.Upgrades.SyncLevel,
				boostAtLevel: rpg.AnimechUpgradeAdditionalSync,
			},
			{
				label:        "Shield",
				level:        &ctx.GameSave.Animech.Upgrades.ShieldLevel,
				boostAtLevel: rpg.AnimechUpgradeAdditionalShield,
			},
		},
		actions: []*action{
			{
				label: "[COMMIT]",
				handler: func(ctx *game.Context, m *animechMenu) {
					log.Info().Msgf("in commit")
					for _, upgrade := range m.upgrades {
						for l := 0; l < upgrade.uncommitedLevelIncrease; l++ {
							ctx.GameSave.Animech.AnimechExperience -= rpg.AnimechUpgradeExperienceRequiredToUpgrade(ctx.GameSave.Animech.Upgrades.GetLevel())
							*upgrade.level++
						}
						upgrade.uncommitedLevelIncrease = 0
					}
					m.selection = 0
					if err := ctx.GameSave.Save(); err != nil {
						log.Error().Err(err).Msg("failed to save game")
						ctx.Notify("Failed to save game")
					}
				},
			},
			{
				label: "[RESET]",
				handler: func(ctx *game.Context, m *animechMenu) {
					for _, upgrade := range m.upgrades {
						upgrade.uncommitedLevelIncrease = 0
					}
					m.selection = 0
				},
			},
			{
				label: "[RE-SPEC]",
				handler: func(ctx *game.Context, m *animechMenu) {
					for _, upgrade := range m.upgrades {
						upgrade.uncommitedLevelIncrease = 0
					}
					for ctx.GameSave.Animech.Upgrades.SyncLevel > 0 {
						ctx.GameSave.Animech.Upgrades.SyncLevel--
						ctx.GameSave.Animech.AnimechExperience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(ctx.GameSave.Animech.Upgrades.GetLevel())
					}
					for ctx.GameSave.Animech.Upgrades.ShieldLevel > 0 {
						ctx.GameSave.Animech.Upgrades.ShieldLevel--
						ctx.GameSave.Animech.AnimechExperience += rpg.AnimechUpgradeExperienceRequiredToUpgrade(ctx.GameSave.Animech.Upgrades.GetLevel())
					}
					m.selection = 0
				},
			},
		},
	}
}

func (m *animechMenu) isUpgradeAvailable(ctx *game.Context) bool {
	available := ctx.GameSave.Animech.AnimechExperience
	level := ctx.GameSave.Animech.Upgrades.GetLevel()
	if available >= rpg.AnimechUpgradeExperienceRequiredToUpgrade(level+1) {
		return true
	}
	return false
}

func (m *animechMenu) RenderAnimech(s *Screen, ctx *game.Context, target *pixel.Batch, timeDelta float64) {
	if ctx.Controls.ButtonB().JustPressed() {
		s.returnToMenu(ctx)
	}

	availableExperience := ctx.GameSave.Animech.AnimechExperience
	currentLevel := ctx.GameSave.Animech.Upgrades.GetLevel()
	uncommitedLevel := currentLevel
	for _, upgrade := range m.upgrades {
		for l := 0; l < upgrade.uncommitedLevelIncrease; l++ {
			availableExperience -= rpg.AnimechUpgradeExperienceRequiredToUpgrade(uncommitedLevel)
			uncommitedLevel++
		}
	}

	nextLevelRequiredExperience := rpg.AnimechUpgradeExperienceRequiredToUpgrade(uncommitedLevel)

	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Up) {
		m.selection--
		if m.selection < 0 {
			m.selection = 0
		}
	}
	if ctx.Controls.DPad().DirectionJustPressedOrRepeated(input.Down) {
		m.selection++
		maxSelection := len(m.upgrades) + len(m.actions) - 1
		if m.selection > maxSelection {
			m.selection = maxSelection
		}
	}
	if m.selection >= 0 && m.selection < len(m.upgrades) {
		if ctx.Controls.DPad().JustPressedDirection() == input.Left {
			m.upgrades[m.selection].uncommitedLevelIncrease--
			if m.upgrades[m.selection].uncommitedLevelIncrease < 0 {
				m.upgrades[m.selection].uncommitedLevelIncrease = 0
			}
		}
		if ctx.Controls.DPad().JustPressedDirection() == input.Right {
			if availableExperience >= nextLevelRequiredExperience {
				m.upgrades[m.selection].uncommitedLevelIncrease++
			}
		}
	}
	if ctx.Controls.ButtonA().JustPressed() {
		actionSelection := m.selection - len(m.upgrades)
		if actionSelection >= 0 && actionSelection < len(m.actions) {
			m.actions[actionSelection].handler(ctx, m)
		}
	}

	smallText := newTextRenderer(ctx, target, smallTextbox)
	mediumText := newTextRenderer(ctx, target, mediumTextbox)

	lineHeight := 14

	statX := 20
	boostX := 45

	y := screenHeight - lineHeight

	{
		smallText.render("LEVEL", statX, y, uiMask)
		smallText.render("BOOST", boostX, y, uiMask, tbcfg.RenderFrom(gfx.TopLeft))
		y -= lineHeight
	}

	for id, upgrade := range m.upgrades {
		mask := uiMask
		if id == m.selection {
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
		if availableExperience >= nextLevelRequiredExperience {
			smallText.render(">", statX+10, y, mask)
		}
		smallText.render(fmt.Sprintf("+%d %s", upgrade.boostAtLevel(level), upgrade.label), boostX, y, mask, tbcfg.RenderFrom(gfx.TopLeft))
		y -= lineHeight
	}

	for id, action := range m.actions {
		mask := uiMask
		label := action.label
		if id == m.selection-len(m.upgrades) {
			mask = colors.White.RGBA
			label = ">> " + label + " <<"
		}
		smallText.render(label, 50, y, mask)
		y -= lineHeight
	}

	expText := fmt.Sprintf("Experience required: %d/%d", availableExperience, nextLevelRequiredExperience)
	smallText.render(expText, 10, 10, uiMask, tbcfg.RenderFrom(gfx.BottomLeft))

	levelText := fmt.Sprintf("%d", currentLevel)
	if currentLevel != uncommitedLevel {
		levelText += fmt.Sprintf("{+c:white} > %d", uncommitedLevel)
	}
	mediumText.render("Level: "+levelText, screenWidth/2+20, screenHeight-10, uiMask, tbcfg.RenderFrom(gfx.TopLeft))
}

type textRenderer struct {
	ctx    *game.Context
	target *pixel.Batch
	tb     *textbox.Instance
}

func newTextRenderer(ctx *game.Context, target *pixel.Batch, tb *textbox.Instance) *textRenderer {
	return &textRenderer{
		ctx:    ctx,
		target: target,
		tb:     tb,
	}
}

func (r *textRenderer) render(text string, x, y int, mask pixel.RGBA, opts ...tbcfg.ConfigOpt) (int, int) {
	c := r.tb.NewComplexContent(text)
	matrix := pixel.IM.Moved(gfx.IVec(x, y))
	r.tb.Render(r.ctx, r.target, matrix, c, append(opts, tbcfg.Foreground(mask))...)
	return c.Width(), c.Height()
}
