package adventure

import (
	"math/rand"
	"sync/atomic"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (s *State) processEffects(effects ...events.DispatchedEffect) {
	for _, dispatched := range effects {
		switch e := dispatched.Effect.(type) {
		// generic effects first
		case *events.EffectFunction:
			s.processEffectFunction(dispatched.Source, e)
		case *events.EffectBatch:
			s.processEffectBatch(dispatched.Source, e)
		case *events.EffectTimer:
			s.processEffectTimer(dispatched.Source, e)
		case *events.EffectSetWorldState:
			s.processSetWorldState(dispatched.Source, e)

		// adventure state changes
		case *events.EffectFade:
			s.processEffectFade(dispatched.Source, e)
		case *events.EffectDeactivateFade:
			s.processEffectDeactivateFade(dispatched.Source, e)
		case *events.EffectTriggerCombat:
			s.processEffectTriggerCombat(dispatched.Source, e)

		// chatter and dialogue
		case *events.EffectDialogue:
			s.processEffectDialogue(dispatched.Source, e)
		case *events.EffectChatter:
			s.processEffectChatter(dispatched.Source, e)

		// rpg effects
		case *events.EffectYieldElythium:
			s.processEffectYieldElythium(dispatched.Source, e)

		// camera effects
		case *events.EffectPopCameraOverride:
			s.processEffectPopCameraOverride(dispatched.Source, e)
		case *events.EffectOverrideCamera:
			s.processEffectOverrideCamera(dispatched.Source, e)
		case *events.EffectMutateFollowCamera:
			s.processEffectMutateFollowCamera(dispatched.Source, e)

		// mutate renderers
		case *events.EffectMutateNPC:
			s.processEffectMutateNPC(dispatched.Source, e)
		case *events.EffectMutateModeBasedEntity:
			s.processEffectMutateModeBasedRenderer(dispatched.Source, e)
		case *events.EffectResetModeBasedEntityAnimation:
			s.processEffectResetModeBasedEntityAnimation(dispatched.Source, e)

		// mutate presences
		case *events.EffectMutateBlockingPresence:
			s.processEffectMutateBlockingPresence(dispatched.Source, e)

		// entity behavior
		case *events.EffectPopEntityBehavior:
			s.processEffectPopEntityBehaviorOverride(dispatched.Source, e)
		case *events.EffectPushEntityBehavior:
			s.processEffectOverrideEntityBehavior(dispatched.Source, e)
		case *events.EffectMutateEntityBehavior:
			s.processEffectMutateEntityBehavior(dispatched.Source, e)

		// movement and movement
		case *events.EffectSetEntityLocation:
			s.processEffectSetEntityLocation(dispatched.Source, e)
		case *events.EffectTeleportPlayer:
			s.processEffectTeleportPlayer(dispatched.Source, e)
		case *events.EffectStartScriptedMotion:
			s.processEffectStartScriptedMotion(dispatched.Source, e)
		case *events.EffectResetMovement:
			s.processEffectResetMovement(dispatched.Source, e)
		case *events.EffectEntityFaceDirection:
			s.processEffectEntityFaceDirection(dispatched.Source, e)
		case *events.EffectTriggerMovement:
			s.processEffectTriggerMovement(dispatched.Source, e)

		// entity lifecycles
		case *events.EffectDeleteEntity:
			s.processEffectDeleteEntity(dispatched.Source, e)
		case *events.EffectRegisterEntity:
			s.processEffectRegisterEntity(dispatched.Source, e)

		default:
			log.Warn().Type("effect_type", e).Msg("Unknown effect type, ignoring")
		}
	}
}

func logEffectInfof(source events.EntityContext, e any, messageFormat string, args ...any) {
	logEffect(zerolog.InfoLevel, source, e, messageFormat, args...)
}

func logEffectWarnf(source events.EntityContext, e any, messageFormat string, args ...any) {
	logEffect(zerolog.WarnLevel, source, e, messageFormat, args...)
}

func logEffect(level zerolog.Level, source events.EntityContext, e any, messageFormat string, args ...any) {
	log.WithLevel(level).Str("caller", source.EntityId()).Interface("e", e).Msgf(messageFormat, args...)
}

func (s *State) processEffectFunction(source events.EntityContext, e *events.EffectFunction) {
	if e == nil {
		return
	}
	e.Fn()
	logEffectInfof(source, e, "function executed")
}

func (s *State) processEffectDialogue(source events.EntityContext, e *events.EffectDialogue) {
	if e == nil {
		return
	}
	s.dialogues.Append(NewBasicDialogue(e.Text, e.DialogueId))
	logEffectInfof(source, e, "dialogue added")
}

func (s *State) processEffectTimer(source events.EntityContext, e *events.EffectTimer) {
	if e == nil {
		return
	}
	s.timers.AddTimer(source.EntityId(), e.TimerId, e.DurationSeconds)
	logEffectInfof(source, e, "timer added")
}

func (s *State) processEffectMutateModeBasedRenderer(source events.EntityContext, e *events.EffectMutateModeBasedEntity) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return
	}
	renderer, ok := entity.GetRenderer()
	if !ok {
		logEffectWarnf(source, e, "failed to find renderer")
		return
	}
	modeBased, ok := renderer.(*ModeBasedEntityRenderer)
	if !ok {
		logEffectWarnf(source, e, "failed to find mode based entity for mutation")
		return
	}
	modeBased.WithConfig(&e.ModeBaseRenderConfig)
	logEffectInfof(source, e, "mode based entity mutated")
}

