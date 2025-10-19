# Getting Started with Entity Scripting

This guide will help you start using the scripting system in 5 minutes.

## Quick Start

### 1. Run the Integration Example

```bash
cd /Users/fisher/dev/project-f
go run -tags=example internal/game/scripting/integration_example.go
```

Or in your code:

```go
import "fisherevans.com/project/f/internal/game/scripting"

func main() {
    scripting.RunIntegrationExample()
}
```

This will show you a complete working example with output.

### 2. Run the Tests

```bash
cd internal/game/scripting
go test -v
```

All tests should pass, demonstrating the system works correctly.

### 3. Create Your First Script

Create a file `assets/scripts/my_first_npc.js`:

```javascript
function OnInit(self, world, state) {
    return {
        state: {
            greeted: false,
            talkCount: 0
        }
    };
}

function OnInteract(self, world, state, event) {
    const count = (state.talkCount || 0) + 1;
    
    if (!state.greeted) {
        return {
            effects: [{
                type: "ShowDialog",
                data: {
                    text: "Hello! I'm your first scripted NPC!",
                    speaker: "Tutorial NPC"
                }
            }],
            state: {
                greeted: true,
                talkCount: count
            }
        };
    } else {
        return {
            effects: [{
                type: "ShowDialog",
                data: {
                    text: "We've talked " + count + " times now.",
                    speaker: "Tutorial NPC"
                }
            }],
            state: {
                greeted: true,
                talkCount: count
            }
        };
    }
}
```

### 4. Test Your Script

```go
package main

import (
    "fmt"
    "os"
    "fisherevans.com/project/f/internal/game/scripting"
)

func main() {
    // Initialize
    registry := scripting.NewScriptRegistry()
    worldVars := scripting.NewWorldVarStore()
    dispatcher := scripting.NewEventDispatcher(registry, worldVars)
    
    // Load your script
    scriptData, _ := os.ReadFile("assets/scripts/my_first_npc.js")
    registry.Register("my_first_npc", string(scriptData))
    
    // Attach to entity
    dispatcher.AttachScript("npc_001", "my_first_npc", nil)
    
    // Dispatch Init
    initEvent := scripting.NewInitEvent("npc_001")
    dispatcher.Dispatch(initEvent)
    
    // Dispatch Interact
    interactEvent := scripting.NewInteractEvent("npc_001", "player")
    dispatcher.Dispatch(interactEvent)
    
    // Check state
    state, _ := dispatcher.GetEntityState("npc_001")
    fmt.Printf("NPC State: %+v\n", state.Data)
}
```

## Common Use Cases

### Stateful NPC Dialogue

```javascript
function OnInteract(self, world, state, event) {
    const stage = state.questStage || 0;
    
    if (stage === 0) {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "Can you help me find my lost item?" }
            }],
            state: { questStage: 1 }
        };
    } else if (stage === 1) {
        const hasItem = world.getVar("global.inventory.lost_item");
        if (hasItem) {
            return {
                effects: [
                    { type: "ShowDialog", data: { text: "You found it! Thank you!" } },
                    { type: "SetVar", data: { key: "global.quests.lost_item.complete", value: true } }
                ],
                state: { questStage: 2 }
            };
        } else {
            return {
                effects: [{
                    type: "ShowDialog",
                    data: { text: "Please find my lost item!" }
                }]
            };
        }
    } else {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "Thanks again for your help!" }
            }]
        };
    }
}
```

### Door That Reacts to Power

```javascript
const SUBSCRIPTION = {
    VarPrefix: ["global.power.*"]
};

function OnInit(self, world, state) {
    const powered = world.getVar("global.power.grid_online");
    return {
        effects: powered ? [] : [{ type: "CloseDoor", data: { doorId: self.id } }],
        state: { locked: !powered }
    };
}

function OnFlagChanged(self, world, state, event) {
    if (event.varName === "global.power.grid_online") {
        if (event.newValue && state.locked) {
            return {
                effects: [
                    { type: "OpenDoor", data: { doorId: self.id } },
                    { type: "ShowDialog", data: { text: "Door unlocked!" } }
                ],
                state: { locked: false }
            };
        }
    }
}
```

### Patrol AI with Timers

```javascript
function OnInit(self, world, state) {
    return {
        effects: [{ type: "StartTimer", data: { timerName: "patrol", duration: 3.0 } }],
        state: { 
            patrolIndex: 0,
            waypoints: [
                { x: 10, y: 5 },
                { x: 15, y: 5 },
                { x: 15, y: 10 },
                { x: 10, y: 10 }
            ]
        }
    };
}

function OnTimer(self, world, state, event) {
    if (event.timerName === "patrol") {
        const nextIndex = (state.patrolIndex + 1) % state.waypoints.length;
        const next = state.waypoints[nextIndex];
        
        return {
            effects: [
                { type: "MoveTo", data: { x: next.x, y: next.y } },
                { type: "StartTimer", data: { timerName: "patrol", duration: 3.0 } }
            ],
            state: { 
                patrolIndex: nextIndex,
                waypoints: state.waypoints
            }
        };
    }
}
```

## Debugging Your Scripts

### View Event Log

```go
debug := scripting.NewDebugTools(dispatcher, worldVars)
debug.PrintEventLog(10) // Last 10 events
```

### View World Variables

```go
debug.PrintWorldVars()
```

### View Entity States

```go
debug.PrintEntityStates()
```

### Export Full State

```go
jsonData, _ := debug.ExportToJSON()
os.WriteFile("debug_state.json", jsonData, 0644)
```

## Next Steps

1. **Read the Quick Reference**: `assets/scripts/QUICK_REFERENCE.md`
2. **Study Examples**: `assets/scripts/examples/`
3. **Read Full Documentation**: `internal/game/scripting/README.md`
4. **Integration Guide**: `docs/scripting_integration_guide.md`

## TypeScript Support

For IDE autocomplete, reference the type definitions:

```javascript
/// <reference path="./global.d.ts" />

function OnInteract(self, world, state, event) {
    // Now you get autocomplete for:
    // - self.id
    // - world.getVar()
    // - event.sourceEntityId
    // - etc.
}
```

## Common Pitfalls

### ❌ Don't use Date or Math.random

```javascript
// BAD - not deterministic
const now = new Date();
const rand = Math.random();

// GOOD - deterministic
const now = getTime();
const rand = random();
```

### ❌ Don't mutate state directly

```javascript
// BAD - doesn't persist
state.count++;

// GOOD - return new state
return {
    state: { count: (state.count || 0) + 1 }
};
```

### ❌ Don't forget to validate effect data

```javascript
// BAD - missing required field
{
    type: "SetVar",
    data: { value: 123 } // missing 'key'
}

// GOOD - all required fields
{
    type: "SetVar",
    data: { key: "global.test", value: 123 }
}
```

## Performance Tips

1. **Use subscriptions** to filter events
2. **Batch effects** instead of multiple handlers
3. **Use plans** for sequences instead of nested timers
4. **Cache computed values** in state

## Getting Help

- **Quick Reference**: `assets/scripts/QUICK_REFERENCE.md`
- **Full Docs**: `internal/game/scripting/README.md`
- **Examples**: `assets/scripts/examples/`
- **Tests**: `internal/game/scripting/example_test.go`

## Summary

You now have:
- ✅ A working scripting system
- ✅ Example scripts to learn from
- ✅ TypeScript definitions for IDE support
- ✅ Debugging tools
- ✅ Complete documentation

Start by running the integration example, then modify the example scripts to experiment!
