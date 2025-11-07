package adventure

import (
	"math/rand"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Effect is the interface that all effect types must implement
type Effect interface {
	FillDefaultsAndValidate() error
	Process(source EntityContext, s *State) bool
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

type RunnableFunction func()

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	Fn RunnableFunction
}

func (e *EffectFunction) Process(source EntityContext, s *State) bool {
	e.Fn()
	logEffectInfof(source, e, "function executed")
	return true
}

type EffectDialogue struct {
	DialogueId string `auto_generate:"true"`
	Text       string
}

func (e *EffectDialogue) Process(source EntityContext, s *State) bool {
	s.dialogues.Append(NewBasicDialogue(e.Text, e.DialogueId))
	logEffectInfof(source, e, "dialogue added")
	return true
}

type EffectChatter struct {
	ChatterId       string `auto_generate:"true"`
	EntityId        string
	DurationSeconds float64
	Message         string
}

func (e *EffectChatter) Process(source EntityContext, s *State) bool {
	s.chatters.Add(newBasicEntityChatter(e.EntityId, e.DurationSeconds, e.Message, e.ChatterId))
	logEffectInfof(source, e, "chatter added")
	return true
}

type EffectYieldElythium struct {
	Amount int
}

func (e *EffectYieldElythium) Process(source EntityContext, s *State) bool {
	newValue := max(s.RunState().Get(rpg.RunStateKeyElythium).AsInt(0)+e.Amount, 0)
	s.runState.Set(rpg.RunStateKeyElythium, newValue)
	logEffectInfof(source, e, "elythium updated to: %d", newValue)
	return true
}

type EffectTimer struct {
	TimerId         string `auto_generate:"true"`
	DurationSeconds float64
}

func (e *EffectTimer) Process(source EntityContext, s *State) bool {
	s.timers.AddTimer(source.EntityId(), e.TimerId, e.DurationSeconds)
	logEffectInfof(source, e, "timer added")
	return true
}

type EffectMutateModeBasedEntity struct {
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

type EffectSetWorldState struct {
	Key   string
	Value any
}

func (e *EffectSetWorldState) Process(source EntityContext, s *State) bool {
	s.worldState.Set(e.Key, e.Value)
	logEffectInfof(source, e, "world state updated")
	return true
}

type EffectSetRunState struct {
	Key   string
	Value any
}

func (e *EffectSetRunState) Process(source EntityContext, s *State) bool {
	s.runState.Set(e.Key, e.Value)
	logEffectInfof(source, e, "run state updated")
	return true
}

type EffectSetEntityLocation struct {
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

type EffectBatch struct {
	BatchId           string `auto_generate:"true"`
	Effects           []Effect
	ExecuteInParallel *bool // defaults to false (serial)
}

func (e *EffectBatch) Process(source EntityContext, s *State) bool {
	logEffectInfof(source, e, "batch starting")
	s.planExecutor.StartPlan(source, e)
	return true
}

func NewSerialPlan(effects ...Effect) *EffectBatch {
	return &EffectBatch{
		Effects: effects,
	}
}

func NewParallelPlan(effects ...Effect) *EffectBatch {
	parallel := true
	return &EffectBatch{
		Effects:           effects,
		ExecuteInParallel: &parallel,
	}
}

func (b *EffectBatch) WithId(id string) *EffectBatch {
	b.BatchId = id
	return b
}

func (b *EffectBatch) IsParallel() bool {
	return b.ExecuteInParallel != nil && *b.ExecuteInParallel
}

type EffectMutateEntityBehavior struct {
	EntityId  string
	DisableBy *string `one_of:"enablement"`
	EnableBy  *string `one_of:"enablement"`
	Reset     *bool
}

func (e *EffectMutateEntityBehavior) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to mutate")
		return false
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
			return false
		}
		b.Reset()
	}
	logEffectInfof(source, e, "behavior mutated")
	return true
}

type EffectFade struct {
	FadeId          string `auto_generate:"true"`
	DurationSeconds float64
	Transitions     int // number of transitions (1 = fade out, 2 = fade out+in, etc.)
	AutoDeactivate  *bool
	FromColor       *string
	ToColor         *string
}

func (e *EffectFade) WithColors(from, to string) *EffectFade {
	e.FromColor = &from
	e.ToColor = &to
	return e
}

func (e *EffectFade) Process(source EntityContext, s *State) bool {

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
	return true
}

type EffectDeactivateFade struct {
	FadeId string
}

func (e *EffectDeactivateFade) Process(source EntityContext, s *State) bool {
	s.overlays.Deactivate(e.FadeId)
	logEffectInfof(source, e, "fade deactivated")
	return true
}

type EffectTriggerMovement struct {
	EntityId  string
	Direction *input.Direction `one_of:"to"`
	Location  *MapLocation     `one_of:"to"`
	MoveState *MoveState
}

func (e *EffectTriggerMovement) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to move")
		return false
	}
	moveState := MoveStateWalking
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
	return true
}

type EffectResetMovement struct {
	EntityId string
}

func (e *EffectResetMovement) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return false
	}
	entity.CancelMovement()
	b, ok := entity.GetBehavior()
	if ok {
		b.Reset()
	}
	return true
}

type EffectOverrideCamera struct {
	Follow *FollowCamera `one_of:"type"`
}

