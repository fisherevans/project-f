package scripting

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// EntityScript represents a script attached to an entity
type EntityScript struct {
	EntityID   string
	ModuleName string
	State      *ScriptState
	Executor   *ScriptExecutor
	Module     *ScriptModule
}

// EventDispatcher manages event routing and handler execution
type EventDispatcher struct {
	mu            sync.RWMutex
	registry      *ScriptRegistry
	worldVars     *WorldVarStore
	applier       *EffectApplier
	planRunner    *PlanRunner
	entityScripts map[string]*EntityScript // entityID -> script
	stateStore    *StateStore
	trace         *DispatchTrace
}

// NewEventDispatcher creates a new event dispatcher
func NewEventDispatcher(registry *ScriptRegistry, worldVars *WorldVarStore) *EventDispatcher {
	applier := NewEffectApplier(worldVars)
	return &EventDispatcher{
		registry:      registry,
		worldVars:     worldVars,
		applier:       applier,
		planRunner:    NewPlanRunner(applier),
		entityScripts: make(map[string]*EntityScript),
		stateStore:    NewStateStore(),
		trace:         NewDispatchTrace(),
	}
}

// AttachScript attaches a script module to an entity
func (d *EventDispatcher) AttachScript(entityID, moduleName string, initialState map[string]interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get module
	module, ok := d.registry.Get(moduleName)
	if !ok {
		return fmt.Errorf("module %s not found", moduleName)
	}

	// Create executor
	executor, err := d.registry.CreateExecutor(module, d.worldVars)
	if err != nil {
		return err
	}

	// Create state
	state := &ScriptState{
		Module:  moduleName,
		Version: 1,
		Data:    initialState,
	}

	if initialState == nil {
		state.Data = make(map[string]interface{})
	}

	// Store entity script
	d.entityScripts[entityID] = &EntityScript{
		EntityID:   entityID,
		ModuleName: moduleName,
		State:      state,
		Executor:   executor,
		Module:     module,
	}

	// Store state
	d.stateStore.Set(entityID, state)

	return nil
}

// DetachScript removes a script from an entity
func (d *EventDispatcher) DetachScript(entityID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.entityScripts, entityID)
	d.stateStore.Delete(entityID)
}

// Dispatch processes an event through the reduce + commit pipeline
func (d *EventDispatcher) Dispatch(event Event) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Start trace
	d.trace.StartEvent(event)

	// REDUCE PHASE: Collect all handler outputs
	handlers := d.findHandlers(event)
	
	// Sort handlers by priority (DESC) and entityID (ASC)
	sort.Slice(handlers, func(i, j int) bool {
		if handlers[i].Module.Subscriptions.Priority != handlers[j].Module.Subscriptions.Priority {
			return handlers[i].Module.Subscriptions.Priority > handlers[j].Module.Subscriptions.Priority
		}
		return handlers[i].EntityID < handlers[j].EntityID
	})

	allEffects := make([]Effect, 0)
	allPlans := make(map[string]*Plan)

	for _, handler := range handlers {
		// Prepare self object
		self := map[string]interface{}{
			"id": handler.EntityID,
		}

		// Call handler
		result, err := handler.Executor.CallHandler(
			event.Type(),
			self,
			event,
			handler.State.Data,
		)

		if err != nil {
			d.trace.RecordError(handler.EntityID, err)
			continue
		}

		// Record trace
		d.trace.RecordHandler(handler.EntityID, result)

		// Collect effects
		if result.Effects != nil {
			for _, effect := range result.Effects {
				if effect.EntityID == "" {
					effect.EntityID = handler.EntityID
				}
				allEffects = append(allEffects, effect)
			}
		}

		// Collect plan
		if result.Plan != nil {
			allPlans[handler.EntityID] = result.Plan
		}

		// Update state
		if result.State != nil {
			handler.State.Data = result.State
			d.stateStore.Set(handler.EntityID, handler.State)
		}
	}

	// COMMIT PHASE: Validate and apply effects
	if err := d.applier.ApplyEffects(allEffects); err != nil {
		d.trace.RecordCommitError(err)
		return fmt.Errorf("failed to apply effects: %w", err)
	}

	// Record conflicts
	d.trace.RecordConflicts(d.applier.GetConflicts())

	// Start plans
	for entityID, plan := range allPlans {
		planID := d.planRunner.StartPlan(entityID, plan)
		d.trace.RecordPlan(entityID, planID)
	}

	d.trace.EndEvent()

	return nil
}

