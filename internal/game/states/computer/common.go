package computer

import (
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/badges"
)

var (
	badgeButtonStyle = badges.ButtonColorStyle{
		Action:    screenColors.Text,
		Button:    screenColors.Dark,
		Highlight: screenColors.Highlight,
	}

	badgeASelect        *badges.ButtonAction
	badgeASelectFlipped *badges.ButtonAction
	badgeALaunchFlipped *badges.ButtonAction
	badgeBCancel        *badges.ButtonAction
	badgeBBack          *badges.ButtonAction
	badgeBExit          *badges.ButtonAction
)

func init() {
	resources.RunOnceInitialized(func() {
		badgeASelect = badges.Using(atlas).ButtonAction("A", "select", badgeButtonStyle)
		badgeASelectFlipped = badges.Using(atlas).ButtonAction("A", "select", badgeButtonStyle).Flipped()
		badgeALaunchFlipped = badges.Using(atlas).ButtonAction("A", "launch", badgeButtonStyle).Flipped()
		badgeBExit = badges.Using(atlas).ButtonAction("B", "exit", badgeButtonStyle)
		badgeBBack = badges.Using(atlas).ButtonAction("B", "back", badgeButtonStyle)
		badgeBCancel = badges.Using(atlas).ButtonAction("B", "cancel", badgeButtonStyle)
	})
}
