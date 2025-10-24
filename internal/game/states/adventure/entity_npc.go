package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	registerDynamicEntity().byClass("NPC").byTile(tiles.NPC).register(func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler) {
		doesMove := true
		horizOnly := false
		walkSpeed := 2.0
		
		switch mapEntity.GetStringMetadata("movement", "") {
		case "static":
			doesMove = false
		case "horiz":
			horizOnly = true
		}
		switch mapEntity.GetStringMetadata("speed", "") {
		case "fast":
			walkSpeed = 4.0
		}
		
		npc := &NPC{
			BaseEntity: BaseEntity{
				id:             entityId,
				isInteractable: true,
			},
			Passable: newPassablePreventIngress(true),
			Animations: map[MoveState]map[input.Direction]*anim.AnimatedSprite{
				MoveStateIdle:    anim.AshaIdle(atlas),
				MoveStateWalking: anim.AshaWalk(atlas),
				MoveStateRunning: anim.AshaRun(atlas),
			},
			ColorMask: pixel.RGB(rand.Float64(), rand.Float64(), rand.Float64()),
		}
		
		// Register with movement controller and behavior
		speeds := map[MoveState]float64{
			MoveStateWalking: walkSpeed,
		}
		movementState := s.movementController.Register(entityId, location, speeds)
		npc.SetMovementState(movementState)
		
		behavior := NewNPCBehavior(entityId, doesMove, horizOnly)
		s.behaviors[entityId] = behavior
		
		return npc, events.NewBasicHandler(None{}).
			WithOnInteract(func(ctx events.EntityContext, world events.WorldStateReader, state None, event *events.EventOnInteract) *events.HandlerOutput {
				if ctx.Id() != event.TargetId {
					return nil
				}
				npcCtx, ok := ctx.(NPCEntityContext)
				if !ok {
					return nil
				}
				if npcCtx.IsTalking() {
					return nil
				}
				return events.NewOutput().WithSerialPlan(
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:          ctx.Id(),
							TalkingAtEntityId: util.Ptr(world.GetAsString("player_id")),
						},
						Chatter: &events.EffectChatter{
							EntityId:        ctx.Id(),
							DurationSeconds: 5,
							Message:         util.OneOffDialogues.Random(),
						},
					},
					events.Effect{
						MutateNPC: &events.EffectMutateNPC{
							EntityId:  ctx.Id(),
							IsTalking: util.Ptr(false),
						},
					},
				)
			}).
			CreateHandler()
	})
}

type NPCEntityContext interface {
	events.EntityContext
	IsTalking() bool
}

type NPC struct {
	BaseEntity
	Passable
	Animations             map[MoveState]map[input.Direction]*anim.AnimatedSprite
	ColorMask              pixel.RGBA
	movementState          *MovementState
	lastAnimationDirection input.Direction
	lastAnimationState     MoveState

	Talking        bool
	TalkingTowards EntityId
}

func (n *NPC) SetMovementState(state *MovementState) {
	n.movementState = state
}

func (n *NPC) IsTalking() bool {
	return n.Talking
}

func (n *NPC) RenderScene(target pixel.Target, matrix pixel.Matrix) {
	if n.movementState == nil {
		return
	}
	
	animations, ok := n.Animations[n.movementState.MoveState()]
	if !ok {
		return
	}
	
	animation, ok := animations[n.movementState.FacingDirection()]
	if !ok {
		return
	}
	
	sprite := animation.Sprite()
	if n.ColorMask.A > 0 {
		sprite.DrawColorMask(target, matrix, n.ColorMask)
	} else {
		sprite.Draw(target, matrix)
	}
}

func (n *NPC) RenderLight(target pixel.Target, matrix pixel.Matrix) {
	// NPC has no light
}

func (n *NPC) Location() MapLocation {
	if n.movementState == nil {
		return MapLocation{}
	}
	return n.movementState.Location()
}

func (n *NPC) PreciseMapLocation() pixel.Vec {
	if n.movementState == nil {
		return pixel.Vec{}
	}
	return n.movementState.PreciseLocation()
}

func (n *NPC) RenderMapLocation() pixel.Vec {
	location := n.PreciseMapLocation()
	// Add Y offset for character rendering (feet position)
	return location.Add(pixel.V(0, 0.25))
}

func (n *NPC) IsMoving() bool {
	if n.movementState == nil {
		return false
	}
	return n.movementState.IsMoving()
}

func (n *NPC) TeleportTo(s *State, location MapLocation) bool {
	return s.movementController.TeleportTo(s, n.GetEntityId(), location)
}
