package adventure

import (
	"math/rand"

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
		case *events.EffectPlan:
			s.processEffectPlan(dispatched.Source, e)
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
		case *events.EffectPopEntityBehaviorOverride:
			s.processEffectPopEntityBehaviorOverride(dispatched.Source, e)
		case *events.EffectOverrideEntityBehavior:
			s.processEffectOverrideEntityBehavior(dispatched.Source, e)
		case *events.EffectMutateEntityBehavior:
			s.processEffectMutateEntityBehavior(dispatched.Source, e)

		// position and movement
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
	entity := s.entities.GetEntity(e.EntityId)
	modeBased, ok := entity.EntityRenderer.(*ModeBasedEntityRenderer)
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
	entity := s.entities.GetEntity(e.EntityId)
	modeBased, ok := entity.EntityRenderer.(*ModeBasedEntityRenderer)
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
	entity := s.entities.GetEntity(e.EntityId)
	presence, ok := entity.EntityPresence.(*blockIngressPresence)
	if !ok {
		logEffectWarnf(source, e, "failed to find presence block ingress")
		return
	}
	if e.IsBlockingIngress != nil {
		presence.isBlockingIngress = *e.IsBlockingIngress
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
		entity := s.entities.GetEntity(*e.ToEntityId)
		toLocation = entity.GetPrimaryLocation()
	} else {
		logEffectWarnf(source, e, "failed to find teleport destination")
		return
	}
	s.entities.TeleportEntity(e.EntityId, toLocation)
}

// processEffectPlan starts a new plan (effects are dispatched internally by the plan executor)
func (s *State) processEffectPlan(source events.EntityContext, e *events.EffectPlan) {
	if e == nil {
		return
	}
	s.planExecutor.StartPlan(source, e)
	logEffectInfof(source, e, "plan started")
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectMutateEntityBehavior(source events.EntityContext, e *events.EffectMutateEntityBehavior) {
	if e == nil {
		return
	}
	b, ok := s.entities.behaviors[e.EntityId]
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior to mutate")
		return
	}
	if e.DisableBy != nil {
		b.Disable(*e.DisableBy)
	}
	if e.EnableBy != nil {
		b.Enable(*e.EnableBy)
	}
	if e.Reset != nil && *e.Reset {
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
	p, ok := s.entities.positions[e.EntityId]
	if !ok {
		logEffectWarnf(source, e, "failed to find position for entity")
		return
	}
	p.CancelMovement()
	b, ok := s.entities.behaviors[e.EntityId]
	if ok {
		b.Reset()
	}
}

func (s *State) processEffectTriggerMovement(source events.EntityContext, e *events.EffectTriggerMovement) {
	if e == nil {
		return
	}
	entity := s.entities.GetEntity(e.EntityId)
	moveState := types.MoveStateWalking
	if e.MoveState != nil {
		moveState = *e.MoveState
	}
	var location MapLocation
	if e.Direction != nil {
		location = entity.GetPrimaryLocation().Moved(*e.Direction)
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

	// Build the teleport plan based on transition style
	var plan *events.EffectPlan

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
		plan = events.NewSerialPlan(effects...)
	case "instant":
		// Instant teleport with no transition
		plan = events.NewSerialPlan(&events.EffectSetEntityLocation{
			EntityId:    s.player,
			ToReference: e.ToReference,
			ToLocation:  e.ToLocation,
			ToEntityId:  e.ToEntityId,
		})
	default:
		logEffectWarnf(source, e, "unknown transition style, using fade")
		return
	}

	// Execute the plan as a system effect
	s.ExecuteSystemEffects(plan)
	logEffectInfof(source, e, "player teleport plan started with style: %s", transitionStyle)
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
	target := s.entities.GetEntity(*e.Follow.EntityId)
	location := s.camera.CurrentLocation()
	if e.Follow.ResetPosition {
		location = target.PreciseLocation()
	}
	camera := NewFollowCamera(target.Id, location, EntityCameraSpeedPlayerDefault)
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
		followCamera.location = s.entities.GetEntity(followCamera.target).PreciseLocation()
	}
	logEffectInfof(source, e, "follow camera mutated")
}

func (s *State) processEffectMutateNPC(source events.EntityContext, e *events.EffectMutateNPC) {
	if e == nil {
		return
	}
	entity := s.entities.GetEntity(e.EntityId)
	npc, ok := entity.EntityBehavior.(*NPCBehavior)
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
	p, ok := s.entities.positions[e.EntityId]
	if !ok {
		logEffectWarnf(source, e, "failed to find position for entity")
	}
	p.FacingDirection = e.Direction
}

func (s *State) processEffectStartScriptedMotion(source events.EntityContext, e *events.EffectStartScriptedMotion) {
	if e == nil {
		return
	}
	behvaior, ok := s.entities.behaviors[e.EntityId]
	if !ok {
		logEffectWarnf(source, e, "failed to find behave motion")
		return
	}
	override, ok := behvaior.(*EntityBehaviorOverride)
	if ok {
		behvaior = override.newBehavior
	}
	scripted, ok := behvaior.(*ScriptedMotionBehavior)
	if !ok {
		logEffectWarnf(source, e, "behavior is not a scripted motion behavior")
		return
	}
	var target MotionTarget
	if e.Location != nil {
		target = NewPathfindingMotion(e.MotionId, s.entities, e.EntityId, LocationFromEvent(*e.Location))
	} else if e.Relative != nil {
		target = NewRelativeMotion(e.MotionId, e.Relative.Direction, e.Relative.Steps)
	} else if e.ToEntityId != nil {
		toPosition, ok := s.entities.positions[*e.ToEntityId]
		if !ok {
			logEffectWarnf(source, e, "failed to find position for target entity")
			return
		}
		target = NewPathfindingMotion(e.MotionId, s.entities, e.EntityId, toPosition.GetPrimaryLocation())
	}
	scripted.SetTarget(target)
	logEffectInfof(source, e, "scripted motion started")
}

func (s *State) processEffectOverrideEntityBehavior(source events.EntityContext, e *events.EffectOverrideEntityBehavior) {
	if e == nil {
		return
	}
	if e.ScriptedMotion == nil {
		logEffectWarnf(source, e, "only scripted motion behaviors can be used to override")
		return
	}
	s.entities.OverrideBehavior(e.EntityId, NewScriptedMotionBehavior(e.EntityId, s.entities))
	logEffectInfof(source, e, "entity behavior overridden")
}

func (s *State) processEffectPopEntityBehaviorOverride(source events.EntityContext, e *events.EffectPopEntityBehaviorOverride) {
	if e == nil {
		return
	}
	s.entities.PopOverrideBehavior(e.EntityId)
	logEffectInfof(source, e, "entity behavior override popped")
}
