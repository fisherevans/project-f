package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (s *State) processEffects(effects ...DispatchedEffect) {
	for _, dispatched := range effects {
		switch e := dispatched.Effect.(type) {
		// generic effects first
		case *EffectFunction:
			s.processEffectFunction(dispatched.Source, e)
		case *EffectBatch:
			s.processEffectBatch(dispatched.Source, e)
		case *EffectTimer:
			s.processEffectTimer(dispatched.Source, e)
		case *EffectWaitForCondition:
			s.processEffectWaitForCondition(dispatched.Source, e)
		case *EffectSetWorldState:
			s.processSetWorldState(dispatched.Source, e)
		case *EffectSetRunState:
			s.processSetRunState(dispatched.Source, e)
		case *EffectLoadMap:
			s.processEffectLoadMap(dispatched.Source, e)
		case *EffectSendEvent:
			s.processEffectSendEvent(dispatched.Source, e)

		// adventure state changes
		case *EffectFade:
			s.processEffectFade(dispatched.Source, e)
		case *EffectDeactivateFade:
			s.processEffectDeactivateFade(dispatched.Source, e)
		case *EffectTriggerCombat:
			s.processEffectTriggerCombat(dispatched.Source, e)

		// chatter and dialogue
		case *EffectDialogue:
			s.processEffectDialogue(dispatched.Source, e)
		case *EffectChatter:
			s.processEffectChatter(dispatched.Source, e)

		// rpg effects
		case *EffectYieldElythium:
			s.processEffectYieldElythium(dispatched.Source, e)

		// camera effects
		case *EffectPopCameraOverride:
			s.processEffectPopCameraOverride(dispatched.Source, e)
		case *EffectOverrideCamera:
			s.processEffectOverrideCamera(dispatched.Source, e)
		case *EffectMutateFollowCamera:
			s.processEffectMutateFollowCamera(dispatched.Source, e)

		// mutate renderers
		case *EffectMutateNPC:
			s.processEffectMutateNPC(dispatched.Source, e)
		case *EffectMutateModeBasedEntity:
			s.processEffectMutateModeBasedRenderer(dispatched.Source, e)
		case *EffectResetModeBasedEntityAnimation:
			s.processEffectResetModeBasedEntityAnimation(dispatched.Source, e)

		// mutate presences
		case *EffectMutateBlockingPresence:
			s.processEffectMutateBlockingPresence(dispatched.Source, e)

		// entity behavior
		case *EffectPopEntityBehavior:
			s.processEffectPopEntityBehaviorOverride(dispatched.Source, e)
		case *EffectPushEntityBehavior:
			s.processEffectOverrideEntityBehavior(dispatched.Source, e)
		case *EffectMutateEntityBehavior:
			s.processEffectMutateEntityBehavior(dispatched.Source, e)

		// movement and movement
		case *EffectSetEntityLocation:
			s.processEffectSetEntityLocation(dispatched.Source, e)
		case *EffectTeleportPlayer:
			s.processEffectTeleportPlayer(dispatched.Source, e)
		case *EffectStartScriptedMotion:
			s.processEffectStartScriptedMotion(dispatched.Source, e)
		case *EffectResetMovement:
			s.processEffectResetMovement(dispatched.Source, e)
		case *EffectEntityFaceDirection:
			s.processEffectEntityFaceDirection(dispatched.Source, e)
		case *EffectTriggerMovement:
			s.processEffectTriggerMovement(dispatched.Source, e)

		// entity lifecycles
		case *EffectDeleteEntity:
			s.processEffectDeleteEntity(dispatched.Source, e)
		case *EffectRegisterEntity:
			s.processEffectRegisterEntity(dispatched.Source, e)

		default:
			log.Warn().Type("effect_type", e).Msg("Unknown effect type, ignoring")
			return
		}
		logEffectInfof(dispatched.Source, dispatched.Effect, "event processed: %T", dispatched.Effect)
	}
}

func logEffectInfof(source EntityContext, e any, messageFormat string, args ...any) {
	logEffect(zerolog.InfoLevel, source, e, messageFormat, args...)
}

func logEffectWarnf(source EntityContext, e any, messageFormat string, args ...any) {
	logEffect(zerolog.WarnLevel, source, e, messageFormat, args...)
}

func logEffect(level zerolog.Level, source EntityContext, e any, messageFormat string, args ...any) {
	log.WithLevel(level).Str("caller", source.EntityId()).Interface("e", e).Msgf(messageFormat, args...)
}

func (s *State) processEffectFunction(source EntityContext, e *EffectFunction) {
	e.Fn()
	logEffectInfof(source, e, "function executed")
}

func (s *State) processEffectDialogue(source EntityContext, e *EffectDialogue) {
	s.dialogues.Append(NewBasicDialogue(e.Text, e.DialogueId))
	logEffectInfof(source, e, "dialogue added")
}

