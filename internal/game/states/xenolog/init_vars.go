package xenolog

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/combat/tick_bar"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"fisherevans.com/project/f/internal/util/pixelutil"
	"fisherevans.com/project/f/internal/util/sprites"
	"fisherevans.com/project/f/internal/util/textbox"
	"fisherevans.com/project/f/internal/util/textbox/tbcfg"
)

// Textboxes - shared text rendering instances
var (
	smallTextbox   *textbox.Instance // FF font, narrow width for detail text
	regularTextbox *textbox.Instance // M5x7 font, full width for body text
	titleTextbox   *textbox.Instance // AddStandard font, full width for titles
)

// Frames - shared UI frame instances (consolidated duplicates)
var (
	frame1px       *frames.Instance
	frame1pxBorder *frames.Instance
	frame2px       *frames.Instance
	frame2pxBorder *frames.Instance
	frame4px       *frames.Instance // Used for 3px, 4px, and 5px variants
	frame4pxBorder *frames.Instance
)

// Sprites - menu and UI sprites
var (
	spriteAnimech                        pixelutil.BoundedDrawable
	spritePrimortal                      pixelutil.BoundedDrawable
	spriteUnknown                        pixelutil.BoundedDrawable
	selectBoxArrow                       pixelutil.BoundedDrawable
	scrollCursor                         pixelutil.BoundedDrawable
	noneTickSprite                       pixelutil.BoundedDrawable
	particleSprite                       pixelutil.BoundedDrawable
	arrowUp                              pixelutil.BoundedDrawable
	arrowRight                           pixelutil.BoundedDrawable
	arrowDown                            pixelutil.BoundedDrawable
	arrowLeft                            pixelutil.BoundedDrawable
	arrowLeft6px                         pixelutil.BoundedDrawable
	dot                                  pixelutil.BoundedDrawable
	skillNodeSpriteHidden                pixelutil.BoundedDrawable
	skillNodeSpriteUnlockedHighlighted   pixelutil.BoundedDrawable
	skillNodeSpriteUnlockableHighlighted pixelutil.BoundedDrawable
	skillNodeSpriteHighlightCursor       pixelutil.BoundedDrawable
	skillNodeSpriteUnlocked              pixelutil.BoundedDrawable
	skillNodeSpriteUnlockable            pixelutil.BoundedDrawable
)

// Badges - button action badges
var (
	badgeASelect             *badges.ButtonAction
	badgeBBack               *badges.ButtonAction
	badgeStartClose          *badges.ButtonAction
	badgeF1Reset             *badges.ButtonAction
	badgeADetails            *badges.ButtonAction
	badgeAUnlock             *badges.ButtonAction
	badgeASwap               *badges.ButtonAction
	badgeSelectDetails       *badges.ButtonAction
	badgeBCancelOnDark       *badges.ButtonAction
	badgeSelectDetailsOnDark *badges.ButtonAction
	badgeBack                badges.Instance
	badgeNextUnlock          badges.Instance
	badgeBackToTop           badges.Instance
)

// Specialized textboxes - context-specific configurations
var (
	selectBoxLabelText               *textbox.Instance
	selectBoxSubLabelText            *textbox.Instance
	skillSetDetailSmallTextbox       *textbox.Instance
	smallTextboxSkillDetail          *textbox.Instance
	primortalDescriptionTextbox      *textbox.Instance
	primortalSkillDescriptionTextbox *textbox.Instance
)

// Combat/skill rendering
var (
	tickBarRenderer *tick_bar.Renderer
	stanceIcons     map[rpg.CombatStance]pixelutil.BoundedDrawable
	statusIcons     map[rpg.StatusType]pixelutil.BoundedDrawable
)

