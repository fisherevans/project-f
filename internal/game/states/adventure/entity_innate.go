package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
)

type InnateEntity struct {
	EntityId
	MapLocation
}

func (i *InnateEntity) Move(adv *State, timeDelta float64) float64 {
	return 0
}

func (i *InnateEntity) Update(ctx *game.Context, adv *State, timeDelta float64) {
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

func (i *InnateEntity) Interact(ctx *game.Context, adv *State, source Entity) {

}
