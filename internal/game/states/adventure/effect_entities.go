package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
)

type EffectDeleteEntity struct {
	instantEffect
	EntityId string
}

func (e *EffectDeleteEntity) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to delete")
		return false
	}
	s.eventDispatcher.Unregister(entity.GetEntityContext())
	s.entities.DeleteEntity(e.EntityId)
	logEffectInfof(source, e, "entity deleted")
	return true
}

type EffectRegisterEntity struct {
	instantEffect
	EntityId string `auto_generate:"true"`

	Class      *string
	SpriteId   *resources.TilesheetSpriteId
	Properties **util.Properties

	MapLocation    *MapLocation `one_of:"location"`
	EntityLocation *string      `one_of:"location"`
}

func (e *EffectRegisterEntity) Process(source EntityContext, s *State) bool {
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
			return false
		}
		params.Location = toEntity.GetLocation()
	}
	if !s.registerParameterizedEntity(params) {
		logEffectWarnf(source, e, "failed to register entity")
		return false
	}
	logEffectInfof(source, e, "entity registered: %v", params)
	return true
}

type EffectMutateModeBasedEntity struct {
	instantEffect
	EntityId string
	ModeBaseRenderConfig
}

func (e *EffectMutateModeBasedEntity) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return false
	}
	renderer, ok := entity.GetRenderer()
	if !ok {
		logEffectWarnf(source, e, "failed to find renderer")
		return false
	}
	modeBased, ok := renderer.(*ModeBasedEntityRenderer)
	if !ok {
		logEffectWarnf(source, e, "failed to find mode based entity for mutation")
		return false
	}
	modeBased.WithConfig(&e.ModeBaseRenderConfig)
	logEffectInfof(source, e, "mode based entity mutated")
	return true
}

func (e *EffectMutateModeBasedEntity) WithMode(mode string) *EffectMutateModeBasedEntity {
	e.Mode = &mode
	return e
}

func (e *EffectMutateModeBasedEntity) WithAnimations(anims map[string][]AnimationReference) *EffectMutateModeBasedEntity {
	e.Animations = anims
	return e
}

func (e *EffectMutateModeBasedEntity) WithLights(lights map[string][]LightConfig) *EffectMutateModeBasedEntity {
	e.Lights = lights
	return e
}

type EffectResetModeBasedEntityAnimation struct {
	instantEffect
	EntityId string
}

func (e *EffectResetModeBasedEntityAnimation) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return false
	}
	renderer, ok := entity.GetRenderer()
	if !ok {
		logEffectWarnf(source, e, "failed to find renderer")
		return false
	}
	modeBased, ok := renderer.(*ModeBasedEntityRenderer)
	if !ok {
		logEffectWarnf(source, e, "failed to find mode based entity for mutation")
		return false
	}
	for _, animation := range modeBased.getBasicEntityRenderer(modeBased.currentMode).animations {
		animation.Animation.Reset()
	}
	logEffectInfof(source, e, "mode based entity animations reset")
	return true
}

type EffectMutateBlockingPresence struct {
	instantEffect
	EntityId          string
	IsBlockingIngress *bool
}

func (e *EffectMutateBlockingPresence) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return false
	}
	presence, ok := entity.GetPresence()
	if !ok {
		logEffectWarnf(source, e, "failed to find presence")
		return false
	}
	blockingPresence, ok := presence.(*BlockIngressPresence)
	if !ok {
		logEffectWarnf(source, e, "failed to find presence block ingress")
		return false
	}
	if e.IsBlockingIngress != nil {
		blockingPresence.isBlockingIngress = *e.IsBlockingIngress
	}
	logEffectInfof(source, e, "block presence mutated")
	return true
}

type EffectSetEntityLocation struct {
	instantEffect
	EntityId    string
	ToReference *string      `one_of:"destination"`
	ToLocation  *MapLocation `one_of:"destination"`
	ToEntityId  *string      `one_of:"destination"`
}

func (e *EffectSetEntityLocation) Process(source EntityContext, s *State) bool {
	var toLocation MapLocation
	if e.ToReference != nil {
		tele, ok := s.teleports[TeleportReference(*e.ToReference)]
		if !ok {
			logEffectWarnf(source, e, "failed to find teleport reference")
			return false
		}
		toLocation = tele.Location
	} else if e.ToLocation != nil {
		toLocation = MapLocation{X: e.ToLocation.X, Y: e.ToLocation.Y}
	} else if e.ToEntityId != nil {
		entity, ok := s.entities.GetEntity(*e.ToEntityId)
		if !ok {
			logEffectWarnf(source, e, "failed to find entity to teleport to")
			return false
		}
		toLocation = entity.GetLocation()
	} else {
		logEffectWarnf(source, e, "failed to find teleport destination")
		return false
	}
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to teleport")
		return false
	}
	entity.Teleport(toLocation)
	return true
}

type EffectTeleportPlayer struct {
	instantEffect
	ToReference     *string      `one_of:"destination"`
	ToLocation      *MapLocation `one_of:"destination"`
	ToEntityId      *string      `one_of:"destination"`
	ExitDirection   *input.Direction
	TransitionStyle *string
}

func (e *EffectTeleportPlayer) Process(source EntityContext, s *State) bool {

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
		// getEventHandler exit direction from teleport reference
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
			NewPlaySoundEffect("adventure/transition"),
		)

		if exitDirection != nil && *exitDirection != input.NotPressed {
			effects = append(effects,
				NewMutateEntityBehaviorEffect(s.player).WithReset(true),
				NewTriggerMovementEffect(s.player).WithDirection(*exitDirection))
		}

		effects = append(effects,
			NewParallelPlan(
				NewFadeEffect(fadeDuration, 1).
					WithAutoDeactivate(true).
					WithFromColor("#000000FF").
					WithToColor("#00000000"),
				NewDeactivateFadeEffect(fadeOutId),
			),
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
		return false
	}

	s.ExecuteSystemEffects(batch)
	logEffectInfof(source, e, "player teleport batch started with style: %s", transitionStyle)
	return true
}

type EffectMutateNPC struct {
	instantEffect
	EntityId          string
	TalkingAtEntityId *string
}

func (e *EffectMutateNPC) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity for mutation")
		return false
	}
	behavior, ok := entity.GetBehavior()
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior for mutation")
		return false
	}
	npc, ok := behavior.(*NPCBehavior)
	if !ok {
		logEffectWarnf(source, e, "failed to find npc entity for mutation")
		return false
	}
	if e.TalkingAtEntityId != nil {
		npc.talkingTowards = *e.TalkingAtEntityId
	}
	logEffectInfof(source, e, "npc mutated")
	return true
}