func (s *State) processEffectResetModeBasedEntityAnimation(source events.EntityContext, e *events.EffectResetModeBasedEntityAnimation) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return
	}
	renderer, ok := entity.GetRenderer()
	if !ok {
		logEffectWarnf(source, e, "failed to find renderer")
		return
	}
	modeBased, ok := renderer.(*ModeBasedEntityRenderer)
	if !ok {
		logEffectWarnf(source, e, "failed to find mode based entity for mutation")
		return
	}
	for _, animation := range modeBased.getBasicEntityRenderer(modeBased.currentMode).animations {
		animation.Animation.Reset()
	}
	logEffectInfof(source, e, "mode based entity animations reset")
}

func (s *State) processEffectMutateBlockingPresence(source events.EntityContext, e *events.EffectMutateBlockingPresence) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return
	}
	presence, ok := entity.GetPresence()
	if !ok {
		logEffectWarnf(source, e, "failed to find presence")
		return
	}
	blockingPresence, ok := presence.(*BlockIngressPresence)
	if !ok {
		logEffectWarnf(source, e, "failed to find presence block ingress")
		return
	}
	if e.IsBlockingIngress != nil {
		blockingPresence.isBlockingIngress = *e.IsBlockingIngress
	}
	logEffectInfof(source, e, "block presence mutated")
}

func (s *State) processEffectYieldElythium(source events.EntityContext, e *events.EffectYieldElythium) {
	if e == nil {
		return
	}
	s.run.Elythium += e.Amount
	logEffectInfof(source, e, "elythium granted")
}

func (s *State) processEffectChatter(source events.EntityContext, e *events.EffectChatter) {
	if e == nil {
		return
	}
	s.chatters.Add(newBasicEntityChatter(e.EntityId, e.DurationSeconds, e.Message, e.ChatterId))
	logEffectInfof(source, e, "chatter added")
}

func (s *State) processSetWorldState(source events.EntityContext, e *events.EffectSetWorldState) {
	if e == nil {
		return
	}
	s.setWorldState(e.Key, e.Value, source.EntityId())
	logEffectInfof(source, e, "world state updated")
}

func (s *State) processEffectSetEntityLocation(source events.EntityContext, e *events.EffectSetEntityLocation) {
	if e == nil {
		return
	}
	var toLocation MapLocation
	if e.ToReference != nil {
		tele, ok := s.teleports[TeleportReference(*e.ToReference)]
		if !ok {
			logEffectWarnf(source, e, "failed to find teleport reference")
			return
		}
		toLocation = tele.Location
	} else if e.ToLocation != nil {
		toLocation = MapLocation{X: e.ToLocation.X, Y: e.ToLocation.Y}
	} else if e.ToEntityId != nil {
		entity, ok := s.entities.GetEntity(*e.ToEntityId)
		if !ok {
			logEffectWarnf(source, e, "failed to find entity to teleport to")
			return
		}
		toLocation = entity.GetLocation()
	} else {
		logEffectWarnf(source, e, "failed to find teleport destination")
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to teleport")
		return
	}
	entity.Teleport(toLocation)
}

