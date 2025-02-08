package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"github.com/gopxl/pixel/v2"
	"log"
)

type EntityId string

func (i EntityId) GetEntityId() EntityId {
	return i
}

type Entity interface {
	Move(adv *State, timeDelta float64) float64
	Update(ctx *game.Context, adv *State, timeDelta float64)
	RenderMapLocation() pixel.Vec
	Location() MapLocation
	Render(target pixel.Target, matrix pixel.Matrix)
	GetEntityId() EntityId
	Interact(ctx *game.Context, adv *State, source Entity)
}

func (s *State) AddEntity(e Entity) bool {
	_, exists := s.entities[e.GetEntityId()]
	if exists {
		log.Fatal("failed to add entity due to id conflict", e)
		return false
	}
	worked := s.attemptToOccupy(e.Location(), e.GetEntityId())
	if !worked {
		log.Fatal("failed to add entity due to location conflict", e)
		return false
	}
	s.entities[e.GetEntityId()] = e
	return true
}

// occupiedBy returns the entity and true if the location is occupied
func (s *State) occupiedBy(location MapLocation) (EntityId, bool) {
	ent, occupied := s.occupiedLocations[location]
	return ent, occupied
}

// attemptToOccupy will do nothing and return false if the desired location is occupoied
// if it is not, it will occupy it and return true
func (s *State) attemptToOccupy(location MapLocation, entityId EntityId) bool {
	existingEnt, occupied := s.occupiedLocations[location]
	if occupied {
		if existingEnt == entityId {
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
	s.occupiedLocations[location] = entityId
	return true
}

// unoccupy will remove any occupancy in the location. it will return true if occupancy changed
func (s *State) unoccupy(location MapLocation, entityId EntityId) bool {
	existingEntity, wasOccupied := s.occupiedLocations[location]
	if wasOccupied {
		if existingEntity == entityId {
			delete(s.occupiedLocations, location)
			return true
		}
		return false
	}
	return false
}
