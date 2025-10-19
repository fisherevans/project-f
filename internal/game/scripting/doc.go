// Package scripting provides a deterministic, event-driven scripting system for Project F.
//
// This package integrates JavaScript scripts (via Goja) with the Go-based game engine,
// allowing entities to react to world events, emit effects, and persist state cleanly.
//
// # Architecture
//
// The scripting system consists of several key components:
//
//   - ScriptRegistry: Compiles and caches JavaScript modules, detects event handlers
//   - EventDispatcher: Routes events to matching entity handlers with deterministic ordering
//   - EffectApplier: Validates and applies effects with conflict resolution
//   - PlanRunner: Executes multi-step sequences over time (pausable and serializable)
//   - WorldVarStore: Namespaced key-value store for shared game state
//   - StateStore: Per-entity script state with versioning and migration support
//
// # Event Flow
//
// Events follow a two-phase pipeline:
//
//  1. Reduce Phase: Collect all handler outputs (effects, plans, state updates)
//  2. Commit Phase: Validate and apply effects, start plans, persist states
//
// Handlers are executed in deterministic order: priority DESC, then entityID ASC.
//
// # Example Usage
//
//	// Initialize system
//	registry := scripting.NewScriptRegistry()
//	worldVars := scripting.NewWorldVarStore()
//	dispatcher := scripting.NewEventDispatcher(registry, worldVars)
//
//	// Register a script
//	script := `
//	    function OnInteract(self, world, state, event) {
//	        return {
//	            effects: [{
//	                type: "ShowDialog",
//	                data: { text: "Hello!" }
//	            }]
//	        };
//	    }
//	`
//	registry.Register("npc_guard", script)
//
//	// Attach to entity
//	dispatcher.AttachScript("guard_001", "npc_guard", nil)
//
//	// Dispatch event
//	event := scripting.NewInteractEvent("guard_001", "player")
//	dispatcher.Dispatch(event)
//
// # Event Types
//
// The system supports the following event types:
//
//   - Init: Entity first created
//   - Interact: Entity interacted with by player
//   - EnterZone: Entity enters a named zone
//   - LeaveZone: Entity leaves a named zone
//   - Timer: Named timer expires
//   - Trigger: Named trigger activated
//   - FlagChanged: World variable changes
//   - Step: Periodic background update
//
// # Effect Types
//
// Scripts can return effects that modify the game world:
//
//   - SetVar, IncVar, ClearVar: Modify world variables
//   - MoveTo: Move entity to position
//   - OpenDoor, CloseDoor: Control doors
//   - Trigger: Activate named trigger
//   - StartBattle: Initiate combat
//   - SetCamera: Control camera target
//   - PlayMusic, PlaySound: Audio control
//   - ShowDialog: Display text
//   - StartTimer, StopTimer: Timer control
//   - SpawnEntity, RemoveEntity: Entity lifecycle
//   - Transaction: Atomic multi-effect operation
//
// # Plans
//
// Plans allow multi-step sequences executed over time:
//
//   - Seq: Sequential execution
//   - Par: Parallel execution
//   - Wait: Delay for duration
//   - If: Conditional branching
//   - Choice: Player choice (waits for input)
//   - Effect: Single effect node
//
// Example plan:
//
//	{
//	    type: "Seq",
//	    children: [
//	        { type: "Effect", data: { type: "ShowDialog", data: { text: "Step 1" } } },
//	        { type: "Wait", data: { duration: 2.0 } },
//	        { type: "Effect", data: { type: "ShowDialog", data: { text: "Step 2" } } }
//	    ]
//	}
//
// # Subscriptions
//
// Scripts can filter which events they receive:
//
//	const SUBSCRIPTION = {
//	    EnterZone: ["power_room"],      // Only these zones
//	    Trigger: ["QuestComplete"],     // Only these triggers
//	    VarPrefix: ["map.meadow.*"],    // Watch variable prefixes
//	    Priority: 10                    // Handler priority
//	};
//
// # State Management
//
// Each entity has versioned state that persists across save/load:
//
//	{
//	    "module": "npc_guard",
//	    "version": 1,
//	    "data": { "greeted": true, "visits": 3 }
//	}
//
// State migrations are supported:
//
//	const STATE_VERSION = 2;
//
//	function Migrate(state, fromVersion) {
//	    if (fromVersion === 1) {
//	        return { ...state, newField: "default" };
//	    }
//	    return state;
//	}
//
// # World Variables
//
// Namespaced key-value store for shared state:
//
//   - global.*: Global game state
//   - map.*: Map-specific state
//   - topic.*: Thematic groupings (e.g., topic.security.alarm_level)
//
// Variables support change listeners with prefix matching.
//
// # Determinism
//
// The system ensures deterministic execution:
//
//   - Single-threaded script execution
//   - No direct mutation of Go state from JavaScript
//   - No Date or Math.random (use engine-provided getTime() and random())
//   - Stable ordering (priority DESC, entityID ASC)
//   - All updates via validated effects
//
// # Conflict Resolution
//
// When multiple effects target the same resource:
//
//   - Variables: Last-writer-wins (by priority, then order)
//   - Doors: Last-writer-wins with conflict logged
//   - Camera/Music: Highest priority wins
//   - Transactions: All-or-nothing (atomic)
//
// # Save/Load
//
// The system serializes to JSON:
//
//	{
//	    "world": { "vars": { "global.power.online": true } },
//	    "entities": [
//	        { "id": "guard_1", "script": { "module": "npc_guard", "version": 1, "data": {...} } }
//	    ],
//	    "plans": [
//	        { "entity": "console_a", "cursor": {...} }
//	    ]
//	}
//
// # Debugging
//
// Debug tools provide introspection:
//
//	debug := scripting.NewDebugTools(dispatcher, worldVars)
//	debug.PrintEventLog(10)        // Recent events
//	debug.PrintWorldVars()         // All variables
//	debug.PrintEntityStates()      // Entity states
//	debug.PrintActivePlans()       // Running plans
//	report := debug.GenerateReport() // Full report
//
// # TypeScript Support
//
// TypeScript definitions are provided in assets/scripts/global.d.ts for IDE support.
//
// # Performance
//
// The system is optimized for game use:
//
//   - Scripts compiled once and cached by hash
//   - Event filtering before handler execution
//   - Incremental plan execution (doesn't block)
//   - Conflict detection is O(n) per effect type
//
// # Integration
//
// See docs/scripting_integration_guide.md for integration with the adventure state.
package scripting
