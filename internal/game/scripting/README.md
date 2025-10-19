# Project F - Entity Scripting & Event System

A deterministic, event-driven scripting system for Project F that integrates JavaScript scripts (via Goja) with the Go-based game engine.

## Overview

This system allows entities to react to world events, emit effects, and persist local and shared state cleanly. All scripting is deterministic, single-threaded, and designed for save/load compatibility.

## Architecture

### Core Components

1. **ScriptRegistry** (`registry.go`)
   - Compiles and caches JavaScript modules
   - Detects event handlers and subscriptions
   - Creates isolated VM executors per entity

2. **EventDispatcher** (`dispatcher.go`)
   - Routes events to matching entity handlers
   - Implements two-phase pipeline (Reduce + Commit)
   - Maintains deterministic ordering (priority DESC, entityID ASC)

3. **EffectApplier** (`effects.go`)
   - Validates and applies effects with conflict resolution
   - Supports transactions (all-or-nothing)
   - Policies: last-writer-wins for vars, highest-priority for camera/music

4. **PlanRunner** (`plan.go`)
   - Executes multi-step sequences (Seq, Par, Wait, If, Choice)
   - Pausable and serializable
   - Resumes from save files

5. **WorldVarStore** (`worldvars.go`)
   - Namespaced key-value store (`global.*`, `map.*`, `topic.*`)
   - Change listeners with prefix matching
   - Persisted in save files

6. **StateStore** (`state.go`)
   - Per-entity script state with versioning
   - Migration support for schema upgrades
   - Save/load serialization

## Event Types

| Event | Description | Handler |
|-------|-------------|---------|
| `Init` | Entity first created | `OnInit(self, world, state)` |
| `Interact` | Entity interacted with | `OnInteract(self, world, state, event)` |
| `EnterZone` | Entity enters zone | `OnEnterZone(self, world, state, event)` |
| `LeaveZone` | Entity leaves zone | `OnLeaveZone(self, world, state, event)` |
| `Timer` | Timer expires | `OnTimer(self, world, state, event)` |
| `Trigger` | Named trigger activated | `OnTrigger(self, world, state, event)` |
| `FlagChanged` | World variable changes | `OnFlagChanged(self, world, state, event)` |
| `Step` | Periodic background tick | `OnStep(self, world, state, event)` |

## Effect Types

Effects are immediate actions applied to the game world:

- **Variables**: `SetVar`, `IncVar`, `ClearVar`
- **Movement**: `MoveTo`
- **Doors**: `OpenDoor`, `CloseDoor`
- **Events**: `Trigger`
- **Combat**: `StartBattle`
- **Camera**: `SetCamera`
- **Audio**: `PlayMusic`, `PlaySound`
- **UI**: `ShowDialog`
- **Timers**: `StartTimer`, `StopTimer`
- **Entities**: `SpawnEntity`, `RemoveEntity`
- **Transactions**: `Transaction` (atomic multi-effect)

## Plan Types

Plans are multi-step sequences executed over time:

- **Seq**: Sequential execution
- **Par**: Parallel execution
- **Wait**: Delay for duration
- **If**: Conditional branching
- **Choice**: Player choice (waits for input)
- **Effect**: Single effect node

## Usage

### 1. Initialize the System

```go
import "fisherevans.com/project/f/internal/game/scripting"

// Create core components
registry := scripting.NewScriptRegistry()
worldVars := scripting.NewWorldVarStore()
dispatcher := scripting.NewEventDispatcher(registry, worldVars)
saveSystem := scripting.NewSaveSystem(dispatcher, worldVars)
```

### 2. Register Scripts

```go
script := `
    const SUBSCRIPTION = {
        Priority: 10,
        EnterZone: ["power_room"]
    };

    function OnInit(self, world, state) {
        return { state: { initialized: true } };
    }

    function OnInteract(self, world, state, event) {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "Hello, traveler!" }
            }],
            state: { lastInteraction: getTime() }
        };
    }
`

module, err := registry.Register("npc_guard", script)
```

### 3. Attach to Entities

```go
err := dispatcher.AttachScript("guard_001", "npc_guard", nil)
```

### 4. Dispatch Events

```go
// Init event
initEvent := scripting.NewInitEvent("guard_001")
dispatcher.Dispatch(initEvent)

// Interact event
interactEvent := scripting.NewInteractEvent("guard_001", "player")
dispatcher.Dispatch(interactEvent)
```

### 5. Access World Variables

```go
// Set variables
worldVars.Set("global.power.grid_online", true)
worldVars.Set("map.meadow.gate.open", false)

// Get variables
powerOn := worldVars.GetBool("global.power.grid_online")

// Listen to changes
worldVars.AddListener("map.meadow.*", func(key string, oldValue, newValue interface{}) {
    fmt.Printf("Variable changed: %s = %v\n", key, newValue)
})
```

### 6. Save/Load

```go
// Save
save, err := saveSystem.Save()
jsonData, err := saveSystem.SerializeToJSON(save)
os.WriteFile("save.json", jsonData, 0644)

// Load
jsonData, err := os.ReadFile("save.json")
save, err := saveSystem.DeserializeFromJSON(jsonData)
err = saveSystem.Load(save)
```

## Script Examples

### Basic NPC

```javascript
function OnInit(self, world, state) {
    return { state: { greeted: false } };
}

function OnInteract(self, world, state, event) {
    if (!state.greeted) {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "Welcome, stranger!", speaker: "Guard" }
            }],
            state: { greeted: true }
        };
    } else {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "We've met before.", speaker: "Guard" }
            }]
        };
    }
}
```

