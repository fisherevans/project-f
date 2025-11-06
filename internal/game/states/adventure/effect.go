package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/game/states/adventure/types"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util"
)

// Effect is the interface that all effect types must implement
type Effect interface {
	FillDefaultsAndValidate() error
}

type RunnableFunction func()

func (RunnableFunction) String() string {
	return "<inline function>"
}

type EffectFunction struct {
	Fn RunnableFunction
}

type EffectDialogue struct {
	DialogueId string `auto_generate:"true"`
	Text       string
}

type EffectChatter struct {
	ChatterId       string `auto_generate:"true"`
	EntityId        string
	DurationSeconds float64
	Message         string
}

type EffectYieldElythium struct {
	Amount int
}

type EffectTimer struct {
	TimerId         string `auto_generate:"true"`
	DurationSeconds float64
}

type EffectMutateModeBasedEntity struct {
	EntityId string
	types.ModeBaseRenderConfig
}

func (e *EffectMutateModeBasedEntity) WithMode(mode string) *EffectMutateModeBasedEntity {
	e.Mode = &mode
	return e
}

func (e *EffectMutateModeBasedEntity) WithAnimations(anims map[string][]types.AnimationReference) *EffectMutateModeBasedEntity {
	e.Animations = anims
	return e
}

func (e *EffectMutateModeBasedEntity) WithLights(lights map[string][]types.LightConfig) *EffectMutateModeBasedEntity {
	e.Lights = lights
	return e
}

type EffectResetModeBasedEntityAnimation struct {
	EntityId string
}

type EffectMutateBlockingPresence struct {
	EntityId          string
	IsBlockingIngress *bool
}

type EffectSetWorldState struct {
	Key   string
	Value any
}

type EffectSetRunState struct {
	Key   string
	Value any
}

type EffectSetEntityLocation struct {
	EntityId    string
	ToReference *string      `one_of:"destination"`
	ToLocation  *MapLocation `one_of:"destination"`
	ToEntityId  *string      `one_of:"destination"`
}

type EffectTeleportPlayer struct {
	ToReference     *string      `one_of:"destination"`
	ToLocation      *MapLocation `one_of:"destination"`
	ToEntityId      *string      `one_of:"destination"`
	ExitDirection   *input.Direction
	TransitionStyle *string
}

// Batch execution: serial and parallel effect coordination
type EffectBatch struct {
	BatchId           string `auto_generate:"true"`
	Effects           []Effect
	ExecuteInParallel *bool // defaults to false (serial)
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

type EffectFade struct {
	FadeId          string `auto_generate:"true"`
	DurationSeconds float64
	Transitions     int // number of transitions (1 = fade out, 2 = fade out+in, etc.)
	AutoDeactivate  *bool
	FromColor       *string
	ToColor         *string
}

type EffectDeactivateFade struct {
	FadeId string
}

type EffectTriggerMovement struct {
	EntityId  string
	Direction *input.Direction `one_of:"to"`
	Location  *MapLocation     `one_of:"to"`
	MoveState *types.MoveState
}

type EffectResetMovement struct {
	EntityId string
}

type EffectOverrideCamera struct {
	Follow *FollowCamera `one_of:"type"`
}

type EffectPopCameraOverride struct {
	MaintainCurrentLocation bool
}

type EffectMutateFollowCamera struct {
	FollowEntityId *string
	ResetPosition  *bool
}

type FollowCamera struct {
	EntityId      *string `one_of:"target"`
	ResetPosition bool
}

type EffectMutateNPC struct {
	EntityId          string
	TalkingAtEntityId *string
}

type EffectTriggerCombat struct {
	CombatId   string `auto_generate:"true"`
	Opponent   *rpg.PrimortalType
	Background string
}

type EffectEntityFaceDirection struct {
	EntityId  string
	Direction input.Direction
}

type EffectStartScriptedMotion struct {
	MotionId   string `auto_generate:"true"`
	EntityId   string
	Location   *MapLocation      `one_of:"target"`
	ToEntityId *string           `one_of:"target"`
	Relative   *RelativeLocation `one_of:"target"`
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

type EffectDeleteEntity struct {
	EntityId string
}

type EffectRegisterEntity struct {
	EntityId string `auto_generate:"true"`

	Class      *string
	SpriteId   *resources.TilesheetSpriteId
	Properties **util.Properties

	MapLocation    *MapLocation `one_of:"location"`
	EntityLocation *string      `one_of:"location"`
}

type EffectWaitForCondition struct {
	ConditionId string `auto_generate:"true"`
	Check       ConditionCheck
}

type EffectLoadMap struct {
	MapName string
}

type EffectSendEvent struct {
	Event any
}
