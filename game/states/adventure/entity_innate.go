package adventure

import (
	"fisherevans.com/project/f/game"
	"github.com/gopxl/pixel/v2"
)

type InnateEntity struct {
	EntityReference
	MapLocation
}

func (i *InnateEntity) Move(adv *AdventureState, timeDelta float64) float64 {
	return 0
}

func (i *InnateEntity) Update(ctx *game.Context, adv *AdventureState, timeDelta float64) {
}

func (i *InnateEntity) Location() MapLocation {
	return i.MapLocation
}

func (i *InnateEntity) RenderMapLocation() pixel.Vec {
	return pixel.V(float64(i.X), float64(i.Y))
}

func (i *InnateEntity) Sprite() *pixel.Sprite {
	return nil
}
