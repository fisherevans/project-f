package adventure

import (
	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/tiles"
	"github.com/gopxl/pixel/v2"
)

func init() {
	targetRegistration().byTile(tiles.Player).registrar(func(entityId string, location MapLocation, mapEntity *resources.Entity, system *EntitySystem) events.EventHandler {
		state := NewMovementModeEntityState(entityId, system)
		renderer := NewMovementBasedEntityRenderer(entityId, system)
		dashPlayerLight := &Light{
			RenderDetails: LightRenderDetails{
				SizeScale: 1.5,
				ColorMask: colors.HexString("#88f"),
			},
		}
		normalPlayerLight := &Light{
			RenderDetails: LightRenderDetails{
				SizeScale: 1.5,
				ColorMask: colors.HexString("#888"),
			},
		}
		for moveState, animations := range map[types.MoveState]map[input.Direction]*anim.AnimatedSprite{
			types.MoveStateIdle:    anim.AshaIdle(atlas),
			types.MoveStateWalking: anim.AshaWalk(atlas),
			types.MoveStateRunning: anim.AshaRun(atlas),
			types.MoveStateDashing: anim.Dash(atlas),
		} {
			for direction, animation := range animations {
				moveStateRenderer := NewBasicEntityRenderer().WithAnimations(animation).
					WithAnimationSpeedScaler(NewMoveAnimationSpeedScaler(system, entityId)).
					WithAnimationOriginOffset(pixel.V(0, 0.25))
				if moveState == types.MoveStateDashing {
					moveStateRenderer.WithLights(dashPlayerLight)
				} else {
					moveStateRenderer.WithLights(normalPlayerLight)
				}
				renderer.WithMovementStateRenderer(moveState, direction, moveStateRenderer)
			}
		}
		presence := newBlockIngressPresence(false)
		behavior := NewPlayerBehavior(entityId, system)
		position := system.RegisterEntity(entityId, location, presence, behavior, renderer, state)
		position.MovementSpeeds = map[types.MoveState]float64{
			types.MoveStateWalking: characterSpeed,
			types.MoveStateRunning: characterSpeed * 1.75,
			types.MoveStateDashing: characterSpeed * 1.5,
		}
		// todo this seems gross
		system.state.player = entityId
		system.state.camera = NewFollowCamera(entityId, location.ToVec(), EntityCameraSpeedPlayerDefault)
		system.state.worldState.Set("player_id", entityId)
		return nil
	})
}

type PlayerBehavior struct {
	id              string
	system          *EntitySystem
	isEnabled       func() bool
	intentDirection input.Direction
	intentDuration  float64
}

func NewPlayerBehavior(id string, system *EntitySystem) *PlayerBehavior {
	return &PlayerBehavior{
		id:        id,
		system:    system,
		isEnabled: func() bool { return !system.state.blockInput },
	}
}

func (b *PlayerBehavior) MovementComplete(dispatcher Dispatcher) {
	b.triggerMovement(dispatcher)
}

func (b *PlayerBehavior) triggerMovement(dispatcher Dispatcher) {
	if b.isEnabled() && game.Controls[*State]().DPad().IsPressed() && b.intentDuration > 0.075 {
		direction := game.Controls[*State]().DPad().GetDirection()
		moveState := types.MoveStateWalking
		if game.Controls[*State]().ButtonB().IsPressed() {
			moveState = types.MoveStateRunning
		}
		dispatcher.ProcessEffects(events.Effect{
			TriggerMovement: events.NewTriggerMovementEffect(b.id, direction).WithMoveState(moveState),
		})
	}
}

func (b *PlayerBehavior) Update(timeDelta float64, p *EntityPosition, dispatcher Dispatcher) {
	if !b.isEnabled() {
		return
	}
	// trigger running or face new direction after movement
	if p.IsMoving() {
		// todo - make effects?
		if game.Controls[*State]().DPad().IsPressed() {
			b.intentDirection = game.Controls[*State]().DPad().GetDirection()
		}
		if game.Controls[*State]().ButtonB().IsPressed() {
			if p.MovementState == types.MoveStateWalking {
				p.MovementState = types.MoveStateRunning
			}
		} else {
			if p.MovementState == types.MoveStateRunning {
				p.MovementState = types.MoveStateWalking
			}
		}
	}
	// face direction of intent after movement
	// todo effect?
	if b.intentDirection != input.NotPressed && b.intentDirection != p.FacingDirection {
		p.FacingDirection = b.intentDirection
	}
	// trigger movement if player is pressing a direction
	if game.Controls[*State]().DPad().IsPressed() {
		direction := game.Controls[*State]().DPad().GetDirection()
		p.FacingDirection = direction
		if b.intentDirection != direction {
			b.intentDirection = direction
			b.intentDuration = 0
		}
		b.intentDuration += timeDelta
	}
	b.triggerMovement(dispatcher)
	// interact with item if player is pressing A
	if game.Controls[*State]().ButtonA().JustPressedOrRepeated() && p.FacingDirection != input.NotPressed {
		interactLocation := p.Location.MovedDelta(p.FacingDirection.GetVector())
		for targetEntityId := range b.system.occupyingEntityIds(interactLocation) {
			dispatcher.EmitEvents(events.EventOnInteract{
				SourceId:        b.id,
				SourceDirection: p.FacingDirection,
				TargetId:        targetEntityId,
			})
		}
	}
}

// todo dashing
//func (p *Player) DashTowards(s *State, direction input.Direction, location MapLocation) {
//	location = p.findDashDestination(s, direction, location)
//	p.TriggerMovement(s, location, MoveStateDashing)
//}
//
//func (p *Player) findDashDestination(s *State, direction input.Direction, location MapLocation) MapLocation {
//	for {
//		ts := s.GetTileState(location)
//		moved := false
//		for _, entityId := range ts.EntitiesWithin {
//			if entity, ok := s.entities[entityId]; ok {
//				if _, ok := entity.(*EntityDashGap); ok {
//					moved = true
//					location = location.MovedDelta(direction.GetVector())
//				}
//			}
//		}
//		if !moved {
//			return location
//		}
//	}
//}
