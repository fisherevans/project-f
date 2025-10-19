package scripting

import (
	"encoding/json"
	"time"
)

// EventType represents the type of event being dispatched
type EventType string

const (
	EventTypeInit        EventType = "Init"
	EventTypeInteract    EventType = "Interact"
	EventTypeEnterZone   EventType = "EnterZone"
	EventTypeLeaveZone   EventType = "LeaveZone"
	EventTypeTimer       EventType = "Timer"
	EventTypeTrigger     EventType = "Trigger"
	EventTypeFlagChanged EventType = "FlagChanged"
	EventTypeStep        EventType = "Step"
)

// Event represents a game event that can be dispatched to entity scripts
type Event interface {
	Type() EventType
	TargetEntityID() string
	Timestamp() time.Time
	ToMap() map[string]interface{}
}

// BaseEvent provides common event fields
type BaseEvent struct {
	EventType      EventType `json:"type"`
	EntityID       string    `json:"entityId"`
	EventTimestamp time.Time `json:"timestamp"`
}

func (e *BaseEvent) Type() EventType {
	return e.EventType
}

func (e *BaseEvent) TargetEntityID() string {
	return e.EntityID
}

func (e *BaseEvent) Timestamp() time.Time {
	return e.EventTimestamp
}

func (e *BaseEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":      string(e.EventType),
		"entityId":  e.EntityID,
		"timestamp": e.EventTimestamp.Unix(),
	}
}

// InitEvent is fired when an entity is first created
type InitEvent struct {
	BaseEvent
}

func NewInitEvent(entityID string) *InitEvent {
	return &InitEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeInit,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
	}
}

// InteractEvent is fired when an entity is interacted with
type InteractEvent struct {
	BaseEvent
	SourceEntityID string `json:"sourceEntityId"`
}

func NewInteractEvent(entityID, sourceEntityID string) *InteractEvent {
	return &InteractEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeInteract,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		SourceEntityID: sourceEntityID,
	}
}

func (e *InteractEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["sourceEntityId"] = e.SourceEntityID
	return m
}

// EnterZoneEvent is fired when an entity enters a zone
type EnterZoneEvent struct {
	BaseEvent
	ZoneName string `json:"zoneName"`
}

func NewEnterZoneEvent(entityID, zoneName string) *EnterZoneEvent {
	return &EnterZoneEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeEnterZone,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		ZoneName: zoneName,
	}
}

func (e *EnterZoneEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["zoneName"] = e.ZoneName
	return m
}

// LeaveZoneEvent is fired when an entity leaves a zone
type LeaveZoneEvent struct {
	BaseEvent
	ZoneName string `json:"zoneName"`
}

func NewLeaveZoneEvent(entityID, zoneName string) *LeaveZoneEvent {
	return &LeaveZoneEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeLeaveZone,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		ZoneName: zoneName,
	}
}

func (e *LeaveZoneEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["zoneName"] = e.ZoneName
	return m
}

// TimerEvent is fired when a timer expires
type TimerEvent struct {
	BaseEvent
	TimerName string `json:"timerName"`
}

func NewTimerEvent(entityID, timerName string) *TimerEvent {
	return &TimerEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeTimer,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		TimerName: timerName,
	}
}

func (e *TimerEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["timerName"] = e.TimerName
	return m
}

// TriggerEvent is fired when a named trigger is activated
type TriggerEvent struct {
	BaseEvent
	TriggerName string                 `json:"triggerName"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

func NewTriggerEvent(entityID, triggerName string, data map[string]interface{}) *TriggerEvent {
	return &TriggerEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeTrigger,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		TriggerName: triggerName,
		Data:        data,
	}
}

func (e *TriggerEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["triggerName"] = e.TriggerName
	if e.Data != nil {
		m["data"] = e.Data
	}
	return m
}

// FlagChangedEvent is fired when a world variable changes
type FlagChangedEvent struct {
	BaseEvent
	VarName  string      `json:"varName"`
	OldValue interface{} `json:"oldValue"`
	NewValue interface{} `json:"newValue"`
}

func NewFlagChangedEvent(entityID, varName string, oldValue, newValue interface{}) *FlagChangedEvent {
	return &FlagChangedEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeFlagChanged,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		VarName:  varName,
		OldValue: oldValue,
		NewValue: newValue,
	}
}

func (e *FlagChangedEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["varName"] = e.VarName
	m["oldValue"] = e.OldValue
	m["newValue"] = e.NewValue
	return m
}

// StepEvent is fired periodically (low frequency) for background processing
type StepEvent struct {
	BaseEvent
	DeltaTime float64 `json:"deltaTime"`
}

func NewStepEvent(entityID string, deltaTime float64) *StepEvent {
	return &StepEvent{
		BaseEvent: BaseEvent{
			EventType:      EventTypeStep,
			EntityID:       entityID,
			EventTimestamp: time.Now(),
		},
		DeltaTime: deltaTime,
	}
}

func (e *StepEvent) ToMap() map[string]interface{} {
	m := e.BaseEvent.ToMap()
	m["deltaTime"] = e.DeltaTime
	return m
}

// HandlerResult represents the output from a script event handler
type HandlerResult struct {
	Effects []Effect               `json:"effects,omitempty"`
	Plan    *Plan                  `json:"plan,omitempty"`
	State   map[string]interface{} `json:"state,omitempty"`
}

// ParseHandlerResult converts a Goja value to a HandlerResult
func ParseHandlerResult(data interface{}) (*HandlerResult, error) {
	if data == nil {
		return &HandlerResult{}, nil
	}

	// Convert to JSON and back to ensure proper typing
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result HandlerResult
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