// Update advances plan runners and returns any generated effects
func (d *EventDispatcher) Update(deltaTime float64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// This would need the root plan structure - simplified for now
	// In practice, you'd store plans with their cursors
	return nil
}

// findHandlers finds all entity scripts that should handle an event
func (d *EventDispatcher) findHandlers(event Event) []*EntityScript {
	handlers := make([]*EntityScript, 0)

	for _, script := range d.entityScripts {
		if d.shouldHandle(script, event) {
			handlers = append(handlers, script)
		}
	}

	return handlers
}

func (d *EventDispatcher) shouldHandle(script *EntityScript, event Event) bool {
	// Check if handler exists
	if !script.Module.Handlers[event.Type()] {
		return false
	}

	// Check entity-specific events
	if event.TargetEntityID() != "" && event.TargetEntityID() != script.EntityID {
		return false
	}

	// Check subscription filters
	sub := script.Module.Subscriptions

	switch event.Type() {
	case EventTypeEnterZone:
		if len(sub.EnterZone) > 0 {
			ev := event.(*EnterZoneEvent)
			return d.containsString(sub.EnterZone, ev.ZoneName)
		}

	case EventTypeLeaveZone:
		if len(sub.LeaveZone) > 0 {
			ev := event.(*LeaveZoneEvent)
			return d.containsString(sub.LeaveZone, ev.ZoneName)
		}

	case EventTypeTrigger:
		if len(sub.Trigger) > 0 {
			ev := event.(*TriggerEvent)
			return d.containsString(sub.Trigger, ev.TriggerName)
		}

	case EventTypeFlagChanged:
		if len(sub.VarPrefix) > 0 {
			ev := event.(*FlagChangedEvent)
			for _, prefix := range sub.VarPrefix {
				if d.matchesPrefix(ev.VarName, prefix) {
					return true
				}
			}
			return false
		}
	}

	return true
}

func (d *EventDispatcher) containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (d *EventDispatcher) matchesPrefix(key, pattern string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(key, prefix)
	}
	return key == pattern
}

// GetEntityState returns the current state for an entity
func (d *EventDispatcher) GetEntityState(entityID string) (*ScriptState, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.stateStore.Get(entityID)
}

// GetTrace returns the dispatch trace for debugging
func (d *EventDispatcher) GetTrace() *DispatchTrace {
	return d.trace
}

// GetPlanRunner returns the plan runner
func (d *EventDispatcher) GetPlanRunner() *PlanRunner {
	return d.planRunner
}

// DispatchTrace records execution details for debugging
type DispatchTrace struct {
	mu              sync.RWMutex
	currentEvent    Event
	handlerResults  map[string]*HandlerResult
	errors          map[string]error
	conflicts       []ConflictReport
	plans           map[string]string // entityID -> planID
	commitError     error
	eventStartTime  int64
	eventEndTime    int64
}

// NewDispatchTrace creates a new dispatch trace
func NewDispatchTrace() *DispatchTrace {
	return &DispatchTrace{
		handlerResults: make(map[string]*HandlerResult),
		errors:         make(map[string]error),
		plans:          make(map[string]string),
	}
}

func (t *DispatchTrace) StartEvent(event Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentEvent = event
	t.handlerResults = make(map[string]*HandlerResult)
	t.errors = make(map[string]error)
	t.conflicts = nil
	t.plans = make(map[string]string)
	t.commitError = nil
	t.eventStartTime = event.Timestamp().UnixNano()
}

func (t *DispatchTrace) RecordHandler(entityID string, result *HandlerResult) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.handlerResults[entityID] = result
}

func (t *DispatchTrace) RecordError(entityID string, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.errors[entityID] = err
}

func (t *DispatchTrace) RecordConflicts(conflicts []ConflictReport) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.conflicts = conflicts
}

func (t *DispatchTrace) RecordPlan(entityID, planID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.plans[entityID] = planID
}

func (t *DispatchTrace) RecordCommitError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.commitError = err
}

func (t *DispatchTrace) EndEvent() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.eventEndTime = t.currentEvent.Timestamp().UnixNano()
}

// GetSummary returns a summary of the trace
func (t *DispatchTrace) GetSummary() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"event":          t.currentEvent.Type(),
		"handlerCount":   len(t.handlerResults),
		"errorCount":     len(t.errors),
		"conflictCount":  len(t.conflicts),
		"planCount":      len(t.plans),
		"hasCommitError": t.commitError != nil,
		"durationNs":     t.eventEndTime - t.eventStartTime,
	}
}
