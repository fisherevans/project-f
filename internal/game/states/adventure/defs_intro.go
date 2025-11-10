package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

func init() {
	attemptsKey := "intro.not_your_paper.attempts"
	registerEventHandler("intro.not_your_paper", func(_ *util.Properties) EventHandler {
		messages := []string{
			"This isn't my paper... I probably shouldn't touch it.",
			"Still not my paper.",
			"Yep, definitely not mine.",
			"Why am I even looking at this.",
			"Someone's gonna think I'm snooping.",
			"I really should stop touching things that aren't mine.",
			"This is getting weirdly habitual.",
			"I swear I'm not a thief, just nosy.",
			"Okay, this is a problem.",
			"Fine, it's not mine, I'll stop... probably",
		}
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				attempts := gameState.RunState().Get(attemptsKey).AsInt(0)
				message := messages[attempts%len(messages)]
				return NewOutput().WithEffects(NewDialogueEffect(message), NewSetRunStateEffect(attemptsKey, attempts+1))
			},
		}.CreateHandler()
	})
	hasPapersKey := "intro.has_papers"
	registerEventHandler("intro.is_your_paper", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				attempts := gameState.RunState().Get(attemptsKey).AsInt(0)
				var message string
				if attempts == 0 {
					message = "Okay, I've got my papers. I need to give them to the instructor."
				} else if attempts < 3 {
					message = "Oh good, this one's actually mine. That shouldn't have been so hard."
				} else {
					message = "Ah, the elusive papers that are actually mine."
				}
				return NewOutput().WithEffects(
					NewDialogueEffect(message),
					NewDeleteEntityEffect(ctx.EntityId()),
					NewSetRunStateEffect(hasPapersKey, true))
			},
		}.CreateHandler()
	})
	papersTurnedInKey := "intro.papers_turned_in"
	type InstructorState struct {
		Blabbed bool
	}
	registerEventHandler("intro.paper_instructor", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[InstructorState]{
			OnInteract: func(ctx EntityContext, gameState GameState, state InstructorState, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				papersTurnedIn := gameState.RunState().Get(papersTurnedInKey).AsBool(false)
				if papersTurnedIn {
					return NewFocusedSequenceBuilder(ctx.EntityId(), event.SourceId).
						WithMiddleEffects(NewDialogueEffect("What do you want? Go find the equipment specialist.")).
						Build()
				}
				hasPapers := gameState.RunState().Get(hasPapersKey).AsBool(false)
				if !hasPapers {
					return NewFocusedSequenceBuilder(ctx.EntityId(), event.SourceId).
						WithMiddleEffects(NewDialogueEffect("Are you confused? Fill out your paper work, and give it to me.")).
						Build()
				}
				attempts := gameState.RunState().Get(attemptsKey).AsInt(0)
				message := "Ah, thank you. Now go find the equipment specialist for your Animech training."
				if attempts >= 2 {
					message = "What were you doing snooping around everyone else's paperwork? I almost wrote you up. Anyways, go find the equipment specialist for your Animech training."
				}
				return NewFocusedSequenceBuilder(ctx.EntityId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect(message),
						NewMutateFollowCameraEffect().WithFollowEntityId("intro.papers_door"),
						NewTimerEffect(1.5),
						NewSetRunStateEffect(papersTurnedInKey, true),
						NewTimerEffect(1.5),
						NewMutateFollowCameraEffect().WithFollowEntityId(gameState.RunState().Get(runStateKeyPlayerId).AsString("")),
					).
					Build()
			},
			EntityZoneActivity: func(ctx EntityContext, gameState GameState, state InstructorState, event *EventEntityZoneActivity) *HandlerOutput {
				papersTurnedIn := gameState.RunState().Get(papersTurnedInKey).AsBool(false)
				if !event.IsEntering || event.ZoneId != "papers.instructions" || state.Blabbed || papersTurnedIn {
					return nil
				}
				state.Blabbed = true
				return NewFocusedSequenceBuilder(ctx.EntityId(), event.EntityId).
					WithMoveCamera(true).
					WithFacePlayer(true).
					WithMiddleEffects(NewChatterEffect(ctx.EntityId(), 3, "You've gotta fill out your paperwork before you can leave...")).
					Build().WithState(state)
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.door.run_state_based", func(props *util.Properties) EventHandler {
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
						WithMode(stateValue(gameState)).
						WithAnimations(map[string][]AnimationReference{
							doorClosed: {
								{Name: "adventure/doors/shield_front_1:closed"},
								{Name: "adventure/doors/shield_front_1:waves"},
							},
							doorOpen: {
								{Name: "adventure/doors/shield_front_1:open"},
							},
						}).
						WithLights(map[string][]LightConfig{
							doorClosed: {
								{Color: "#127fd7", Size: 1.5, Modifier: util.Ptr("pulse_slow")},
							},
						}),
				)
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
	guardedEntryDenialsKey := "intro.do_not_enter.attempts"
	registerEventHandler("intro.guarded_entry", func(props *util.Properties) EventHandler {
		noEntryChatters := []string{
			"Sorry, restricted zone. You’ll need clearance to pass.",
			"Hold up - your ID’s not on the list.",
			"Access denied. You know the rules.",
			"You’ve been flagged for repeated entry attempts.",
			"Central already notified us about you.",
			"You’re persistent, I’ll give you that.",
			"Security’s keeping an eye on your route.",
			"Still not authorized. Try the proper channels.",
			"You’re on record, just so you know.",
			"Another attempt? You’re starting to stand out.",
			"Command says to remind you this area is sealed.",
			"We’ve got your ID tagged facility-wide now.",
			"You must really like closed doors.",
			"Everyone’s been briefed about your “tours.”",
			"You’re not going to find a guard who says yes.",
			"You’re pushing your luck here.",
			"They told us to expect you.",
			"Next stop’s a security report if you keep at it.",
			"You’ve officially become a training example.",
			"You’re done here. Move along.",
		}
		dialogues := []string{
			"Everything in order, recruit? Stay clear of restricted zones.",
			"You again? Don’t wander too close to the barriers.",
			"I’ve seen your name on a few access logs. Keep it clean.",
			"You’re getting a reputation for curiosity. That’s not a compliment.",
			"You’ve tested every guard in this facility. Maybe take a hint.",
		}
		zoneId := props.GetString("zone_id", "")
		walkBack := input.DirectionFromString(props.GetString("walk_back_direction", ""))
		return BasicHandlerBuilder[None]{
			EntityZoneActivity: func(ctx EntityContext, gameState GameState, state None, event *EventEntityZoneActivity) *HandlerOutput {
				if !event.IsEntering || event.ZoneId != zoneId || event.EntityId != gameState.RunState().Get(runStateKeyPlayerId).AsString("") {
					return nil
				}
				attempts := gameState.RunState().Get(guardedEntryDenialsKey).AsInt(0)
				message := noEntryChatters[min(attempts, len(noEntryChatters)-1)]
				return NewFocusedSequenceBuilder(ctx.EntityId(), event.EntityId).
					WithMoveCamera(true).
					WithFacePlayer(true).
					WithMiddleEffects(NewDialogueEffect(message)).
					WithPostEffects(NewTriggerMovementEffect(event.EntityId).WithDirection(walkBack), NewSetRunStateEffect(guardedEntryDenialsKey, attempts+1)).
					Build()
			},
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if ctx.EntityId() != event.TargetId {
					return nil
				}
				attempts := gameState.RunState().Get(guardedEntryDenialsKey).AsInt(0)
				message := dialogues[min(attempts/len(dialogues), len(dialogues)-1)]
				return NewFocusedSequenceBuilder(ctx.EntityId(), event.SourceId).
					WithMiddleEffects(NewDialogueEffect(message)).
					Build()
			},
		}.CreateHandler()
	})

	registerEventHandler("intro.hall_way_npc", func(_ *util.Properties) EventHandler {
		targets := []string{
			"intro.hall_way_end.west",
			"intro.hall_way_end.north",
			"intro.hall_way_end.east",
		}
		messages := util.StringList{
			"I respect curiosity. I just don’t have time for it right now.",
			"If it’s not urgent, take it to your supervisor, not me.",
			"You’re not cleared for curiosity yet, kid. Stick to your assignments.",
			"Whatever you’re thinking, don’t. Trust me, it’s not worth the paperwork.",
			"You’re asking the wrong person, and probably at the wrong time.",
			"Don’t worry, we were all clueless once. Some of us still are.",
			"Curiosity’s fine. Just don’t make it my problem.",
			"You’re eager. That’ll fade after the first audit.",
			"Ask your mentor before you start poking at things.",
			"Whatever you heard, forget it. You’ll sleep better.",
			"That question sounds expensive. I’d drop it.",
			"Keep that energy. The system loves optimism—at first.",
			"You’re not in trouble yet. Emphasis on yet.",
			"The less you know, the longer you last here.",
			"If you value sanity, stay out of the details.",
			"First day? You’ll stop asking by the third.",
			"Everyone starts curious. Then the reports pile up.",
			"Save the questions for someone still enthusiastic.",
			"Don’t look at me like I have answers.",
			"I admire initiative. I also ignore it.",
			"You’ll learn fast. Or not at all.",
			"Be careful what you notice. It notices back.",
			"Advice? Stop asking for advice.",
			"The rules change faster than explanations.",
			"Curiosity’s contagious. Try not to spread it.",
		}
		randomDestination := func() string {
			return targets[rand.Intn(len(targets))]
		}
		type TalkingState struct {
			IsTalking bool
		}
		return BasicHandlerBuilder[TalkingState]{
			Init: func(ctx EntityContext, gameState GameState, state TalkingState) *HandlerOutput {
				return NewOutput().WithEffects(
					NewPushEntityBehaviorEffect(ctx.EntityId()).WithScriptedMotion(EntityBehaviorScriptedMotion{}),
					NewStartScriptedMotionEffect(ctx.EntityId()).WithToEntityId(randomDestination()),
				)
			},
			OnInteract: func(ctx EntityContext, gameState GameState, state TalkingState, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() || ctx.GetBoolMetadata(MetadataKeyIsTalking) {
					return nil
				}
				return NewOutput().WithSerialPlan(
					NewPushEntityBehaviorEffect(ctx.EntityId()).WithFacingEntity(EntityBehaviorFacingEntity{EntityId: util.Ptr(event.SourceId)}),
					NewChatterEffect(ctx.EntityId(), 3, messages.Random()),
					NewPopEntityBehaviorEffect(ctx.EntityId()),
				)
			},
			ScriptedMotionComplete: func(ctx EntityContext, gameState GameState, state TalkingState, event *EventScriptedMotionComplete) *HandlerOutput {
				if event.EntityId != ctx.EntityId() {
					return nil
				}
				var plan []Effect
				if event.WasCanceled {
					log.Info().Str("entity", ctx.EntityId()).Msg("motion canceled, waiting before scheduling next")
					plan = append(plan, NewTimerEffect(rand.Float64()*3))
				}
				plan = append(plan, NewStartScriptedMotionEffect(ctx.EntityId()).WithToEntityId(randomDestination()))
				return NewOutput().WithSerialPlan(plan...)
			},
		}.CreateHandler()
	})

	equipmentDoorStateKey := "intro.equipment_door_state"
	hasEquipmentKeyKey := "intro.has_equipment_key"
	registerEventHandler("intro.equipment_specialist", func(_ *util.Properties) EventHandler {
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}
				hasEquipmentKey := gameState.RunState().Get(hasEquipmentKeyKey).AsBool(false)
				if hasEquipmentKey {
					return NewOutput().WithEffects(NewDialogueEffect("Do you need help finding testing room 1...?"))
				}

				return NewFocusedSequenceBuilder(ctx.EntityId(), event.SourceId).
					WithMiddleEffects(
						NewDialogueEffect("Ah! Hello. Yes, this is the equipment testing facility. No.. No, I'm not a specialist; they're inside the testing rooms."),
						NewDialogueEffect("Here, take this key card. They're waiting for you in testing room 1."),
						NewSetRunStateEffect(hasEquipmentKeyKey, true),
					).
					Build()
			},
		}.CreateHandler()
	})
	registerEventHandler("intro.equipment_door.key_slot", func(props *util.Properties) EventHandler {
		runStateKey := props.GetString("run_state_key", "not_it")
		return BasicHandlerBuilder[None]{
			OnInteract: func(ctx EntityContext, gameState GameState, state None, event *EventOnInteract) *HandlerOutput {
				if event.TargetId != ctx.EntityId() {
					return nil
				}
				defaultMessage := "This looks like a slot for a key card..."
				hasEquipmentKey := gameState.RunState().Get(hasEquipmentKeyKey).AsBool(false)
				if runStateKey == "not_it" {
					if hasEquipmentKey {
						return NewOutput().WithSerialPlan(
							NewPlaySoundEffect("adventure/beeps/error"),
							NewDialogueEffect("Yeah, that key card doesn't work in this slot..."),
						)
					}
					return NewOutput().WithSerialPlan(
						NewDialogueEffect(defaultMessage),
					)
				}
				equipmentDoorState := gameState.RunState().Get(equipmentDoorStateKey).AsString(doorClosed)
				if equipmentDoorState == doorOpen {
					return NewOutput().WithEffects(NewDialogueEffect("The door is already open."))
				}
				if !hasEquipmentKey {
					return NewOutput().WithEffects(NewDialogueEffect(defaultMessage))
				}
				message := "That worked!"
				return NewOutput().WithSerialPlan(
					NewPlaySoundEffect("adventure/beeps/success"),
					NewDialogueEffect(message),
					NewSetRunStateEffect(runStateKey, doorOpen),
					NewPlaySoundEffect("adventure/sealed_door_opens"),
				)
			},
		}.CreateHandler()
	})
}

