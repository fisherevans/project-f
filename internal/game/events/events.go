package events

import (
	"fisherevans.com/project/f/internal/game/input"
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
	NewValue any
	OldValue any
}

type EventCombatComplete struct {
	CombatId string
	Result   string
}
