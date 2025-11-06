package adventure

import (
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
)

// Event types - add new events here and run go generate

type EventOnInteract struct {
	SourceId              string
	TargetId              string
	SourceFacingDirection input.Direction
}

type EventDialogueComplete struct {
	DialogueId string
}

type EventChatterComplete struct {
	ChatterId string
	EntityId  string
}

type EventTimerComplete struct {
	CreatedBy       string
	TimerId         string
	DurationSeconds float64
	TriggerCount    int
}

type EventEntityZoneActivity struct {
	EntityId      string
	ZoneId        string
	IsEntering    bool
	WasTeleported bool
}

type EventWorldStateUpdated struct {
	Key      string
	NewValue *rpg.StateValue
	OldValue *rpg.StateValue
}

func NewEventWorldStateUpdated(key string, newValue, oldValue *rpg.StateValue) any {
	return EventWorldStateUpdated{
		Key:      key,
		NewValue: newValue,
		OldValue: oldValue,
	}
}

type EventWorldStateDeleted struct {
	Key      string
	OldValue *rpg.StateValue
}

func NewEventWorldStateDeleted(key string, oldValue *rpg.StateValue) any {
	return EventWorldStateDeleted{
		Key:      key,
		OldValue: oldValue,
	}
}

type EventRunStateUpdated struct {
	Key      string
	NewValue *rpg.StateValue
	OldValue *rpg.StateValue
}

func NewEventRunStateUpdated(key string, newValue, oldValue *rpg.StateValue) any {
	return EventRunStateUpdated{
		Key:      key,
		NewValue: newValue,
		OldValue: oldValue,
	}
}

type EventRunStateDeleted struct {
	Key      string
	OldValue *rpg.StateValue
}

func NewEventRunStateDeleted(key string, oldValue *rpg.StateValue) any {
	return EventRunStateDeleted{
		Key:      key,
		OldValue: oldValue,
	}
}

type EventCombatComplete struct {
	CombatId string
	Result   string
}

type EventScriptedMotionComplete struct {
	EntityId    string
	MotionId    string
	WasCanceled bool
}