### Zone Trigger

```javascript
const SUBSCRIPTION = {
    EnterZone: ["power_room"],
    Priority: 5
};

function OnEnterZone(self, world, state, event) {
    const powerOn = world.getVar("global.power.grid_online");
    
    if (!powerOn) {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "The power is offline." }
            }]
        };
    }
}
```

### Quest with Plan

```javascript
function OnInteract(self, world, state, event) {
    return {
        plan: {
            type: "Seq",
            children: [
                {
                    type: "Effect",
                    data: {
                        type: "ShowDialog",
                        data: { text: "I have a quest for you!" }
                    }
                },
                {
                    type: "Wait",
                    data: { duration: 2.0 }
                },
                {
                    type: "Effect",
                    data: {
                        type: "SetVar",
                        data: { key: "global.quests.main.active", value: true }
                    }
                }
            ]
        }
    };
}
```

## Subscriptions

Control which events an entity receives:

```javascript
const SUBSCRIPTION = {
    // Zone filters
    EnterZone: ["power_room", "security_office"],
    LeaveZone: ["power_room"],
    
    // Trigger filters
    Trigger: ["QuestComplete", "AlarmTriggered"],
    
    // Variable watch patterns
    VarPrefix: ["map.meadow.*", "global.power.*"],
    
    // Handler priority (higher = earlier)
    Priority: 10
};
```

## Debugging

### Event Tracing

```go
trace := dispatcher.GetTrace()
summary := trace.GetSummary()
fmt.Printf("Handlers: %d, Effects: %d, Conflicts: %d\n",
    summary["handlerCount"],
    summary["effectCount"],
    summary["conflictCount"])
```

### Debug Tools

```go
debug := scripting.NewDebugTools(dispatcher, worldVars)

// Print event log
debug.PrintEventLog(10)

// Print world variables
debug.PrintWorldVars()

// Print entity states
debug.PrintEntityStates()

// Print active plans
debug.PrintActivePlans()

// Generate report
report := debug.GenerateReport()
fmt.Println(report)

// Export to JSON
jsonData, _ := debug.ExportToJSON()
```

## Conflict Resolution

When multiple effects target the same resource:

- **Variables**: Last-writer-wins (by priority, then order)
- **Doors**: Last-writer-wins with conflict logged
- **Camera**: Highest priority wins
- **Music**: Highest priority wins
- **Transactions**: All-or-nothing (atomic)

Conflicts are logged and available via `applier.GetConflicts()`.

## Determinism

The system ensures deterministic execution:

- Single-threaded script execution
- No direct mutation of Go state from JS
- No `Date` or `Math.random` (use engine-provided `getTime()` and `random()`)
- Stable ordering (priority DESC, entityID ASC)
- All cross-system updates via validated effects

## State Versioning

Scripts can define version and migration:

```javascript
const STATE_VERSION = 2;

function Migrate(state, fromVersion) {
    if (fromVersion === 1) {
        return {
            ...state,
            newField: "default_value"
        };
    }
    return state;
}
```

## Testing

Run the test suite:

```bash
cd internal/game/scripting
go test -v
```

Run the integration example:

```go
import "fisherevans.com/project/f/internal/game/scripting"

scripting.RunIntegrationExample()
```

## TypeScript Support

TypeScript definitions are provided in `assets/scripts/global.d.ts` for IDE autocomplete and type checking.

## File Structure

```
internal/game/scripting/
├── events.go              # Event types and definitions
├── effects.go             # Effect types and applier
├── worldvars.go           # World variable store
├── plan.go                # Plan runner
├── registry.go            # Script compilation and caching
├── dispatcher.go          # Event routing and dispatch
├── state.go               # State persistence and versioning
├── debug.go               # Debugging utilities
├── integration_example.go # Complete integration example
├── example_test.go        # Test suite
└── README.md              # This file

assets/scripts/
├── global.d.ts            # TypeScript definitions
└── examples/
    ├── npc_gatekeeper.js  # Stateful NPC example
    ├── zone_trigger.js    # Zone event example
    ├── quest_npc.js       # Quest with plans
    └── patrol_guard.js    # Timer-based patrol
```

## Integration with Adventure State

To integrate with the existing adventure state system:

```go
// In adventure state initialization
func (s *State) initScripting() {
    s.scriptRegistry = scripting.NewScriptRegistry()
    s.worldVars = scripting.NewWorldVarStore()
    s.scriptDispatcher = scripting.NewEventDispatcher(s.scriptRegistry, s.worldVars)
}

// When entity interacts
func (e *Entity) Interact(adv *State, source Entity) {
    // Dispatch to scripting system
    event := scripting.NewInteractEvent(string(e.GetEntityId()), string(source.GetEntityId()))
    adv.scriptDispatcher.Dispatch(event)
}
```

## Performance Considerations

- Scripts are compiled once and cached by hash
- VM instances are pooled per entity
- Event filtering happens before handler execution
- Conflict detection is O(n) per effect type
- Plan execution is incremental (doesn't block)

## Future Enhancements

- [ ] Hot-reload scripts during development
- [ ] Script profiling and performance metrics
- [ ] Visual plan debugger
- [ ] Script sandboxing with resource limits
- [ ] Network event synchronization
- [ ] Script unit testing framework
