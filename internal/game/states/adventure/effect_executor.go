package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/events"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog/log"
)

func (s *State) processEffects(effects []events.DispatchedEffect) {
	for _, e := range effects {
		s.processEffectFunction(e.Source, e.Function)
		s.processEffectDialogue(e.Source, e.Dialogue)
		s.processEffectTimer(e.Source, e.Timer)
		s.processEffectMutateEntity(e.Source, e.MutateEntity)
		s.processEffectYieldElythium(e.Source, e.YieldElythium)
		s.processEffectChatter(e.Source, e.Chatter)
		s.processSetWorldState(e.Source, e.SetWorldState)
		s.processTeleportEntity(e.Source, e.SetEntityLocation)
		s.processEffectPlan(e.Source, e.Plan)
		s.processEffectBlockInput(e.Source, e.BlockInput)
		s.processEffectFade(e.Source, e.Fade)
		s.processEffectDeactivateFade(e.Source, e.DeactivateFade)
		s.processEffectTeleportPlayer(e.Source, e.TeleportPlayer)
		s.processEffectTriggerMovement(e.Source, e.TriggerMovement)
		s.processEffectSetFollowCamera(e.Source, e.SetFollowCamera)
		s.processEffectMutateNPC(e.Source, e.MutateNPC)
		s.processEffectTriggerCombat(e.Source, e.TriggerCombat)
	}
}

func logEffectf(source events.EntityContext, e any, messageFormat string, args ...any) {
	log.Info().Str("caller", source.Id()).Interface("e", e).Msgf(messageFormat, args...)
}
func (s *State) processEffectFunction(source events.EntityContext, e *events.EffectFunction) {
	if e == nil {
		return
	}
	e.Fn()
	logEffectf(source, e, "function executed")
}
func (s *State) processEffectDialogue(source events.EntityContext, e *events.EffectDialogue) {
	if e == nil {
		return
	}
	s.dialogues.Append(NewBasicDialogue(e.Text, e.DialogueId))
	logEffectf(source, e, "dialogue added")
}

func (s *State) processEffectTimer(source events.EntityContext, e *events.EffectTimer) {
	if e == nil {
		return
	}
	s.timers.AddTimer(source.Id(), e.TimerId, e.DurationSeconds)
	logEffectf(source, e, "timer added")
}

func (s *State) processEffectMutateEntity(source events.EntityContext, e *events.EffectMutateEntity) {
	if e == nil {
		return
	}
	entity, exists := s.entities[EntityId(e.EntityId)]
	if !exists {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to find entity for mutation")
		return
	}
	if e.Mode != nil {
		entity.SetMode(*e.Mode)
	}
	dynamic, isDynamic := entity.(*DynamicEntity)
	if isDynamic {
		if e.IsPassable != nil {
			dynamic.SetDefaultIsPassable(*e.IsPassable)
		}
		if e.DynamicAnimations != nil && *e.DynamicAnimations != nil {
			dynamic.SetDynamicAnimations(*e.DynamicAnimations)
		}
		if e.DynamicLights != nil && *e.DynamicLights != nil {
			dynamic.SetDynamicLights(*e.DynamicLights)
		}
	}
	logEffectf(source, e, "entity mutated")
}

func (s *State) processEffectYieldElythium(source events.EntityContext, e *events.EffectYieldElythium) {
	if e == nil {
		return
	}
	s.run.Elythium += e.Amount
	logEffectf(source, e, "elythium granted")
}

func (s *State) processEffectChatter(source events.EntityContext, e *events.EffectChatter) {
	if e == nil {
		return
	}
	s.chatters.Add(newBasicEntityChatter(EntityId(e.EntityId), e.DurationSeconds, e.Message, e.ChatterId))
	logEffectf(source, e, "chatter added")
}

func (s *State) processSetWorldState(source events.EntityContext, e *events.EffectSetWorldState) {
	if e == nil {
		return
	}
	s.setWorldState(e.Key, e.Value, source.Id())
	logEffectf(source, e, "world state updated")
}

func (s *State) processTeleportEntity(source events.EntityContext, e *events.EffectSetEntityLocation) {
	if e == nil {
		return
	}
	var toLocation MapLocation
	if e.ToReference != nil {
		tele, ok := s.teleports[TeleportReference(*e.ToReference)]
		if !ok {
			log.Warn().Str("toReference", *e.ToReference).Msg("failed to find teleport reference")
			return
		}
		toLocation = tele.Location
	} else if e.ToLocation != nil {
		toLocation = MapLocation{X: e.ToLocation.X, Y: e.ToLocation.Y}
	} else if e.ToEntityId != nil {
		toEntity, exists := s.entities[EntityId(*e.ToEntityId)]
		if !exists {
			log.Warn().Str("toEntityId", *e.ToEntityId).Msg("failed to find entity to teleport to")
			return
		}
		toLocation = toEntity.Location()
	} else {
		log.Warn().Msg("failed to find teleport destination")
		return
	}
	entity, exists := s.entities[EntityId(e.EntityId)]
	if !exists {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to find entity for teleport")
		return
	}
	if entity.TeleportTo(s, toLocation) {
		logEffectf(source, e, "entity teleported")
	} else {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to teleport entity")
	}
}

