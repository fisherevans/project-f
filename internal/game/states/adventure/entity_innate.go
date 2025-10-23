package adventure

import (
	"github.com/gopxl/pixel/v2"
)

type InnateEntity struct {
	BaseEntity
	MapLocation
}

func (i *InnateEntity) RenderScene(target pixel.Target, matrix pixel.Matrix) {
}

func (i *InnateEntity) Move(adv *State, timeDelta float64) float64 {
	return 0
}

func (i *InnateEntity) IsMoving() bool {
	return false
}

func (i *InnateEntity) Update(adv *State, timeDelta float64) {
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

func (i *InnateEntity) Interact(adv *State, source Entity) {

}

func (i *InnateEntity) RenderLight(target pixel.Target, matrix pixel.Matrix) {

}

func (i *InnateEntity) TeleportTo(s *State, location MapLocation) bool {
	s.GetTileState(i.Location()).RemoveEntity(i.GetEntityId())
	s.GetTileState(location).AddEntity(i.GetEntityId())
	i.MapLocation = location
	return true
}