// processEffectBatch starts a new batch (effects are dispatched internally by the plan executor)
func (s *State) processEffectBatch(source events.EntityContext, e *events.EffectBatch) {
	if e == nil {
		return
	}
	s.planExecutor.StartPlan(source, e)
	logEffectInfof(source, e, "batch started")
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectMutateEntityBehavior(source events.EntityContext, e *events.EffectMutateEntityBehavior) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to mutate")
		return
	}
	if e.DisableBy != nil {
		entity.DisableBehavior(*e.DisableBy)
	}
	if e.EnableBy != nil {
		entity.EnableBehavior(*e.EnableBy)
	}
	if e.Reset != nil && *e.Reset {
		b, ok := entity.GetBehavior()
		if !ok {
			logEffectWarnf(source, e, "failed to find behavior to mutate")
			return
		}
		b.Reset()
	}
	logEffectInfof(source, e, "behavior mutated")
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectDeactivateFade(source events.EntityContext, e *events.EffectDeactivateFade) {
	if e == nil {
		return
	}
	s.overlays.Deactivate(e.FadeId)
	logEffectInfof(source, e, "fade deactivated")
}

// processEffectFade creates a fade overlay
func (s *State) processEffectFade(source events.EntityContext, e *events.EffectFade) {
	if e == nil {
		return
	}

	// Determine colors
	fromColor := pixel.RGBA{R: 0, G: 0, B: 0, A: 0} // transparent
	toColor := pixel.RGBA{R: 0, G: 0, B: 0, A: 1}   // black

	if e.FromColor != nil {
		fromColor = colors.HexString(*e.FromColor)
	}
	if e.ToColor != nil {
		toColor = colors.HexString(*e.ToColor)
	}

	transitions := e.Transitions
	if transitions < 1 {
		transitions = 1
	}

	autoDeactivate := true
	if e.AutoDeactivate != nil {
		autoDeactivate = *e.AutoDeactivate
	}
	// Create base overlay with auto-complete
	base := NewBaseOverlay(e.FadeId, e.DurationSeconds, autoDeactivate)

	// Create fade overlay
	fade := NewFadeOverlay(fromColor, toColor, transitions, base)

	s.overlays.Add(fade)
	logEffectInfof(source, e, "fade overlay added")
}

func (s *State) processEffectResetMovement(source events.EntityContext, e *events.EffectResetMovement) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return
	}
	entity.CancelMovement()
	b, ok := entity.GetBehavior()
	if ok {
		b.Reset()
	}
}

func (s *State) processEffectTriggerMovement(source events.EntityContext, e *events.EffectTriggerMovement) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to move")
		return
	}
	moveState := types.MoveStateWalking
	if e.MoveState != nil {
		moveState = *e.MoveState
	}
	var location MapLocation
	if e.Direction != nil {
		location = entity.GetLocation().Moved(*e.Direction)
	} else if e.Location != nil {
		location = MapLocation{
			X: e.Location.X,
			Y: e.Location.Y,
		}
	}
	s.entities.AttemptMovement(e.EntityId, location, moveState)
}

// processEffectTeleportPlayer creates a plan to teleport the player with fade transition
func (s *State) processEffectTeleportPlayer(source events.EntityContext, e *events.EffectTeleportPlayer) {
	if e == nil {
		return
	}

	// Determine transition style (default to fade)
	transitionStyle := "fade"
	if e.TransitionStyle != nil {
		transitionStyle = *e.TransitionStyle
	}

	// Determine exit direction: use explicit if provided, otherwise get from teleport reference
	var exitDirection *input.Direction
	if e.ExitDirection != nil {
		exitDirection = e.ExitDirection
	} else if e.ToReference != nil {
		// Get exit direction from teleport reference
		if tele, ok := s.teleports[TeleportReference(*e.ToReference)]; ok {
			if tele.ExitDirection != input.NotPressed {
				exitDirection = &tele.ExitDirection
			}
		}
	}

	// Build the teleport batch based on transition style
	var batch *events.EffectBatch

	switch transitionStyle {
	case "fade":
		// Classic fade transition (like the old teleport function)
		fadeOutId := s.planExecutor.GenerateEffectId("fade")
		fadeDuration := .33
		var effects []events.Effect

		effects = append(effects,
			events.NewMutateEntityBehaviorEffect(s.player).WithDisableBy(fadeOutId),
			events.NewFadeEffect(fadeDuration, 1).
				WithFadeId(fadeOutId).
				WithAutoDeactivate(false).
				WithFromColor("#00000000").
				WithToColor("#000000FF"),
			&events.EffectSetEntityLocation{
				EntityId:    s.player,
				ToReference: e.ToReference,
				ToLocation:  e.ToLocation,
				ToEntityId:  e.ToEntityId,
			},
			events.NewMutateFollowCameraEffect().
				WithFollowEntityId(s.player).
				WithResetPosition(true),
		)

		// todo consider adding "wait for idle" for player to stop moving

		if exitDirection != nil && *exitDirection != input.NotPressed {
			effects = append(effects,
				events.NewMutateEntityBehaviorEffect(s.player).WithReset(true),
				events.NewTriggerMovementEffect(s.player).WithDirection(*exitDirection))
		}

		effects = append(effects,
			events.NewFadeEffect(fadeDuration, 1).
				WithAutoDeactivate(true).
				WithFromColor("#000000FF").
				WithToColor("#00000000"),
			events.NewDeactivateFadeEffect(fadeOutId),
			events.NewMutateEntityBehaviorEffect(s.player).
				WithEnableBy(fadeOutId))
		batch = events.NewSerialPlan(effects...)
	case "instant":
		// Instant teleport with no transition
		batch = events.NewSerialPlan(&events.EffectSetEntityLocation{
			EntityId:    s.player,
			ToReference: e.ToReference,
			ToLocation:  e.ToLocation,
			ToEntityId:  e.ToEntityId,
		})
	default:
		logEffectWarnf(source, e, "unknown transition style, using fade")
		return
	}

	s.ExecuteSystemEffects(batch)
	logEffectInfof(source, e, "player teleport batch started with style: %s", transitionStyle)
}

