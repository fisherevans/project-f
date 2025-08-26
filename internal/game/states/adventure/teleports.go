package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
)

type TeleportReference string

type Teleport struct {
	Destination   TeleportReference
	Location      MapLocation
	ExitDirection input.Direction
}
