package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/input"
)

type NPCBehavior struct {
	entityId        EntityId
	doesMove        bool
	horizOnly       bool
	idleChance      float64
	maxIdleDuration float64
	idleDuration    float64
}

func NewNPCBehavior(entityId EntityId, doesMove bool, horizOnly bool) *NPCBehavior {
	return &NPCBehavior{
		entityId:        entityId,
		doesMove:        doesMove,
		horizOnly:       horizOnly,
		idleChance:      0.05,
		maxIdleDuration: 6,
	}
}

func (nb *NPCBehavior) Update(s *State, timeDelta float64) {
	movement := s.movementController.Get(nb.entityId)
	if movement == nil {
		return
	}

	// Don't do anything if moving
	if movement.IsMoving() {
		return
	}

	// Check if NPC is talking
	npc, ok := s.entities[nb.entityId].(*NPC)
	if ok && npc.Talking {
		// Face the entity they're talking to
		if ent, exists := s.entities[npc.TalkingTowards]; exists {
			npcLoc := movement.PreciseLocation()
			targetLoc := s.movementController.Get(ent.GetEntityId())
			if targetLoc != nil {
				direction := DirectionTowards(npcLoc, targetLoc.PreciseLocation())
				movement.SetFacingDirection(direction)
			}
		}
		return
	}

	// Don't move if configured not to
	if !nb.doesMove {
		return
	}

	// Handle idle duration
	if nb.idleDuration > 0 {
		nb.idleDuration -= timeDelta
		return
	}

	// Random chance to idle
	if rand.Float64() < nb.idleChance {
		nb.idleDuration = rand.Float64() * nb.maxIdleDuration
		return
	}

	// Try to move in facing direction
	if s.movementController.TriggerMovement(s, nb.entityId, movement.GetFacingLocation(), MoveStateWalking) {
		return
	}

	// Pick a new direction
	var dir input.Direction
	if nb.horizOnly {
		if movement.FacingDirection() == input.NotPressed {
			dir = input.Left
		} else {
			dir = movement.FacingDirection().Opposite()
		}
	} else {
		dir = input.Directions[int(rand.Float64()*float64(len(input.Directions)))]
	}
	movement.SetFacingDirection(dir)
	s.movementController.TriggerMovement(s, nb.entityId, movement.GetLocationInDirection(dir), MoveStateWalking)
}

func (nb *NPCBehavior) OnMovementComplete(s *State, movement *MovementState) bool {
	// NPCs don't chain movements
	return false
}
