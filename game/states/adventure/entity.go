package adventure

import (
	"fisherevans.com/project/f/game"
	"github.com/gopxl/pixel/v2"
)

type EntityReference struct {
	EntityId string
}

func (e EntityReference) Reference() EntityReference {
	return e
}

type Entity interface {
	Move(adv *AdventureState, timeDelta float64) float64
	Update(ctx *game.Context, adv *AdventureState, timeDelta float64)
	RenderMapLocation() pixel.Vec
	Location() MapLocation
	Sprite() *pixel.Sprite
	Reference() EntityReference
}

func (a *AdventureState) AddEntity(e Entity) bool {
	_, exists := a.entities[e.Reference()]
	if exists {
		return false
	}
	worked := a.attemptToOccupy(e.Location(), e.Reference())
	if !worked {
		return false
	}
	a.entities[e.Reference()] = e
	return true
}

// occupiedBy returns the entity and true if the location is occupied
func (a *AdventureState) occupiedBy(location MapLocation) (EntityReference, bool) {
	ent, occupied := a.occupiedLocations[location]
	return ent, occupied
}

// attemptToOccupy will do nothing and return false if the desired location is occupoied
// if it is not, it will occupy it and return true
func (a *AdventureState) attemptToOccupy(location MapLocation, entity EntityReference) bool {
	existingEnt, occupied := a.occupiedLocations[location]
	if occupied {
		if existingEnt == entity {
			return true
		}
		return false
	}
	a.occupiedLocations[location] = entity
	return true
}

// unoccupy will remove any occupancy in the location. it will return true if occupancy changed
func (a *AdventureState) unoccupy(location MapLocation, entity EntityReference) bool {
	existingEntity, wasOccupied := a.occupiedLocations[location]
	if wasOccupied {
		if existingEntity == entity {
			delete(a.occupiedLocations, location)
			return true
		}
		return false
	}
	return false
}
