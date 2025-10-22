package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntityId string

func (i EntityId) GetEntityId() EntityId {
	return i
}

type Entity interface {
	Move(adv *State, timeDelta float64) float64
	IsMoving() bool
	Update(adv *State, timeDelta float64)
	PreciseMapLocation() pixel.Vec
	RenderMapLocation() pixel.Vec
	Location() MapLocation
	RenderScene(target pixel.Target, matrix pixel.Matrix)
	RenderLight(target pixel.Target, matrix pixel.Matrix)
	GetEntityId() EntityId
	Interact(adv *State, source Entity)
	GetRenderZPriority() int

	GetMode() string
	SetMode(mode string)
	IsPassable() bool
	SetIsPassable(p bool)
}

type BaseEntity struct {
	Id              EntityId
	Passable        bool
	RenderZPriority int
	Mode            string
}

func (b *BaseEntity) IsPassable() bool {
	return b.Passable
}

func (b *BaseEntity) SetIsPassable(p bool) {
	b.Passable = p
}

func (b *BaseEntity) GetEntityId() EntityId {
	return b.Id
}

func (b *BaseEntity) GetRenderZPriority() int {
	return b.RenderZPriority
}

func (b *BaseEntity) GetMode() string {
	return b.Mode
}

func (b *BaseEntity) SetMode(mode string) {
	b.Mode = mode
}

func (s *State) AddEntity(e Entity) bool {
	_, exists := s.entities[e.GetEntityId()]
	if exists {
		log.Fatal().Str("entityId", string(e.GetEntityId())).Msg("failed to add entity due to duplicate id")
		return false
	}
	if !e.IsPassable() && !s.attemptToOccupy(e.Location(), e.GetEntityId()) {
		log.Fatal().Str("entityId", string(e.GetEntityId())).Str("location", e.Location().String()).Msg("failed to add entity due to location conflict")
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
		if !restriction.EntryAllowed(s, entityId) {
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

func (s *State) requirePlayerEntity(id EntityId) (*Player, bool) {
	e, isE := s.entities[id]
	if !isE {
		return nil, false
	}
	player, isPlayer := e.(*Player)
	return player, isPlayer
}

type eventEntityWrapper struct {
	entity Entity
}

func newEventEntityWrapper(entity Entity) *eventEntityWrapper {
	return &eventEntityWrapper{
		entity: entity,
	}
}

func (w eventEntityWrapper) Id() string {
	return string(w.entity.GetEntityId())
}

func (w eventEntityWrapper) Mode() string {
	return w.entity.GetMode()
}

func (w eventEntityWrapper) Position() *events.EntityPosition {
	loc := w.entity.Location()
	return &events.EntityPosition{
		X:        loc.X,
		Y:        loc.Y,
		IsMoving: w.entity.IsMoving(),
	}
}
