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
	Passable

	GetMode() string
	SetMode(mode string)

	IsInteractable() bool

	Move(adv *State, timeDelta float64) float64
	IsMoving() bool
	Update(adv *State, timeDelta float64)
	PreciseMapLocation() pixel.Vec
	RenderMapLocation() pixel.Vec
	Location() MapLocation
	TeleportTo(s *State, location MapLocation) bool
	RenderScene(target pixel.Target, matrix pixel.Matrix)
	RenderLight(target pixel.Target, matrix pixel.Matrix)
	GetEntityId() EntityId
	Interact(adv *State, source Entity)
	GetRenderZPriority() int
}

type BaseEntity struct {
	Id              EntityId
	Interactable    bool
	RenderZPriority int
	Mode            string
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

func (b *BaseEntity) IsInteractable() bool {
	return b.Interactable
}

func (s *State) AddEntity(e Entity) bool {
	_, exists := s.entities[e.GetEntityId()]
	if exists {
		log.Fatal().Str("entityId", string(e.GetEntityId())).Msg("failed to add entity due to duplicate id")
		return false
	}
	tileState := s.GetTileState(e.Location())
	tileState.AddEntity(e.GetEntityId())
	s.entities[e.GetEntityId()] = e
	return true
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
