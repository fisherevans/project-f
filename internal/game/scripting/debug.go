package scripting

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// DebugTools provides debugging utilities for the scripting system
type DebugTools struct {
	dispatcher *EventDispatcher
	worldVars  *WorldVarStore
	eventLog   *EventLog
}

// NewDebugTools creates a new debug tools instance
func NewDebugTools(dispatcher *EventDispatcher, worldVars *WorldVarStore) *DebugTools {
	return &DebugTools{
		dispatcher: dispatcher,
		worldVars:  worldVars,
		eventLog:   NewEventLog(1000), // Keep last 1000 events
	}
}

// EventLog maintains a history of dispatched events
type EventLog struct {
	maxSize int
	entries []*EventLogEntry
}

// EventLogEntry represents a single event in the log
type EventLogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   EventType              `json:"eventType"`
	EntityID    string                 `json:"entityId"`
	EventData   map[string]interface{} `json:"eventData"`
	HandlerCount int                   `json:"handlerCount"`
	EffectCount int                    `json:"effectCount"`
	Conflicts   []ConflictReport       `json:"conflicts,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
	DurationUs  int64                  `json:"durationUs"`
}

// NewEventLog creates a new event log
func NewEventLog(maxSize int) *EventLog {
	return &EventLog{
		maxSize: maxSize,
		entries: make([]*EventLogEntry, 0, maxSize),
	}
}

// Add adds an entry to the log
func (l *EventLog) Add(entry *EventLogEntry) {
	l.entries = append(l.entries, entry)
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[1:]
	}
}

// GetRecent returns the most recent N entries
func (l *EventLog) GetRecent(n int) []*EventLogEntry {
	if n > len(l.entries) {
		n = len(l.entries)
	}
	start := len(l.entries) - n
	return l.entries[start:]
}

// GetAll returns all entries
func (l *EventLog) GetAll() []*EventLogEntry {
	return l.entries
}

// Clear clears the log
func (l *EventLog) Clear() {
	l.entries = make([]*EventLogEntry, 0, l.maxSize)
}

// LogEvent logs an event dispatch
func (d *DebugTools) LogEvent(event Event, trace *DispatchTrace) {
	summary := trace.GetSummary()
	
	entry := &EventLogEntry{
		Timestamp:    event.Timestamp(),
		EventType:    event.Type(),
		EntityID:     event.TargetEntityID(),
		EventData:    event.ToMap(),
		HandlerCount: summary["handlerCount"].(int),
		Conflicts:    trace.conflicts,
		DurationUs:   summary["durationNs"].(int64) / 1000,
	}

	// Count effects
	effectCount := 0
	for _, result := range trace.handlerResults {
		effectCount += len(result.Effects)
	}
	entry.EffectCount = effectCount

	// Collect errors
	errors := make([]string, 0)
	for entityID, err := range trace.errors {
		errors = append(errors, fmt.Sprintf("%s: %v", entityID, err))
	}
	entry.Errors = errors

	d.eventLog.Add(entry)
}

// GetEventLog returns the event log
func (d *DebugTools) GetEventLog() *EventLog {
	return d.eventLog
}

// PrintEventLog prints the event log to stdout
func (d *DebugTools) PrintEventLog(n int) {
	entries := d.eventLog.GetRecent(n)
	
	fmt.Printf("\n=== Event Log (last %d events) ===\n", len(entries))
	for i, entry := range entries {
		fmt.Printf("\n[%d] %s - %s (entity: %s)\n",
			i+1,
			entry.Timestamp.Format("15:04:05.000"),
			entry.EventType,
			entry.EntityID,
		)
		fmt.Printf("    Handlers: %d, Effects: %d, Duration: %dμs\n",
			entry.HandlerCount,
			entry.EffectCount,
			entry.DurationUs,
		)
		if len(entry.Conflicts) > 0 {
			fmt.Printf("    ⚠ Conflicts: %d\n", len(entry.Conflicts))
			for _, conflict := range entry.Conflicts {
				fmt.Printf("      - %s: %s\n", conflict.Type, conflict.Description)
			}
		}
		if len(entry.Errors) > 0 {
			fmt.Printf("    ✗ Errors: %d\n", len(entry.Errors))
			for _, err := range entry.Errors {
				fmt.Printf("      - %s\n", err)
			}
		}
	}
}

// PrintWorldVars prints all world variables
func (d *DebugTools) PrintWorldVars() {
	vars := d.worldVars.GetAll()
	
	// Group by scope
	scopes := make(map[string]map[string]interface{})
	for key, value := range vars {
		parts := strings.Split(key, ".")
		if len(parts) < 2 {
			continue
		}
		scope := parts[0]
		if scopes[scope] == nil {
			scopes[scope] = make(map[string]interface{})
		}
		scopes[scope][key] = value
	}

	fmt.Println("\n=== World Variables ===")
	for scope, scopeVars := range scopes {
		fmt.Printf("\n[%s] (%d vars)\n", scope, len(scopeVars))
		
		// Sort keys
		keys := make([]string, 0, len(scopeVars))
		for key := range scopeVars {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Printf("  %s = %v\n", key, scopeVars[key])
		}
	}
}

// PrintEntityStates prints all entity script states
func (d *DebugTools) PrintEntityStates() {
	states := d.dispatcher.stateStore.GetAll()

	fmt.Printf("\n=== Entity States (%d entities) ===\n", len(states))
	
	// Sort by entity ID
	entityIDs := make([]string, 0, len(states))
	for id := range states {
		entityIDs = append(entityIDs, id)
	}
	sort.Strings(entityIDs)

	for _, entityID := range entityIDs {
		state := states[entityID]
		fmt.Printf("\n[%s] (module: %s, version: %d)\n", entityID, state.Module, state.Version)
		
		// Pretty print state data
		jsonData, _ := json.MarshalIndent(state.Data, "  ", "  ")
		fmt.Printf("  %s\n", string(jsonData))
	}
}

// PrintActivePlans prints all active plans
func (d *DebugTools) PrintActivePlans() {
	plans := d.dispatcher.planRunner.GetActivePlans()

	fmt.Printf("\n=== Active Plans (%d plans) ===\n", len(plans))
	
	for planID, cursor := range plans {
		fmt.Printf("\n[%s]\n", planID)
		fmt.Printf("  Entity: %s\n", cursor.EntityID)
		fmt.Printf("  Path: %v\n", cursor.CurrentPath)
		fmt.Printf("  Started: %s\n", cursor.StartTime.Format("15:04:05"))
		if cursor.WaitUntil != nil {
			remaining := time.Until(*cursor.WaitUntil)
			fmt.Printf("  Waiting: %s remaining\n", remaining.Round(time.Millisecond))
		}
		if len(cursor.State) > 0 {
			jsonData, _ := json.MarshalIndent(cursor.State, "  ", "  ")
			fmt.Printf("  State: %s\n", string(jsonData))
		}
	}
}

// PrintRegisteredModules prints all registered script modules
func (d *DebugTools) PrintRegisteredModules() {
	modules := d.dispatcher.registry.GetAll()

	fmt.Printf("\n=== Registered Modules (%d modules) ===\n", len(modules))
	
	// Sort by name
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		module := modules[name]
		fmt.Printf("\n[%s]\n", name)
		fmt.Printf("  Hash: %s\n", module.Hash[:16]+"...")
		
		// List handlers
		handlers := make([]string, 0)
		for eventType, exists := range module.Handlers {
			if exists {
				handlers = append(handlers, string(eventType))
			}
		}
		fmt.Printf("  Handlers: %s\n", strings.Join(handlers, ", "))
		
		// Show subscriptions
		sub := module.Subscriptions
		if sub.Priority != 0 {
			fmt.Printf("  Priority: %d\n", sub.Priority)
		}
		if len(sub.EnterZone) > 0 {
			fmt.Printf("  EnterZone: %v\n", sub.EnterZone)
		}
		if len(sub.LeaveZone) > 0 {
			fmt.Printf("  LeaveZone: %v\n", sub.LeaveZone)
		}
		if len(sub.Trigger) > 0 {
			fmt.Printf("  Trigger: %v\n", sub.Trigger)
		}
		if len(sub.VarPrefix) > 0 {
			fmt.Printf("  VarPrefix: %v\n", sub.VarPrefix)
		}
	}
}

// GenerateReport generates a comprehensive debug report
func (d *DebugTools) GenerateReport() string {
	var sb strings.Builder

	sb.WriteString("=== SCRIPTING SYSTEM DEBUG REPORT ===\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Modules
	modules := d.dispatcher.registry.GetAll()
	sb.WriteString(fmt.Sprintf("Registered Modules: %d\n", len(modules)))

	// Entities
	states := d.dispatcher.stateStore.GetAll()
	sb.WriteString(fmt.Sprintf("Active Entities: %d\n", len(states)))

	// Plans
	plans := d.dispatcher.planRunner.GetActivePlans()
	sb.WriteString(fmt.Sprintf("Active Plans: %d\n", len(plans)))

	// World Vars
	vars := d.worldVars.GetAll()
	sb.WriteString(fmt.Sprintf("World Variables: %d\n", len(vars)))

	// Event Log
	events := d.eventLog.GetAll()
	sb.WriteString(fmt.Sprintf("Event Log Entries: %d\n\n", len(events)))

	// Recent events summary
	recent := d.eventLog.GetRecent(10)
	sb.WriteString("Recent Events (last 10):\n")
	for i, entry := range recent {
		sb.WriteString(fmt.Sprintf("  %d. %s - %s (entity: %s, handlers: %d, effects: %d)\n",
			i+1,
			entry.Timestamp.Format("15:04:05"),
			entry.EventType,
			entry.EntityID,
			entry.HandlerCount,
			entry.EffectCount,
		))
	}

	// Conflicts summary
	conflictCount := 0
	for _, entry := range events {
		conflictCount += len(entry.Conflicts)
	}
	if conflictCount > 0 {
		sb.WriteString(fmt.Sprintf("\n⚠ Total Conflicts Detected: %d\n", conflictCount))
	}

	// Errors summary
	errorCount := 0
	for _, entry := range events {
		errorCount += len(entry.Errors)
	}
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf("✗ Total Errors: %d\n", errorCount))
	}

	return sb.String()
}

// ExportToJSON exports the entire system state to JSON
func (d *DebugTools) ExportToJSON() ([]byte, error) {
	export := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"modules":     d.getModulesExport(),
		"entities":    d.getEntitiesExport(),
		"worldVars":   d.worldVars.GetAll(),
		"activePlans": d.dispatcher.planRunner.GetActivePlans(),
		"eventLog":    d.eventLog.GetRecent(100),
	}

	return json.MarshalIndent(export, "", "  ")
}

func (d *DebugTools) getModulesExport() []map[string]interface{} {
	modules := d.dispatcher.registry.GetAll()
	result := make([]map[string]interface{}, 0, len(modules))

	for name, module := range modules {
		handlers := make([]string, 0)
		for eventType, exists := range module.Handlers {
			if exists {
				handlers = append(handlers, string(eventType))
			}
		}

		result = append(result, map[string]interface{}{
			"name":          name,
			"hash":          module.Hash,
			"handlers":      handlers,
			"subscriptions": module.Subscriptions,
		})
	}

	return result
}

func (d *DebugTools) getEntitiesExport() []map[string]interface{} {
	states := d.dispatcher.stateStore.GetAll()
	result := make([]map[string]interface{}, 0, len(states))

	for entityID, state := range states {
		result = append(result, map[string]interface{}{
			"entityId": entityID,
			"module":   state.Module,
			"version":  state.Version,
			"data":     state.Data,
		})
	}

	return result
}