type FocusedSequenceBuilder struct {
	focusedEntity                          string
	playerId                               string
	moveCamera                             bool
	facePlayer                             bool
	preEffects, middleEffects, postEffects []Effect
}

func NewFocusedSequenceBuilder(focusedEntity, playerId string) *FocusedSequenceBuilder {
	return &FocusedSequenceBuilder{
		focusedEntity: focusedEntity,
		playerId:      playerId,
		postEffects:   []Effect{},
	}
}

func (b *FocusedSequenceBuilder) WithFacePlayer(facePlayer bool) *FocusedSequenceBuilder {
	b.facePlayer = facePlayer
	return b
}

func (b *FocusedSequenceBuilder) WithMoveCamera(moveCamera bool) *FocusedSequenceBuilder {
	b.moveCamera = moveCamera
	return b
}

func (b *FocusedSequenceBuilder) WithPreEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.preEffects = append(b.preEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) WithMiddleEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.middleEffects = append(b.middleEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) WithPostEffects(effects ...Effect) *FocusedSequenceBuilder {
	b.postEffects = append(b.postEffects, effects...)
	return b
}

func (b *FocusedSequenceBuilder) Build() *HandlerOutput {
	var effects []Effect
	effects = append(effects, b.preEffects...)
	effects = append(effects,
		NewMutateEntityBehaviorEffect(b.playerId).WithDisableBy(b.focusedEntity),
		NewMutateNPCEffect(b.focusedEntity).WithTalkingAtEntityId(b.playerId),
	)
	if b.facePlayer {
		effects = append(effects, NewEntityFaceDirectionEffect(b.playerId).WithTargetEntity(b.focusedEntity))
	}
	if b.moveCamera {
		effects = append(effects, NewOverrideCameraEffect().WithFollow(FollowCamera{EntityId: util.Ptr(b.focusedEntity)}))
	}
	effects = append(effects, b.middleEffects...)
	if b.moveCamera {
		effects = append(effects, NewPopCameraOverrideEffect(true))
	}
	if b.facePlayer {
		effects = append(effects, NewResetMovementEffect(b.playerId))
	}
	effects = append(effects,
		NewMutateNPCEffect(b.focusedEntity).WithTalkingAtEntityId(""),
		NewMutateEntityBehaviorEffect(b.playerId).WithEnableBy(b.focusedEntity))
	effects = append(effects, b.postEffects...)
	return NewOutput().WithSerialPlan(effects...)
}
