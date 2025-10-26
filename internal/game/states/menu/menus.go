package menu

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
)

func createMainMenu(s *State) *Menu {
	return &Menu{
		state: s,
		items: []MenuItem{
			newSimpleSelectionOption(s, "Return to game", func(s *State) {
				game.SetActiveStateIntent(game.SwapStateIntent{
					State: s.background,
				})
			}),
			newSimpleSelectionOption(s, "Settings", func(s *State) {
				s.PushMenu(createSettingsMenu(s))
			}),
			newSimpleSelectionOption(s, "Quit", func(s *State) {
				game.SetActiveStateIntent(game.TitleIntent{})
			}),
		},
	}
}

func createSettingsMenu(s *State) *Menu {
	return &Menu{
		state: s,
		items: []MenuItem{
			newItemEnumToggle[rpg.LightingMode](s, "Lighting", []itemEnumToggleOption[rpg.LightingMode]{
				newEnumToggleOption[rpg.LightingMode]("Blended", rpg.LightingModeBlended),
				newEnumToggleOption[rpg.LightingMode]("Over", rpg.LightingModeOver),
				newEnumToggleOption[rpg.LightingMode]("Off", rpg.LightingModeOff),
			}, &game.CurrentSave().SystemSettings.Lighting.LightingMode),
			newItemEnumToggle[rpg.LightingComposition](s, "Lighting Comp", []itemEnumToggleOption[rpg.LightingComposition]{
				newEnumToggleOption[rpg.LightingComposition]("Full", rpg.LightingCompositionFull),
				newEnumToggleOption[rpg.LightingComposition]("Ambient", rpg.LightingCompositionAmbient),
			}, &game.CurrentSave().SystemSettings.Lighting.LightingComposition),
			newItemEnumToggle[rpg.BloomMode](s, "Bloom", []itemEnumToggleOption[rpg.BloomMode]{
				newEnumToggleOption[rpg.BloomMode]("Blended", rpg.BloomModeBlended),
				newEnumToggleOption[rpg.BloomMode]("Over Blur", rpg.BloomModeOverBlurred),
				newEnumToggleOption[rpg.BloomMode]("Over Highlight", rpg.BloomModeOverThreshold),
				newEnumToggleOption[rpg.BloomMode]("Off", rpg.BloomModeOff),
			}, &game.CurrentSave().SystemSettings.Lighting.BloomMode),
			newItemBoolToggle(s, "Pixel Grid", "Off", "On", &game.CurrentSave().SystemSettings.RetroFrame.DisablePixelGrid).withListener(func(bool) {
				game.Flags().Set("retro_frame_reset")
			}),
			newItemFloatSlider(s, "Scanline Darken", "%.3f", 0, 0.25, 0.005, &game.CurrentSave().SystemSettings.RetroFrame.ScanlineDarken).withListener(func(float64) {
				game.Flags().Set("retro_frame_reset")
			}),
			newItemFloatSlider(s, "Grid Darken X", "%.3f", 0, 0.25, 0.005, &game.CurrentSave().SystemSettings.RetroFrame.GridDarkenX).withListener(func(float64) {
				game.Flags().Set("retro_frame_reset")
			}),
			newItemFloatSlider(s, "Grid Darken Y", "%.3f", 0, 0.25, 0.005, &game.CurrentSave().SystemSettings.RetroFrame.GridDarkenY).withListener(func(float64) {
				game.Flags().Set("retro_frame_reset")
			}),
			newItemFloatSlider(s, "Subpixel Tint", "%.3f", 0, 0.25, 0.005, &game.CurrentSave().SystemSettings.RetroFrame.SubpixelTint).withListener(func(float64) {
				game.Flags().Set("retro_frame_reset")
			}),
		},
	}
}
