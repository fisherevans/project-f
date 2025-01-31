package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
)

type EntityReference struct {
	EntityId string
}

func (e EntityReference) Reference() EntityReference {
	return e
}

type Entity interface {
	Move(adv *State, timeDelta float64) float64
	Update(ctx *game.Context, adv *State, timeDelta float64)
	RenderMapLocation() pixel.Vec
	Location() MapLocation
	Sprite() *pixel.Sprite
	Reference() EntityReference
}

func (s *State) AddEntity(e Entity) bool {
	_, exists := s.entities[e.Reference()]
	if exists {
		return false
	}
	worked := s.attemptToOccupy(e.Location(), e.Reference())
	if !worked {
		return false
	}
	s.entities[e.Reference()] = e
	return true
}

// occupiedBy returns the entity and true if the location is occupied
func (s *State) occupiedBy(location MapLocation) (EntityReference, bool) {
	ent, occupied := s.occupiedLocations[location]
	return ent, occupied
}

// attemptToOccupy will do nothing and return false if the desired location is occupoied
// if it is not, it will occupy it and return true
func (s *State) attemptToOccupy(location MapLocation, entity EntityReference) bool {
	existingEnt, occupied := s.occupiedLocations[location]
	if occupied {
		if existingEnt == entity {
			return true
		}
		return false
	}
	restriction, hasRestriction := s.movementRestrictions[location]
	if hasRestriction {
		if !restriction.EntryAllowed() {
			return false
		}
	}
	s.occupiedLocations[location] = entity
	return true
}

// unoccupy will remove any occupancy in the location. it will return true if occupancy changed
func (s *State) unoccupy(location MapLocation, entity EntityReference) bool {
	existingEntity, wasOccupied := s.occupiedLocations[location]
	if wasOccupied {
		if existingEntity == entity {
			delete(s.occupiedLocations, location)
			return true
		}
		return false
	}
	return false
}
