package adventure

import (
	"fisherevans.com/project/f/internal/game/events"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

type EntityId string

type Entity interface {
	events.EntityContext
	Passable

	PreciseMapLocation() pixel.Vec
	RenderMapLocation() pixel.Vec
	Location() MapLocation

	GetMode() string
	SetMode(mode string)

	IsInteractable() bool

	IsMoving() bool
	TeleportTo(s *State, location MapLocation) bool
	GetEntityId() EntityId

	RenderScene(target pixel.Target, matrix pixel.Matrix)
	RenderLight(target pixel.Target, matrix pixel.Matrix)
	GetRenderZPriority() int
}

type BaseEntity struct {
	id              EntityId
	isInteractable  bool
	renderZPriority int
	mode            string
}

func (b *BaseEntity) Id() string {
	return string(b.id)
}

func (b *BaseEntity) Mode() string {
	return b.mode
}

func (b *BaseEntity) GetEntityId() EntityId {
	return b.id
}

func (b *BaseEntity) GetRenderZPriority() int {
	return b.renderZPriority
}

func (b *BaseEntity) GetMode() string {
	return b.mode
}

func (b *BaseEntity) SetMode(mode string) {
	b.mode = mode
}

func (b *BaseEntity) IsInteractable() bool {
	return b.isInteractable
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