// processEffectPlan starts a new plan
func (s *State) processEffectPlan(source events.EntityContext, e *events.EffectPlan) {
	if e == nil {
		return
	}
	s.planExecutor.StartPlan(source, e)
	logEffectf(source, e, "plan started")
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectBlockInput(source events.EntityContext, e *events.EffectBlockInput) {
	if e == nil {
		return
	}
	s.blockInput = e.Blocked
	logEffectf(source, e, "input blocked: %v", e.Blocked)
}

// processEffectBlockInput blocks or unblocks input
func (s *State) processEffectDeactivateFade(source events.EntityContext, e *events.EffectDeactivateFade) {
	if e == nil {
		return
	}
	s.overlays.Deactivate(e.FadeId)
	logEffectf(source, e, "fade deactivated")
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
	logEffectf(source, e, "fade overlay added")
}

func (s *State) processEffectTriggerMovement(source events.EntityContext, e *events.EffectTriggerMovement) {
	if e == nil {
		return
	}
	if e.EntityId != string(s.player.GetEntityId()) {
		log.Warn().Str("entityId", e.EntityId).Msg("only player is supported for trigger movement")
		return
	}
	
	// Get player's movement state and trigger movement
	movement := s.movementController.Get(s.player.GetEntityId())
	if movement != nil {
		movement.SetFacingDirection(e.Direction)
		target := movement.GetLocationInDirection(e.Direction)
		s.movementController.TriggerMovement(s, s.player.GetEntityId(), target, MoveStateWalking)
	}
	
	// Update player behavior intent
	if behavior, ok := s.behaviors[s.player.GetEntityId()].(*PlayerBehavior); ok {
		behavior.intentDirection = e.Direction
	}
}

// processEffectTeleportPlayer creates a plan to teleport the player with fade transition
func (s *State) processEffectTeleportPlayer(source events.EntityContext, e *events.EffectTeleportPlayer) {
	if e == nil {
		return
	}

	// Helper to create string pointer
	strPtr := func(s string) *string { return &s }
	bPtr := func(b bool) *bool { return &b }

	// Determine transition style (default to fade)
	transitionStyle := "fade"
	if e.TransitionStyle != nil {
		transitionStyle = *e.TransitionStyle
	}

	playerEntityId := string(s.player.GetEntityId())

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

		steps := []events.PlanStep{
			// Step 1: Block input and fade out
			{Serial: []events.Effect{
				{BlockInput: &events.EffectBlockInput{Blocked: true}},
				{Fade: &events.EffectFade{
					FadeId:          fadeOutId,
					DurationSeconds: 0.33,
					FromColor:       strPtr("#00000000"), // transparent
					ToColor:         strPtr("#000000FF"), // black
					Transitions:     1,
					AutoDeactivate:  bPtr(false),
				}},
				{SetEntityLocation: &events.EffectSetEntityLocation{
					EntityId:    playerEntityId,
					ToReference: e.ToReference,
					ToLocation:  e.ToLocation,
					ToEntityId:  e.ToEntityId,
				}},
				{SetFollowCamera: &events.EffectSetFollowCamera{
					EntityId:      &playerEntityId,
					ResetPosition: true,
				}},
			}},
		}

		// todo consider adding "wait for idle" for player to stop moving

		// Set exit direction immediately after teleport (if specified)
		if exitDirection != nil && *exitDirection != input.NotPressed {
			steps = append(steps,
				events.PlanStep{Parallel: []events.Effect{
					{TriggerMovement: &events.EffectTriggerMovement{
						EntityId:  playerEntityId,
						Direction: *exitDirection,
					}},
				}},
			)
		}

		// Add fade in and unblock steps
		steps = append(steps,
			// Step 3/4: Fade in and unblock input (parallel)
			events.PlanStep{Parallel: []events.Effect{
				{Fade: &events.EffectFade{
					FadeId:          "", // Auto-generated
					DurationSeconds: 0.33,
					FromColor:       strPtr("#000000FF"), // black
					ToColor:         strPtr("#00000000"), // transparent
					Transitions:     1,
				}},
				{DeactivateFade: &events.EffectDeactivateFade{
					FadeId: fadeOutId,
				}},
				{Timer: &events.EffectTimer{
					TimerId:         "", // Auto-generated
					DurationSeconds: 0.33,
				}},
			}},
			events.PlanStep{Parallel: []events.Effect{
				{BlockInput: &events.EffectBlockInput{Blocked: false}},
			}},
		)

		plan = &events.EffectPlan{
			PlanId: "", // Auto-generated
			Steps:  steps,
		}
	case "instant":
		// Instant teleport with no transition
		plan = &events.EffectPlan{
			PlanId: "", // Auto-generated
			Steps: []events.PlanStep{
				{Serial: []events.Effect{
					{SetEntityLocation: &events.EffectSetEntityLocation{
						EntityId:    playerEntityId,
						ToReference: e.ToReference,
						ToLocation:  e.ToLocation,
						ToEntityId:  e.ToEntityId,
					}},
				}},
			},
		}

	default:
		log.Warn().Str("transitionStyle", transitionStyle).Msg("Unknown transition style, using fade")
		return
	}

	// Start the plan
	s.planExecutor.StartPlan(source, plan)
	logEffectf(source, e, "player teleport plan started with style: %s", transitionStyle)
}

