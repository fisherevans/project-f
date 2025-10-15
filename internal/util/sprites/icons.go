package sprites

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/pixelutil"
)

func StanceIcons(atlas *resources.Atlas) map[rpg.CombatStance]pixelutil.BoundedDrawable {
	return map[rpg.CombatStance]pixelutil.BoundedDrawable{
		rpg.TickStanceDefending:  atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 3), // shield
		rpg.TickStanceReflecting: atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 1, 1), // reflect arrow
		rpg.TickStanceVulnerable: atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 7, 1), // cross out shield
		rpg.TickStanceExposed:    atlas.GetTilesheetSprite("combat/tick_bar/stance_icons", 7, 3), // !!!
	}
}

func StatusIcons(atlas *resources.Atlas) map[rpg.StatusType]pixelutil.BoundedDrawable {
	return map[rpg.StatusType]pixelutil.BoundedDrawable{
		rpg.StatusFortified: atlas.GetTilesheetSprite("combat/status_icons", 5, 1),
		rpg.StatusPoisoned:  atlas.GetTilesheetSprite("combat/status_icons", 1, 1),
		rpg.StatusBurning:   atlas.GetTilesheetSprite("combat/status_icons", 2, 1),
		rpg.StatusIonized:   atlas.GetTilesheetSprite("combat/status_icons", 3, 1),
		rpg.StatusMending:   atlas.GetTilesheetSprite("combat/status_icons", 4, 1),
	}
}