func (s *State) processEffectTimer(source EntityContext, e *EffectTimer) {
	s.timers.AddTimer(source.EntityId(), e.TimerId, e.DurationSeconds)
	logEffectInfof(source, e, "timer added")
}

func (s *State) processEffectWaitForCondition(source EntityContext, e *EffectWaitForCondition) {
	s.conditions.AddCondition(e.ConditionId, e.Check)
}

func (s *State) processEffectMutateModeBasedRenderer(source EntityContext, e *EffectMutateModeBasedEntity) {
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

func (s *State) processEffectResetModeBasedEntityAnimation(source EntityContext, e *EffectResetModeBasedEntityAnimation) {
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

func (s *State) processEffectMutateBlockingPresence(source EntityContext, e *EffectMutateBlockingPresence) {
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

func (s *State) processEffectYieldElythium(source EntityContext, e *EffectYieldElythium) {
	newValue := max(s.RunState().Get(rpg.RunStateKeyElythium).AsInt(0)+e.Amount, 0)
	s.runState.Set(rpg.RunStateKeyElythium, newValue)
	logEffectInfof(source, e, "elythium updated to: %d", newValue)
}

func (s *State) processEffectChatter(source EntityContext, e *EffectChatter) {
	s.chatters.Add(newBasicEntityChatter(e.EntityId, e.DurationSeconds, e.Message, e.ChatterId))
	logEffectInfof(source, e, "chatter added")
}

func (s *State) processSetWorldState(source EntityContext, e *EffectSetWorldState) {
	s.worldState.Set(e.Key, e.Value)
	logEffectInfof(source, e, "world state updated")
}

func (s *State) processSetRunState(source EntityContext, e *EffectSetRunState) {
	s.runState.Set(e.Key, e.Value)
	logEffectInfof(source, e, "run state updated")
}

func (s *State) processEffectSetEntityLocation(source EntityContext, e *EffectSetEntityLocation) {
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
func (s *State) processEffectBatch(source EntityContext, e *EffectBatch) {
	logEffectInfof(source, e, "batch starting")
	s.planExecutor.StartPlan(source, e)
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectMutateEntityBehavior(source EntityContext, e *EffectMutateEntityBehavior) {
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
func (s *State) processEffectDeactivateFade(source EntityContext, e *EffectDeactivateFade) {
	s.overlays.Deactivate(e.FadeId)
	logEffectInfof(source, e, "fade deactivated")
}

// processEffectFade creates a fade overlay
func (s *State) processEffectFade(source EntityContext, e *EffectFade) {

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

func (s *State) processEffectResetMovement(source EntityContext, e *EffectResetMovement) {
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

func (s *State) processEffectTriggerMovement(source EntityContext, e *EffectTriggerMovement) {
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
	entity.AttemptMovement(location, moveState)
}

// processEffectTeleportPlayer creates a plan to teleport the player with fade transition
func (s *State) processEffectTeleportPlayer(source EntityContext, e *EffectTeleportPlayer) {

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
	var batch *EffectBatch

	switch transitionStyle {
	case "fade":
		// Classic fade transition (like the old teleport function)
		fadeOutId := s.planExecutor.GenerateEffectId("fade")
		fadeDuration := .33
		var effects []Effect

		effects = append(effects,
			NewMutateEntityBehaviorEffect(s.player).WithDisableBy(fadeOutId),
			NewFadeEffect(fadeDuration, 1).
				WithFadeId(fadeOutId).
				WithAutoDeactivate(false).
				WithFromColor("#00000000").
				WithToColor("#000000FF"),
			NewWaitForConditionEffect(func(s *State, _ float64) bool {
				e, _ := s.entities.GetEntity(s.player)
				return !e.IsMoving()
			}),
			&EffectSetEntityLocation{
				EntityId:    s.player,
				ToReference: e.ToReference,
				ToLocation:  e.ToLocation,
				ToEntityId:  e.ToEntityId,
			},
			NewMutateFollowCameraEffect().
				WithFollowEntityId(s.player).
				WithResetPosition(true),
		)

		if exitDirection != nil && *exitDirection != input.NotPressed {
			effects = append(effects,
				NewMutateEntityBehaviorEffect(s.player).WithReset(true),
				NewTriggerMovementEffect(s.player).WithDirection(*exitDirection))
		}

		effects = append(effects,
			NewFadeEffect(fadeDuration, 1).
				WithAutoDeactivate(true).
				WithFromColor("#000000FF").
				WithToColor("#00000000"),
			NewDeactivateFadeEffect(fadeOutId),
			NewMutateEntityBehaviorEffect(s.player).
				WithEnableBy(fadeOutId))
		batch = NewSerialPlan(effects...)
	case "instant":
		// Instant teleport with no transition
		batch = NewSerialPlan(&EffectSetEntityLocation{
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

func (s *State) processEffectOverrideCamera(source EntityContext, e *EffectOverrideCamera) {
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

func (s *State) processEffectPopCameraOverride(source EntityContext, e *EffectPopCameraOverride) {
	s.PopOverrideCamera(e.MaintainCurrentLocation)
	logEffectInfof(source, e, "camera popped")
}

func (s *State) processEffectMutateFollowCamera(source EntityContext, e *EffectMutateFollowCamera) {
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

func (s *State) processEffectMutateNPC(source EntityContext, e *EffectMutateNPC) {
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

func (s *State) processEffectTriggerCombat(source EntityContext, e *EffectTriggerCombat) {

	if s.enteringCombat {
		return
	}
	s.enteringCombat = true

	postCombat := func(r game.CombatIntentResult) {
		game.DebugNotificationf("Combat complete!")
		if !r.PlayerWon {
			game.SetActiveStateIntent(game.StartupDeviceIntent{})
			return
		}
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s,
		})
		game.CurrentSave().Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
		s.ExecuteSystemEffects(NewDeactivateFadeEffect("combat_fade"))
		s.eventDispatcher.Dispatch(&EventCombatComplete{
			CombatId: e.CombatId,
			Result:   "completed", // TODO: serialize result properly
		})
		s.planExecutor.MarkCombatComplete(e.CombatId)
	}

	s.ExecuteSystemEffectsInOrder(
		NewMutateEntityBehaviorEffect(s.player).WithDisableBy("combat"),
		&EffectFade{
			DurationSeconds: 1,
			AutoDeactivate:  util.Ptr(true),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     6,
		},
		&EffectFade{
			FadeId:          "combat_fade",
			DurationSeconds: 3,
			AutoDeactivate:  util.Ptr(false),
			FromColor:       util.Ptr("#00000000"),
			ToColor:         util.Ptr("#000000FF"),
			Transitions:     1,
		},
		NewFunctionEffect(func() {
			game.SetCustomShader(game.NewSwirlShader(3))
		}),
		NewMutateEntityBehaviorEffect(s.player).WithEnableBy("combat"),
		NewFunctionEffect(func() {
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
				Run:        s.run,
				Opponent:   opponent,
				Background: e.Background,
				OnComplete: postCombat,
			})
		}),
	)
}

func (s *State) processEffectEntityFaceDirection(source EntityContext, e *EffectEntityFaceDirection) {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find movement for entity")
	}
	entity.SetFacingDirection(e.Direction)
}

func (s *State) processEffectStartScriptedMotion(source EntityContext, e *EffectStartScriptedMotion) {
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
		target = NewPathfindingMotion(e.MotionId, entity, *e.Location)
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

func (s *State) processEffectOverrideEntityBehavior(source EntityContext, e *EffectPushEntityBehavior) {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to override behavior")
		return
	}
	if e.ScriptedMotion != nil {
		AttachScriptedMotionBehavior(entity)
	} else if e.FacingEntity != nil {
		facing := ""
		if e.FacingEntity.EntityId != nil {
			facing = *e.FacingEntity.EntityId
		}
		AttachFaceEntityBehavior(entity, facing)
	}
	logEffectInfof(source, e, "entity behavior overridden")
}

func (s *State) processEffectPopEntityBehaviorOverride(source EntityContext, e *EffectPopEntityBehavior) {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to pop behavior")
		return
	}
	entity.PopBehavior()
	logEffectInfof(source, e, "entity behavior override popped")
}

func (s *State) processEffectDeleteEntity(source EntityContext, e *EffectDeleteEntity) {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to delete")
		return
	}
	s.eventDispatcher.Unregister(entity.GetEntityContext())
	s.entities.DeleteEntity(e.EntityId)
	logEffectInfof(source, e, "entity deleted")
}

func (s *State) processEffectRegisterEntity(source EntityContext, e *EffectRegisterEntity) {
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
		params.Location = *e.MapLocation
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

func (s *State) processEffectLoadMap(source EntityContext, e *EffectLoadMap) {
	if e == nil {
		return
	}
	if err := game.CurrentSave().Save(); err != nil {
		game.DebugNotificationf("failed to save: %v", err)
	}
	s.ExecuteSystemEffectsInOrder(
		NewMutateEntityBehaviorEffect(s.player).WithDisableBy("map_load"),
		NewFadeEffect(1, 1).
			WithAutoDeactivate(false).
			WithFromColor("#0000").
			WithToColor("#000f"),
		NewFunctionEffect(func() {
			game.SetActiveStateIntent(game.AdventureIntent{
				MapName: e.MapName,
			})
		}),
	)
	logEffectInfof(source, e, "adventure intent set")
}

func (s *State) processEffectSendEvent(source EntityContext, e *EffectSendEvent) {
	if e == nil {
		return
	}
	s.eventDispatcher.Dispatch(e.Event)
	logEffectInfof(source, e, "event sent")
}
