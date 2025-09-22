package adventure

import (
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
)

type MovementRestriction interface {
	EntryAllowed(*State, EntityId) bool
	CanDashOver() bool
	OnEntryBegin(*State, EntityId)
	OnEntryComplete(*State, EntityId)
}

type BasicMovementRestriction struct {
}

func (m BasicMovementRestriction) EntryAllowed(*State, EntityId) bool {
	return false
}

func (m BasicMovementRestriction) CanDashOver() bool {
	return false
}

func (m BasicMovementRestriction) OnEntryBegin(*State, EntityId) {
}

func (m BasicMovementRestriction) OnEntryComplete(*State, EntityId) {
}

type MovementNotAllowed struct {
	BasicMovementRestriction
}

type MovementJumpTile struct {
	BasicMovementRestriction
}

func (m MovementJumpTile) CanDashOver() bool {
	return true
}

type TeleportTile struct {
	BasicMovementRestriction
	Reference TeleportReference
}

func (t TeleportTile) EntryAllowed(s *State, id EntityId) bool {
	_, isPlayer := s.requirePlayerEntity(id)
	return isPlayer
}

func (t TeleportTile) CanDashOver() bool {
	return true
}

const teleportFadeTime = 0.33

func (t TeleportTile) OnEntryBegin(s *State, id EntityId) {
	player, isPlayer := s.requirePlayerEntity(id)
	if !isPlayer {
		return
	}
	source, found := s.teleports[t.Reference]
	if !found {
		log.Warn().Msgf("teleport source reference not found: %s", t.Reference)
		return
	}
	if source.Destination == "" {
		return
	}
	destination, found := s.teleports[source.Destination]
	if !found {
		log.Warn().Msgf("teleport destination not found: %s", source.Destination)
		return
	}
	s.teleport(player, destination)
}

func (s *State) teleport(entity *Player, destination Teleport) {

	fadeOut := NewFadeOverlay(
		pixel.RGBA{A: 0},
		pixel.RGBA{A: 1},
		1,
		NewBaseOverlay(teleportFadeTime, false, nil),
	)
	fadeIn := NewFadeOverlay(
		pixel.RGBA{A: 1},
		pixel.RGBA{A: 0},
		1,
		NewBaseOverlay(teleportFadeTime, false, nil),
	)
	s.actions.Add(NewSerialActions(
		NewSimpleAction(func(ctx *game.Context, s *State) {
			s.blockInput = true
			s.overlays.Add(fadeOut)
		}),
		NewSleepAction(teleportFadeTime),
		NewWaitForAction(func() bool {
			return !entity.IsMoving()
		}),
		NewSimpleAction(func(ctx *game.Context, s *State) {
			fadeOut.IsComplete = true
			s.overlays.Add(fadeIn)
			entity.TeleportTo(s, destination.Location)
			if destination.ExitDirection != input.NotPressed {
				entity.FacingDirection = destination.ExitDirection
				entity.intentDirection = destination.ExitDirection
				entity.TriggerMovement(s, entity.GetFacingLocation(), MoveStateWalking)
			}
			s.blockInput = false
		}),
		NewChangeCameraAction(func(ctx *game.Context, s *State) Camera {
			return NewFollowCamera(entity.Id, entity.RenderMapLocation(), EntityCameraSpeedPlayerDefault)
		}),
		NewWaitForAction(func() bool {
			return fadeIn.IsComplete
		}),
	))
}
