package adventure

import (
	"strings"

	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
)

func init() {
	registerHandlerReference("random_chatters", func(props *util.Properties) EventHandler {
		var chatters util.StringList
		if chattersRaw := props.GetString("chatters", ""); chattersRaw != "" {
			chatters = strings.Split(chattersRaw, "\n")
		}
		duration := props.GetFloat("duration", 4)
		return NewBasicHandler(None{}).
			WithOnInteract(func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				if ctx.GetBoolMetadata(types.MetadataKeyIsTalking) || len(chatters) == 0 {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewMutateNPCEffect(ctx.EntityId()).WithTalkingAtEntityId(event.SourceId),
					NewChatterEffect(ctx.EntityId(), duration, chatters.Random()),
					NewMutateNPCEffect(ctx.EntityId()).WithTalkingAtEntityId(""),
				)
			}).
			CreateHandler()
	})

	registerHandlerReference("door.run_state_based", func(props *util.Properties) EventHandler {
		stateKey := props.GetString("run_state_key", "")
		stateValue := func(gs GameState) string {
			v := gs.RunState().Get(stateKey)
			if !v.Exists() {
				return doorClosed
			}
			if b, ok := v.Value().(bool); ok {
				if b {
					return doorOpen
				} else {
					return doorClosed
				}
			}
			return gs.RunState().Get(stateKey).AsString(doorClosed)
		}
		return BasicHandlerBuilder[None]{
			Init: func(ctx EntityContext, gameState GameState, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).
						WithIsBlockingIngress(stateValue(gameState) == doorClosed),
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode(stateValue(gameState)))
			},
			RunStateUpdated: func(ctx EntityContext, gameState GameState, state None, event *EventRunStateUpdated) *HandlerOutput {
				if event.Key != stateKey {
					return nil
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(ctx.EntityId()).
						WithIsBlockingIngress(stateValue(gameState) == doorClosed),
					NewMutateModeBasedEntityEffect(ctx.EntityId()).
						WithMode(stateValue(gameState)))
			},
		}.CreateHandler()
	})
}

func init() {
	newRegistrarBuilder().byClass("DirectInteraction").registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		AttachBlockIngressPresence(entity, true, NewImpassableImpedance())
		target := params.Properties.GetString("target", "")
		return entity, NewBasicHandler(None{}).
			WithOnInteract(func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				return NewOutput().WithEffects(NewSendEventEffect(EventOnInteract{
					SourceId:              event.SourceId,
					TargetId:              target,
					SourceFacingDirection: event.SourceFacingDirection,
				}))
			}).CreateHandler()
	})
}