func initializeXenologVariables() {
	// Shared textboxes
	smallTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth-30, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	regularTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	titleTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(screenWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))

	// Shared frames (consolidated 3px/4px/5px into single 4px instance)
	frame1px = frames.New("common/rounded_frame_1px", atlas, frames.WithRenderOrigin(gfx.Centered))
	frame1pxBorder = frames.New("common/rounded_border_frame_1px", atlas, frames.WithRenderOrigin(gfx.Centered))
	frame2px = frames.New("common/rounded_frame_2px", atlas, frames.WithRenderOrigin(gfx.Centered))
	frame2pxBorder = frames.New("common/rounded_border_frame_2px", atlas, frames.WithRenderOrigin(gfx.Centered))
	frame4px = frames.New("common/rounded_frame_4px", atlas, frames.WithRenderOrigin(gfx.Centered))
	frame4pxBorder = frames.New("common/rounded_border_frame_4px", atlas, frames.WithRenderOrigin(gfx.Centered))

	// Sprites
	spriteAnimech = atlas.GetSprite("xenolog/select_animech")
	spritePrimortal = atlas.GetSprite("xenolog/select_primortal")
	spriteUnknown = atlas.GetSprite("xenolog/select_unknown")
	selectBoxArrow = atlas.GetSprite("xenolog/select_arrow_down")
	scrollCursor = atlas.GetSprite("xenolog/scroll_cursor")
	noneTickSprite = atlas.GetSprite("2x2")
	particleSprite = atlas.GetSprite("1x1")
	arrowUp = atlas.GetTilesheetSprite("common/arrows_5px", 1, 1)
	arrowRight = atlas.GetTilesheetSprite("common/arrows_5px", 2, 1)
	arrowDown = atlas.GetTilesheetSprite("common/arrows_5px", 3, 1)
	arrowLeft = atlas.GetTilesheetSprite("common/arrows_5px", 4, 1)
	arrowLeft6px = atlas.GetTilesheetSprite("common/arrows_6px", 4, 1)
	dot = atlas.GetTilesheetSprite("common/symbols_5px", 1, 1)
	skillNodeSpriteHidden = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 4, 1)
	skillNodeSpriteUnlockedHighlighted = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 3, 1)
	skillNodeSpriteUnlockableHighlighted = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 2, 1)
	skillNodeSpriteHighlightCursor = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 5, 1)
	skillNodeSpriteUnlocked = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 6, 1)
	skillNodeSpriteUnlockable = atlas.GetTilesheetSprite("xenolog/skill_tree/nodes", 7, 1)

	// Badges
	badgeASelect = badges.Using(atlas).ButtonAction("A", "select", badgeButtonStyle)
	badgeBBack = badges.Using(atlas).ButtonAction("B", "back", badgeButtonStyle)
	badgeStartClose = badges.Using(atlas).ButtonAction("start", "close", badgeButtonStyle)
	badgeF1Reset = badges.Using(atlas).ButtonAction("F1", "reset", badgeButtonStyle)
	badgeADetails = badges.Using(atlas).ButtonAction("A", "details", badgeButtonStyle)
	badgeAUnlock = badges.Using(atlas).ButtonAction("A", "unlock", badgeButtonStyle)
	badgeASwap = badges.Using(atlas).ButtonAction("A", "swap skill", skillSetBadgeStyle).Flipped()
	badgeSelectDetails = badges.Using(atlas).ButtonAction("select", "view skill details", skillSetBadgeStyle)
	badgeBCancelOnDark = badges.Using(atlas).ButtonAction("B", "cancel", onDarkStyle)
	badgeSelectDetailsOnDark = badges.Using(atlas).ButtonAction("select", "details", onDarkStyle).Flipped()
	badgeBack = badges.Using(atlas).Of("back", screenColors.Dark, screenColors.Text, 30)
	badgeNextUnlock = badges.Using(atlas).Of("scroll to next unlock", screenColors.Dark, screenColors.Text, 80)
	badgeBackToTop = badges.Using(atlas).Of("back to top", screenColors.Dark, screenColors.Text, 50)

	// Specialized textboxes
	selectBoxLabelText = textbox.NewInstance(atlas.GetFont(resources.FontNameAddStandard),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	selectBoxSubLabelText = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(selectBoxWidth, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	skillSetDetailSmallTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(skillSetCurrentSkillFrameWidth-10, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	smallTextboxSkillDetail = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(140-4, 9,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignLeft),
			tbcfg.VAligned(tbcfg.VAlignMiddle),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.LeftCenter)))
	primortalDescriptionTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(screenWidth-primortalIconSize-detailMargin*2-primortalIconMargin*3, 8,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))
	primortalSkillDescriptionTextbox = textbox.NewInstance(atlas.GetFont(resources.FontNameFF),
		tbcfg.NewConfig(primortalSkillDetailWidth-primortalSkillDetailMargin*2, 10,
			tbcfg.WithExpandMode(tbcfg.ExpandFit),
			tbcfg.HAligned(tbcfg.HAlignCenter),
			tbcfg.VAligned(tbcfg.VAlignTop),
			tbcfg.Foreground(screenColors.Text),
			tbcfg.RenderFrom(gfx.TopCenter)))

	// Skill set arrows
	skillSetArrows[input.NotPressed] = atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 1, 1)
	skillSetArrows[input.Up] = atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 2, 1)
	skillSetArrows[input.Right] = atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 3, 1)
	skillSetArrows[input.Down] = atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 4, 1)
	skillSetArrows[input.Left] = atlas.GetTilesheetSprite("xenolog/skill_set_arrows", 5, 1)

	// Combat/skill rendering
	tickBarRenderer = tick_bar.NewRenderer(atlas)
	stanceIcons = sprites.StanceIcons(atlas)
	statusIcons = sprites.StatusIcons(atlas)
}

func arrowSprite(dir input.Direction) pixelutil.BoundedDrawable {
	switch dir {
	case input.Up:
		return arrowUp
	case input.Right:
		return arrowRight
	case input.Down:
		return arrowDown
	case input.Left:
		return arrowLeft
	default:
		panic("invalid direction")
	}
}