func (s *State) processEffectOverrideCamera(source events.EntityContext, e *events.EffectOverrideCamera) {
	if e == nil {
		return
	}
	if e.Follow == nil {
		logEffectWarnf(source, e, "follow camera override requires a follow camera")
		return
	}
	if e.Follow.EntityId == nil {
		logEffectWarnf(source, e, "currently Id is required to set follow camera")
	}
	target, ok := s.entities.GetEntity(*e.Follow.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to follow")
		return
	}
	location := s.camera.CurrentLocation()
	if e.Follow.ResetPosition {
		location = target.GetPreciseLocation()
	}
	camera := NewFollowCamera(target.GetId(), location, EntityCameraSpeedMedium)
	s.OverrideCamera(camera)
	logEffectInfof(source, e, "camera overriden")
}

func (s *State) processEffectPopCameraOverride(source events.EntityContext, e *events.EffectPopCameraOverride) {
	if e == nil {
		return
	}
	s.PopOverrideCamera(e.MaintainCurrentLocation)
	logEffectInfof(source, e, "camera popped")
}

func (s *State) processEffectMutateFollowCamera(source events.EntityContext, e *events.EffectMutateFollowCamera) {
	if e == nil {
		return
	}
	camera := s.camera
	if override, ok := camera.(*CameraOverride); ok {
		camera = override.newCamera
	}
	followCamera, ok := camera.(*EntityCamera)
	if !ok {
		logEffectWarnf(source, e, "camera is not a follow camera")
		return
	}
	if e.FollowEntityId != nil {
		followCamera.target = *e.FollowEntityId
	}
	if e.ResetPosition != nil && *e.ResetPosition {
		entity, ok := s.entities.GetEntity(followCamera.target)
		if !ok {
			logEffectWarnf(source, e, "failed to find entity to reset position")
			return
		}
		followCamera.location = entity.GetPreciseLocation()
	}
	logEffectInfof(source, e, "follow camera mutated")
}

func (s *State) processEffectMutateNPC(source events.EntityContext, e *events.EffectMutateNPC) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return
	}
	behavior, ok := entity.GetBehavior()
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior for mutation")
		return
	}
	npc, ok := behavior.(*NPCBehavior)
	if !ok {
		logEffectWarnf(source, e, "failed to find npc entity for mutation")
		return
	}
	if e.TalkingAtEntityId != nil {
		npc.talkingTowards = *e.TalkingAtEntityId
	}
	logEffectInfof(source, e, "npc mutated")
}

