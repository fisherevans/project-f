package adventure

import (
	"github.com/gopxl/pixel/v2"

	"fisherevans.com/project/f/internal/game"
)

type InnateEntity struct {
	BaseEntity
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

func (i *InnateEntity) PreciseMapLocation() pixel.Vec {
	return pixel.V(float64(i.X), float64(i.Y))
}

func (i *InnateEntity) RenderMapLocation() pixel.Vec {
	return i.PreciseMapLocation()
}

func (i *InnateEntity) Interact(ctx *game.Context, adv *State, source Entity) {

}

func (i *InnateEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {

}