func (s *State) processEffectSetFollowCamera(source events.EntityContext, e *events.EffectSetFollowCamera) {
	if e == nil {
		return
	}
	if e.EntityId == nil {
		log.Warn().Str("entityId", *e.EntityId).Msg("currently EntityId is required to set follow camera")
	}
	entity, exists := s.entities[EntityId(*e.EntityId)]
	if !exists {
		log.Warn().Str("entityId", *e.EntityId).Msg("failed to find entity for follow camera")
		return
	}
	location := s.camera.CurrentLocation()
	if e.ResetPosition {
		location = entity.RenderMapLocation()
	}
	camera := NewFollowCamera(EntityId(*e.EntityId), location, EntityCameraSpeedPlayerDefault)
	s.camera = camera
	logEffectf(source, e, "follow camera set")
}

func (s *State) processEffectMutateNPC(source events.EntityContext, e *events.EffectMutateNPC) {
	if e == nil {
		return
	}
	entity, exists := s.entities[EntityId(e.EntityId)]
	if !exists {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to find entity for mutation")
		return
	}
	npc, ok := entity.(*NPC)
	if !ok {
		log.Warn().Str("entityId", e.EntityId).Msg("failed to cast entity to NPC")
	}
	if e.TalkingAtEntityId != nil {
		npc.Talking = true
		npc.TalkingTowards = EntityId(*e.TalkingAtEntityId)
	}
	if e.IsTalking != nil {
		npc.Talking = *e.IsTalking
	}
	logEffectf(source, e, "npc mutated")
}

func (s *State) processEffectTriggerCombat(source events.EntityContext, e *events.EffectTriggerCombat) {
	if e == nil {
		return
	}

	if s.enteringCombat {
		return
	}
	s.enteringCombat = true
	s.blockInput = true

	postCombat := func(r game.CombatIntentResult) {
		game.DebugNotification("Combat complete!")
		if !r.PlayerWon {
			game.SetActiveStateIntent(game.InitialState())
			return
		}
		game.SetActiveStateIntent(game.SwapStateIntent{
			State: s,
		})
		game.CurrentSave().Animech.AnimechExperience += r.ResearchPoints // todo this isn't right
		s.AddSystemEffect(events.Effect{
			DeactivateFade: &events.EffectDeactivateFade{
				FadeId: "combat_fade",
			},
		})
		s.eventDispatcher.Dispatch(&events.EventCombatComplete{
			CombatId: e.CombatId,
			Result:   "completed", // TODO: serialize result properly
		})
		s.planExecutor.MarkCombatComplete(e.CombatId)
	}

	s.AddSystemEffect(events.Effect{
		Plan: &events.EffectPlan{
			Steps: []events.PlanStep{
				{
					Serial: []events.Effect{
						{
							Fade: &events.EffectFade{
								DurationSeconds: 1,
								AutoDeactivate:  ptr(true),
								FromColor:       ptr("#00000000"),
								ToColor:         ptr("#000000FF"),
								Transitions:     6,
							},
						},
						{
							Fade: &events.EffectFade{
								FadeId:          "combat_fade",
								DurationSeconds: 3,
								AutoDeactivate:  ptr(false),
								FromColor:       ptr("#00000000"),
								ToColor:         ptr("#000000FF"),
								Transitions:     1,
							},
							Function: &events.EffectFunction{
								Fn: func() {
									game.SetCustomShader(game.NewSwirlShader(3))
								},
							},
						},
						{
							Function: &events.EffectFunction{
								Fn: func() {
									s.enteringCombat = false
									s.blockInput = false
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
								},
							},
						},
					},
				},
			},
		},
	})
}

func ptr[T any](t T) *T {
	return &t
}