func (s *State) processEffectTriggerCombat(source events.EntityContext, e *events.EffectTriggerCombat) {
	if e == nil {
		return
	}

	if s.enteringCombat {
		return
	}
	s.enteringCombat = true

	postCombat := func(r game.CombatIntentResult) {
		game.DebugNotification("Combat complete!")
		if !r.PlayerWon {
			game.SetActiveStateIntent(game.TitleIntent{})
			return
		}
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s,
		})
		game.CurrentSave().Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
		s.ExecuteSystemEffects(events.NewDeactivateFadeEffect("combat_fade"))
		s.eventDispatcher.Dispatch(&events.EventCombatComplete{
			CombatId: e.CombatId,
			Result:   "completed", // TODO: serialize result properly
		})
		s.planExecutor.MarkCombatComplete(e.CombatId)
	}

	s.ExecuteSystemEffectsInOrder(
		events.NewMutateEntityBehaviorEffect(s.player).WithDisableBy("combat"),
		&events.EffectFade{
			DurationSeconds: 1,
			AutoDeactivate:  util.Ptr(true),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     6,
		},
		&events.EffectFade{
			FadeId:          "combat_fade",
			DurationSeconds: 3,
			AutoDeactivate:  util.Ptr(false),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     1,
		},
		events.NewFunctionEffect(func() {
			game.SetCustomShader(game.NewSwirlShader(3))
		}),
		events.NewMutateEntityBehaviorEffect(s.player).WithEnableBy("combat"),
		events.NewFunctionEffect(func() {
			s.enteringCombat = false
			game.RemoveCustomShader()
			var opponent rpg.PrimortalType
			if e.Opponent != nil {
				opponent = *e.Opponent
			} else {
				options := []rpg.PrimortalType{
					rpg.Primortal_Volteel.Type,
					rpg.Primortal_Toxmidge.Type,
					rpg.Primortal_Scintail.Type,
					rpg.Primortal_Myceli.Type,
					rpg.Primortal_Pumbl.Type,
				}
				opponent = options[rand.Intn(len(options))]
			}
			game.SetActiveStateIntent(game.CombatIntent{
				Run:        &rpg.Run{},
				Opponent:   opponent,
				Background: e.Background,
				OnComplete: postCombat,
			})
		}),
	)
}

func (s *State) processEffectEntityFaceDirection(source events.EntityContext, e *events.EffectEntityFaceDirection) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find movement for entity")
	}
	entity.SetFacingDirection(e.Direction)
}

func (s *State) processEffectStartScriptedMotion(source events.EntityContext, e *events.EffectStartScriptedMotion) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return
	}
	behavior, ok := entity.GetBehavior()
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior for entity")
		return
	}
	scripted, ok := behavior.(*ScriptedMotionBehavior)
	if !ok {
		logEffectWarnf(source, e, "behavior is not a scripted motion behavior")
		return
	}
	var target MotionTarget
	if e.Location != nil {
		target = NewPathfindingMotion(e.MotionId, entity, LocationFromEvent(*e.Location))
	} else if e.Relative != nil {
		target = NewRelativeMotion(e.MotionId, e.Relative.Direction, e.Relative.Steps)
	} else if e.ToEntityId != nil {
		toEntity, ok := s.entities.GetEntity(*e.ToEntityId)
		if !ok {
			logEffectWarnf(source, e, "failed to find movement for target entity")
			return
		}
		target = NewPathfindingMotion(e.MotionId, entity, toEntity.GetLocation())
	}
	scripted.SetTarget(target)
	logEffectInfof(source, e, "scripted motion started")
}

func (s *State) processEffectOverrideEntityBehavior(source events.EntityContext, e *events.EffectPushEntityBehavior) {
	if e == nil {
		return
	}
	if e.ScriptedMotion == nil {
		logEffectWarnf(source, e, "only scripted motion behaviors can be used to override")
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to override behavior")
		return
	}
	AttachScriptedMotionBehavior(entity)
	logEffectInfof(source, e, "entity behavior overridden")
}

func (s *State) processEffectPopEntityBehaviorOverride(source events.EntityContext, e *events.EffectPopEntityBehavior) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to pop behavior")
		return
	}
	entity.PopBehavior()
	logEffectInfof(source, e, "entity behavior override popped")
}

func (s *State) processEffectDeleteEntity(source events.EntityContext, e *events.EffectDeleteEntity) {
	if e == nil {
		return
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to delete")
		return
	}
	s.eventDispatcher.Unregister(entity.GetEntityContext())
	s.entities.DeleteEntity(e.EntityId)
	logEffectInfof(source, e, "entity deleted")
}

var nextNewEntityId = atomic.Int64{}

func (s *State) processEffectRegisterEntity(source events.EntityContext, e *events.EffectRegisterEntity) {
	if e == nil {
		return
	}
	params := NewEntityParams{
		EntityId: e.EntityId,
	}
	if e.Class != nil {
		params.Class = *e.Class
	}
	if e.SpriteId != nil {
		params.SpriteId = e.SpriteId
	}
	if e.Properties != nil {
		params.Properties = *e.Properties
	}

	// location
	if e.MapLocation != nil {
		params.Location = LocationFromEvent(*e.MapLocation)
	} else if e.EntityLocation != nil {
		toEntity, ok := s.entities.GetEntity(*e.EntityLocation)
		if !ok {
			logEffectWarnf(source, e, "failed to find entity to set location based on")
			return
		}
		params.Location = toEntity.GetLocation()
	}
	if !s.registerParameterizedEntity(params) {
		logEffectWarnf(source, e, "failed to register entity")
		return
	}
	logEffectInfof(source, e, "entity registered: %v", params)
}
