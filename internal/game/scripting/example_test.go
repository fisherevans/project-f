package scripting

import (
	"encoding/json"
	"testing"
)

// TestBasicScriptExecution demonstrates the complete workflow
func TestBasicScriptExecution(t *testing.T) {
	// Setup
	registry := NewScriptRegistry()
	worldVars := NewWorldVarStore()
	dispatcher := NewEventDispatcher(registry, worldVars)

	// Register a simple script
	script := `
		function OnInit(self, world, state) {
			return {
				state: { initialized: true, count: 0 }
			};
		}

		function OnInteract(self, world, state, event) {
			const count = (state.count || 0) + 1;
			
			return {
				effects: [
					{
						type: "SetVar",
						data: {
							key: "global.test.interactions",
							value: count
						}
					},
					{
						type: "ShowDialog",
						data: {
							text: "Hello! Interaction #" + count
						}
					}
				],
				state: {
					initialized: true,
					count: count
				}
			};
		}
	`

	module, err := registry.Register("test_npc", script)
	if err != nil {
		t.Fatalf("Failed to register script: %v", err)
	}

	// Verify handlers were detected
	if !module.Handlers[EventTypeInit] {
		t.Error("OnInit handler not detected")
	}
	if !module.Handlers[EventTypeInteract] {
		t.Error("OnInteract handler not detected")
	}

	// Attach script to entity
	err = dispatcher.AttachScript("npc_001", "test_npc", nil)
	if err != nil {
		t.Fatalf("Failed to attach script: %v", err)
	}

	// Dispatch Init event
	initEvent := NewInitEvent("npc_001")
	err = dispatcher.Dispatch(initEvent)
	if err != nil {
		t.Fatalf("Failed to dispatch init event: %v", err)
	}

	// Verify state was updated
	state, ok := dispatcher.GetEntityState("npc_001")
	if !ok {
		t.Fatal("Entity state not found")
	}
	if !state.Data["initialized"].(bool) {
		t.Error("Entity not initialized")
	}

	// Dispatch Interact event
	interactEvent := NewInteractEvent("npc_001", "player")
	err = dispatcher.Dispatch(interactEvent)
	if err != nil {
		t.Fatalf("Failed to dispatch interact event: %v", err)
	}

	// Verify world var was set
	interactions := worldVars.GetInt("global.test.interactions")
	if interactions != 1 {
		t.Errorf("Expected 1 interaction, got %d", interactions)
	}

	// Dispatch another interact
	err = dispatcher.Dispatch(interactEvent)
	if err != nil {
		t.Fatalf("Failed to dispatch second interact event: %v", err)
	}

	interactions = worldVars.GetInt("global.test.interactions")
	if interactions != 2 {
		t.Errorf("Expected 2 interactions, got %d", interactions)
	}

	// Verify state count
	state, _ = dispatcher.GetEntityState("npc_001")
	count := int(state.Data["count"].(float64))
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestSubscriptionFiltering tests event filtering based on subscriptions
func TestSubscriptionFiltering(t *testing.T) {
	registry := NewScriptRegistry()
	worldVars := NewWorldVarStore()
	dispatcher := NewEventDispatcher(registry, worldVars)

	// Script that only listens to specific zones
	script := `
		const SUBSCRIPTION = {
			EnterZone: ["power_room"],
			Priority: 10
		};

		function OnEnterZone(self, world, state, event) {
			return {
				effects: [{
					type: "SetVar",
					data: {
						key: "global.test.zone_entered",
						value: event.zoneName
					}
				}]
			};
		}
	`

	_, err := registry.Register("zone_listener", script)
	if err != nil {
		t.Fatalf("Failed to register script: %v", err)
	}

	err = dispatcher.AttachScript("trigger_001", "zone_listener", nil)
	if err != nil {
		t.Fatalf("Failed to attach script: %v", err)
	}

	// Dispatch EnterZone for power_room (should trigger)
	event1 := NewEnterZoneEvent("trigger_001", "power_room")
	err = dispatcher.Dispatch(event1)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	zoneName := worldVars.GetString("global.test.zone_entered")
	if zoneName != "power_room" {
		t.Errorf("Expected 'power_room', got '%s'", zoneName)
	}

	// Dispatch EnterZone for different zone (should not trigger)
	worldVars.Clear("global.test.zone_entered")
	event2 := NewEnterZoneEvent("trigger_001", "security_office")
	err = dispatcher.Dispatch(event2)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	if worldVars.Has("global.test.zone_entered") {
		t.Error("Handler should not have been called for non-subscribed zone")
	}
}

// TestEffectValidation tests effect validation
func TestEffectValidation(t *testing.T) {
	tests := []struct {
		name      string
		effect    Effect
		shouldErr bool
	}{
		{
			name: "valid SetVar",
			effect: Effect{
				Type: EffectTypeSetVar,
				Data: map[string]interface{}{
					"key":   "test.key",
					"value": 123,
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid SetVar - missing key",
			effect: Effect{
				Type: EffectTypeSetVar,
				Data: map[string]interface{}{
					"value": 123,
				},
			},
			shouldErr: true,
		},
		{
			name: "valid MoveTo",
			effect: Effect{
				Type: EffectTypeMoveTo,
				Data: map[string]interface{}{
					"x": 10.0,
					"y": 20.0,
				},
			},
			shouldErr: false,
		},
		{
			name: "invalid MoveTo - missing y",
			effect: Effect{
				Type: EffectTypeMoveTo,
				Data: map[string]interface{}{
					"x": 10.0,
				},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.effect.Validate()
			if tt.shouldErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestWorldVarStore tests the world variable store
func TestWorldVarStore(t *testing.T) {
	store := NewWorldVarStore()

	// Test basic set/get
	store.Set("global.test.value", 42)
	if store.GetInt("global.test.value") != 42 {
		t.Error("Failed to get int value")
	}

	// Test prefix matching
	store.Set("map.meadow.gate.open", true)
	store.Set("map.meadow.chest.looted", false)
	store.Set("map.forest.tree.cut", true)

	meadowVars := store.GetByPrefix("map.meadow.")
	if len(meadowVars) != 2 {
		t.Errorf("Expected 2 meadow vars, got %d", len(meadowVars))
	}

	// Test listeners
	var lastKey string
	var lastValue interface{}

	store.AddListener("map.meadow.*", func(key string, oldValue, newValue interface{}) {
		lastKey = key
		lastValue = newValue
	})

	store.Set("map.meadow.gate.open", false)
	if lastKey != "map.meadow.gate.open" || lastValue.(bool) != false {
		t.Error("Listener not called correctly")
	}

	// Test serialization
	data, err := store.Serialize()
	if err != nil {
		t.Fatalf("Failed to serialize: %v", err)
	}

	newStore := NewWorldVarStore()
	err = newStore.Deserialize(data)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if newStore.GetInt("global.test.value") != 42 {
		t.Error("Value not preserved after serialization")
	}
}

// TestSaveLoad tests the save/load system
func TestSaveLoad(t *testing.T) {
	// Setup
	registry := NewScriptRegistry()
	worldVars := NewWorldVarStore()
	dispatcher := NewEventDispatcher(registry, worldVars)
	saveSystem := NewSaveSystem(dispatcher, worldVars)

	// Register script
	script := `
		function OnInit(self, world, state) {
			return { state: { value: 100 } };
		}
	`
	registry.Register("test_script", script)

	// Setup world state
	worldVars.Set("global.test.value", 42)
	dispatcher.AttachScript("entity_1", "test_script", map[string]interface{}{"value": 100})

	// Create save
	save, err := saveSystem.Save()
	if err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Verify save structure
	if len(save.Entities) != 1 {
		t.Errorf("Expected 1 entity in save, got %d", len(save.Entities))
	}
	if save.World.Vars["global.test.value"] != 42 {
		t.Error("World var not saved correctly")
	}

	// Serialize to JSON
	jsonData, err := saveSystem.SerializeToJSON(save)
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	// Verify JSON structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Create new system and load
	newRegistry := NewScriptRegistry()
	newWorldVars := NewWorldVarStore()
	newDispatcher := NewEventDispatcher(newRegistry, newWorldVars)
	newSaveSystem := NewSaveSystem(newDispatcher, newWorldVars)

	// Register script in new registry
	newRegistry.Register("test_script", script)

	// Load from JSON
	loadedSave, err := newSaveSystem.DeserializeFromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize from JSON: %v", err)
	}

	err = newSaveSystem.Load(loadedSave)
	if err != nil {
		t.Fatalf("Failed to load: %v", err)
	}

	// Verify loaded state
	if newWorldVars.GetInt("global.test.value") != 42 {
		t.Error("World var not loaded correctly")
	}

	state, ok := newDispatcher.GetEntityState("entity_1")
	if !ok {
		t.Fatal("Entity state not loaded")
	}
	if state.Module != "test_script" {
		t.Error("Entity module not loaded correctly")
	}
}

// TestConflictDetection tests conflict detection in effects
func TestConflictDetection(t *testing.T) {
	worldVars := NewWorldVarStore()
	applier := NewEffectApplier(worldVars)

	// Create conflicting effects
	effects := []Effect{
		{
			Type: EffectTypeSetVar,
			Data: map[string]interface{}{
				"key":   "test.value",
				"value": 100,
			},
		},
		{
			Type: EffectTypeSetVar,
			Data: map[string]interface{}{
				"key":   "test.value",
				"value": 200,
			},
		},
	}

	err := applier.ApplyEffects(effects)
	if err != nil {
		t.Fatalf("Failed to apply effects: %v", err)
	}

	// Check for conflicts
	conflicts := applier.GetConflicts()
	if len(conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].Type != "VarConflict" {
		t.Errorf("Expected VarConflict, got %s", conflicts[0].Type)
	}

	// Last writer should win
	if worldVars.GetInt("test.value") != 200 {
		t.Error("Last writer did not win")
	}
}
