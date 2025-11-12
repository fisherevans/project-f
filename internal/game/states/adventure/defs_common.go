package adventure

import (
	"strings"

	"fisherevans.com/project/f/internal/game/anim"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

func init() {
	newRegistrarBuilder().byClass("ModeBasedEntity").registrar(func(params NewEntityParams, system *EntitySystem) (Entity, EventHandler) {
		entity := system.RegisterEntity(params.EntityId, params.Location)
		AttachPresenceFromConfig(entity, params.Properties)
		renderer := AttachModeBasedEntityRenderer(entity)
		if params.SpriteId != nil {
			renderer.getBasicEntityRenderer("").WithAnimations(anim.NewStaticAnimation(atlas.GetTilesheetSpriteById(*params.SpriteId)))
		}
		renderer.WithPropConfigurations(params.Properties)
		return entity, nil
	})
}

func init() {
	registerEventHandler("random_chatters", func(props *util.Properties) EventHandler {
		var chatters util.StringList
		if chattersRaw := props.GetString("chatters", ""); chattersRaw != "" {
			chatters = strings.Split(chattersRaw, "\n")
		}
		if len(chatters) == 0 {
			log.Warn().Msgf("random_chatters has no chatters: %s", props.GetString("dialogues", ""))
			return nil
		}
		duration := props.GetFloat("duration", 4)
		return NewBasicHandler(None{}).
			WithOnInteract(func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				if !IsBehaviorType[*NPCBehavior](thisEntity) || len(chatters) == 0 {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(thisEntity.GetId()).WithFacingEntityId(event.SourceId),
					NewChatterEffect(thisEntity.GetId(), duration, chatters.Random()),
					NewPopEntityBehaviorEffect(thisEntity.GetId()),
				)
			}).
			CreateHandler()
	})

	registerEventHandler("random_dialogues", func(props *util.Properties) EventHandler {
		var dialogues util.StringList
		if chattersRaw := props.GetString("dialogues", ""); chattersRaw != "" {
			dialogues = strings.Split(chattersRaw, "\n")
		}
		if len(dialogues) == 0 {
			log.Warn().Msgf("random_dialogues has no dialogues: %s", props.GetString("dialogues", ""))
			return nil
		}
		return NewBasicHandler(None{}).
			WithOnInteract(func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(thisEntity.GetId()).WithFacingEntityId(event.SourceId),
					NewDialogueEffect(dialogues.Random()),
					NewPopEntityBehaviorEffect(thisEntity.GetId()),
				)
			}).
			CreateHandler()
	})

	registerEventHandler("door.run_state_based", func(props *util.Properties) EventHandler {
		variable := props.GetString("run_state_key", "")
		stateValue := func(gs rpg.GlobalsReader) string {
			v := gs.Get(variable)
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
			return gs.Get(variable).AsString(doorClosed)
		}
		return BasicHandlerBuilder[None]{
			Init: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None) *HandlerOutput {
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(stateValue(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(stateValue(globals)))
			},
			GlobalVariableUpdated: func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventGlobalVariableUpdated) *HandlerOutput {
				if event.Key != variable {
					return nil
				}
				return NewOutput().WithEffects(
					NewMutateBlockingPresenceEffect(thisEntity.GetId()).
						WithIsBlockingIngress(stateValue(globals) == doorClosed),
					NewMutateModeBasedEntityEffect(thisEntity.GetId()).
						WithMode(stateValue(globals)))
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
			WithOnInteract(func(thisEntity EntityReader, globals rpg.GlobalsReader, state None, event *EventOnInteract) *HandlerOutput {
				if thisEntity.GetId() != event.TargetId {
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
