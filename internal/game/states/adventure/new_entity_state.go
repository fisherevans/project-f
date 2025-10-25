package adventure

import (
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"github.com/rs/zerolog/log"
)

type EntityState interface {
}

type EntityStateStringMode struct {
	mode string
}

func NewStringModeEntityState(mode string) *EntityStateStringMode {
	return &EntityStateStringMode{
		mode: mode,
	}
}

func (s *EntityStateStringMode) GetMode() string {
	return s.mode
}

func (s *EntityStateStringMode) SetMode(mode string) {
	s.mode = mode
}

type EntityStateMovementMode struct {
	id     string
	system *EntitySystem
}

func NewMovementModeEntityState(id string, system *EntitySystem) *EntityStateMovementMode {
	return &EntityStateMovementMode{
		id:     id,
		system: system,
	}
}

func (s *EntityStateMovementMode) GetMode() types.MoveState {
	p, ok := s.system.positions[s.id]
	if !ok {
		log.Warn().Str("id", s.id).Msgf("entity movement mode not found")
		return types.MoveStateIdle
	}
	return p.MovementState
}