func (e *EffectOverrideCamera) Process(source EntityContext, s *State) bool {
	if e.Follow == nil {
		logEffectWarnf(source, e, "follow camera override requires a follow camera")
		return false
	}
	if e.Follow.EntityId == nil {
		logEffectWarnf(source, e, "currently Id is required to set follow camera")
	}
	target, ok := s.entities.GetEntity(*e.Follow.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to follow")
		return false
	}
	location := s.camera.CurrentLocation()
	if e.Follow.ResetPosition {
		location = target.GetPreciseLocation()
	}
	camera := NewFollowCamera(target.GetId(), location, EntityCameraSpeedMedium)
	s.OverrideCamera(camera)
	logEffectInfof(source, e, "camera overriden")
	return true
}

type EffectPopCameraOverride struct {
	MaintainCurrentLocation bool
}

func (e *EffectPopCameraOverride) Process(source EntityContext, s *State) bool {
	s.PopOverrideCamera(e.MaintainCurrentLocation)
	logEffectInfof(source, e, "camera popped")
	return true
}

type EffectMutateFollowCamera struct {
	FollowEntityId *string
	ResetPosition  *bool
}

func (e *EffectMutateFollowCamera) Process(source EntityContext, s *State) bool {
	camera := s.camera
	if override, ok := camera.(*CameraOverride); ok {
		camera = override.newCamera
	}
	followCamera, ok := camera.(*EntityCamera)
	if !ok {
		logEffectWarnf(source, e, "camera is not a follow camera")
		return false
	}
	if e.FollowEntityId != nil {
		followCamera.target = *e.FollowEntityId
	}
	if e.ResetPosition != nil && *e.ResetPosition {
		followCamera.ResetPosition(s)
	}
	logEffectInfof(source, e, "follow camera mutated")
	return true
}

type FollowCamera struct {
	EntityId      *string `one_of:"target"`
	ResetPosition bool
}

type EffectMutateNPC struct {
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

type EffectTriggerCombat struct {
	CombatId   string `auto_generate:"true"`
	Opponent   *rpg.PrimortalType
	Background string
}

func (e *EffectTriggerCombat) Process(source EntityContext, s *State) bool {
	if s.enteringCombat {
		return false
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
	return true
}

type EffectEntityFaceDirection struct {
	EntityId     string
	Direction    *input.Direction `one_of:"dir"`
	TargetEntity *string          `one_of:"dir"`
}

func (e *EffectEntityFaceDirection) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find movement for entity")
	}
	if e.Direction != nil {
		entity.SetFacingDirection(*e.Direction)
	}
	if e.TargetEntity != nil {
		targetEntity, ok := s.entities.GetEntity(*e.TargetEntity)
		if !ok {
			logEffectWarnf(source, e, "failed to find target entity")
			return false
		}
		dir := entity.GetLocation().DirectionTowards(targetEntity.GetLocation())
		entity.SetFacingDirection(dir)
	}
	logEffectInfof(source, e, "entity facing direction set")
	return true
}

type EffectStartScriptedMotion struct {
	MotionId   string `auto_generate:"true"`
	EntityId   string
	Location   *MapLocation      `one_of:"target"`
	ToEntityId *string           `one_of:"target"`
	Relative   *RelativeLocation `one_of:"target"`
}

func (e *EffectStartScriptedMotion) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity")
		return false
	}
	behavior, ok := entity.GetBehavior()
	if !ok {
		logEffectWarnf(source, e, "failed to find behavior for entity")
		return false
	}
	scripted, ok := behavior.(*ScriptedMotionBehavior)
	if !ok {
		logEffectWarnf(source, e, "behavior is not a scripted motion behavior")
		return false
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
			return false
		}
		target = NewPathfindingMotion(e.MotionId, entity, toEntity.GetLocation())
	}
	scripted.SetTarget(target)
	logEffectInfof(source, e, "scripted motion started")
	return true
}

type RelativeLocation struct {
	Direction input.Direction
	Steps     int
}

type EffectPushEntityBehavior struct {
	EntityId       string
	ScriptedMotion *EntityBehaviorScriptedMotion `one_of:"type"`
	FacingEntity   *EntityBehaviorFacingEntity   `one_of:"type"`
}

func (e *EffectPushEntityBehavior) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to override behavior")
		return false
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
	return true
}

type EntityBehaviorPlayer struct {
}

type EntityBehaviorNPC struct {
}

type EntityBehaviorScriptedMotion struct {
}

type EntityBehaviorFacingEntity struct {
	EntityId *string
}

type EffectPopEntityBehavior struct {
	EntityId string
}

func (e *EffectPopEntityBehavior) Process(source EntityContext, s *State) bool {
	entity, ok := s.entities.GetEntity(e.EntityId)
	if !ok {
		logEffectWarnf(source, e, "failed to find entity to pop behavior")
		return false
	}
	entity.PopBehavior()
	logEffectInfof(source, e, "entity behavior override popped")
	return true
}

type EffectDeleteEntity struct {
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

type EffectWaitForCondition struct {
	ConditionId string `auto_generate:"true"`
	Check       ConditionCheck
}

func (e *EffectWaitForCondition) Process(source EntityContext, s *State) bool {
	s.conditions.AddCondition(e.ConditionId, e.Check)
	return true
}

type EffectLoadMap struct {
	MapName string
}

func (e *EffectLoadMap) Process(source EntityContext, s *State) bool {
	if e == nil {
		return false
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
	return true
}

type EffectSendEvent struct {
	Event any
}

func (e *EffectSendEvent) Process(source EntityContext, s *State) bool {
	if e == nil {
		return false
	}
	s.eventDispatcher.Dispatch(e.Event)
	logEffectInfof(source, e, "event sent")
	return true
}
