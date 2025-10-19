# Script Quick Reference

## Event Handlers

```javascript
// Called when entity is created
function OnInit(self, world, state) {
    return {
        state: { initialized: true }
    };
}

// Called when entity is interacted with
function OnInteract(self, world, state, event) {
    // event.sourceEntityId - who interacted
    return {
        effects: [/* ... */],
        state: { /* ... */ }
    };
}

// Called when entity enters a zone
function OnEnterZone(self, world, state, event) {
    // event.zoneName - which zone
    return { /* ... */ };
}

// Called when entity leaves a zone
function OnLeaveZone(self, world, state, event) {
    // event.zoneName - which zone
    return { /* ... */ };
}

// Called when a timer expires
function OnTimer(self, world, state, event) {
    // event.timerName - which timer
    return { /* ... */ };
}

// Called when a trigger is activated
function OnTrigger(self, world, state, event) {
    // event.triggerName - which trigger
    // event.data - optional trigger data
    return { /* ... */ };
}

// Called when a watched variable changes
function OnFlagChanged(self, world, state, event) {
    // event.varName - which variable
    // event.oldValue - previous value
    // event.newValue - new value
    return { /* ... */ };
}

// Called periodically for background processing
function OnStep(self, world, state, event) {
    // event.deltaTime - time since last step
    return { /* ... */ };
}
```

## Subscriptions

```javascript
const SUBSCRIPTION = {
    // Only receive EnterZone events for these zones
    EnterZone: ["power_room", "security_office"],
    
    // Only receive LeaveZone events for these zones
    LeaveZone: ["power_room"],
    
    // Only receive Trigger events with these names
    Trigger: ["QuestComplete", "AlarmTriggered"],
    
    // Receive FlagChanged for variables matching these prefixes
    VarPrefix: ["map.meadow.*", "global.power.*"],
    
    // Handler priority (higher = executes earlier)
    Priority: 10
};
```

## World API

```javascript
// Get a variable value
const value = world.getVar("global.power.grid_online");

// Check if a variable exists
if (world.hasVar("map.meadow.gate.open")) {
    // ...
}
```

## Effects

### Variables

```javascript
// Set a variable
{
    type: "SetVar",
    data: {
        key: "global.power.grid_online",
        value: true
    }
}

// Increment a variable
{
    type: "IncVar",
    data: {
        key: "global.score",
        value: 10
    }
}

// Clear a variable
{
    type: "ClearVar",
    data: {
        key: "map.temp.flag"
    }
}
```

### Movement

```javascript
{
    type: "MoveTo",
    data: {
        x: 10,
        y: 5
    }
}
```

### Doors

```javascript
{
    type: "OpenDoor",
    data: {
        doorId: "north_gate"
    }
}

{
    type: "CloseDoor",
    data: {
        doorId: "north_gate"
    }
}
```

### Triggers

```javascript
{
    type: "Trigger",
    data: {
        triggerName: "QuestComplete",
        targetEntity: "quest_giver",  // optional
        data: { questId: "main_quest" }  // optional
    }
}
```

### Combat

```javascript
{
    type: "StartBattle",
    data: {
        battleId: "boss_fight"
    }
}
```

### Camera

```javascript
{
    type: "SetCamera",
    data: {
        target: "player"  // or entity ID
    }
}
```

### Audio

```javascript
{
    type: "PlayMusic",
    data: {
        musicId: "battle_theme",
        fadeIn: 1.0  // optional
    }
}

{
    type: "PlaySound",
    data: {
        soundId: "door_open",
        volume: 0.8  // optional
    }
}
```

### UI

```javascript
{
    type: "ShowDialog",
    data: {
        text: "Hello, traveler!",
        speaker: "Guard"  // optional
    }
}
```

### Timers

```javascript
{
    type: "StartTimer",
    data: {
        timerName: "patrol",
        duration: 3.0  // seconds
    }
}

{
    type: "StopTimer",
    data: {
        timerName: "patrol"
    }
}
```

### Entities

```javascript
{
    type: "SpawnEntity",
    data: {
        entityType: "enemy_guard",
        x: 10,
        y: 5,
        properties: { level: 5 }  // optional
    }
}

{
    type: "RemoveEntity",
    data: {
        entityId: "temp_npc"  // optional if used in effect.entityId
    }
}
```

### Transactions

```javascript
{
    type: "Transaction",
    data: {
        effects: [
            { type: "SetVar", data: { key: "a", value: 1 } },
            { type: "SetVar", data: { key: "b", value: 2 } }
        ]
    }
}
```

## Plans

### Sequential

```javascript
{
    type: "Seq",
    children: [
        { type: "Effect", data: { type: "ShowDialog", data: { text: "Step 1" } } },
        { type: "Wait", data: { duration: 2.0 } },
        { type: "Effect", data: { type: "ShowDialog", data: { text: "Step 2" } } }
    ]
}
```

### Parallel

```javascript
{
    type: "Par",
    children: [
        { type: "Effect", data: { type: "PlaySound", data: { soundId: "alarm" } } },
        { type: "Effect", data: { type: "SetVar", data: { key: "alarm", value: true } } }
    ]
}
```

### Wait

```javascript
{
    type: "Wait",
    data: {
        duration: 2.5  // seconds
    }
}
```

### Conditional

```javascript
{
    type: "If",
    data: {
        condition: world.getVar("global.power.online")
    },
    children: [
        // Then branch
        { type: "Effect", data: { type: "ShowDialog", data: { text: "Power is on" } } },
        // Else branch (optional)
        { type: "Effect", data: { type: "ShowDialog", data: { text: "Power is off" } } }
    ]
}
```

### Player Choice

```javascript
{
    type: "Choice",
    data: {
        prompt: "What do you want to do?",
        options: ["Attack", "Defend", "Run"]
    },
    children: [
        // One plan per option
        { type: "Effect", data: { type: "StartBattle", data: { battleId: "fight" } } },
        { type: "Effect", data: { type: "SetVar", data: { key: "defending", value: true } } },
        { type: "Effect", data: { type: "Trigger", data: { triggerName: "Flee" } } }
    ]
}
```

## Handler Return Values

```javascript
// Return nothing (no changes)
return;

// Return effects only
return {
    effects: [
        { type: "ShowDialog", data: { text: "Hello!" } }
    ]
};

// Return state update only
return {
    state: {
        greeted: true,
        visits: 3
    }
};

// Return plan only
return {
    plan: {
        type: "Seq",
        children: [/* ... */]
    }
};

// Return everything
return {
    effects: [/* ... */],
    plan: { /* ... */ },
    state: { /* ... */ }
};
```

## Utility Functions

```javascript
// Get deterministic game time (milliseconds)
const time = getTime();

// Get deterministic random number [0, 1)
const rand = random();
```

## State Versioning

```javascript
const STATE_VERSION = 2;

function Migrate(state, fromVersion) {
    if (fromVersion === 1) {
        // Migrate from v1 to v2
        return {
            ...state,
            newField: "default_value"
        };
    }
    return state;
}
```

## Common Patterns

### Stateful Counter

```javascript
function OnInteract(self, world, state, event) {
    const count = (state.count || 0) + 1;
    
    return {
        effects: [{
            type: "ShowDialog",
            data: { text: "Interaction #" + count }
        }],
        state: { count: count }
    };
}
```

### Conditional Behavior

```javascript
function OnInteract(self, world, state, event) {
    if (world.getVar("global.quest.complete")) {
        return {
            effects: [{ type: "ShowDialog", data: { text: "Thank you!" } }]
        };
    } else {
        return {
            effects: [{ type: "ShowDialog", data: { text: "Please help me!" } }]
        };
    }
}
```

### Multi-Step Sequence

```javascript
function OnInteract(self, world, state, event) {
    return {
        plan: {
            type: "Seq",
            children: [
                { type: "Effect", data: { type: "ShowDialog", data: { text: "Let me think..." } } },
                { type: "Wait", data: { duration: 2.0 } },
                { type: "Effect", data: { type: "ShowDialog", data: { text: "I have an idea!" } } },
                { type: "Effect", data: { type: "SetVar", data: { key: "quest.active", value: true } } }
            ]
        }
    };
}
```

### Timer-Based Patrol

```javascript
function OnInit(self, world, state) {
    return {
        effects: [{
            type: "StartTimer",
            data: { timerName: "patrol", duration: 3.0 }
        }],
        state: { patrolIndex: 0 }
    };
}

function OnTimer(self, world, state, event) {
    if (event.timerName === "patrol") {
        const points = [
            { x: 10, y: 5 },
            { x: 15, y: 5 },
            { x: 15, y: 10 },
            { x: 10, y: 10 }
        ];
        
        const nextIndex = (state.patrolIndex + 1) % points.length;
        const next = points[nextIndex];
        
        return {
            effects: [
                { type: "MoveTo", data: { x: next.x, y: next.y } },
                { type: "StartTimer", data: { timerName: "patrol", duration: 3.0 } }
            ],
            state: { patrolIndex: nextIndex }
        };
    }
}
```

### Zone-Based Trigger

```javascript
const SUBSCRIPTION = {
    EnterZone: ["danger_zone"]
};

function OnEnterZone(self, world, state, event) {
    if (!state.warned) {
        return {
            effects: [{
                type: "ShowDialog",
                data: { text: "Warning: Danger ahead!" }
            }],
            state: { warned: true }
        };
    }
}
```

### Variable Watcher

```javascript
const SUBSCRIPTION = {
    VarPrefix: ["global.power.*"]
};

function OnFlagChanged(self, world, state, event) {
    if (event.varName === "global.power.grid_online") {
        if (event.newValue) {
            return {
                effects: [{
                    type: "ShowDialog",
                    data: { text: "Power restored!" }
                }]
            };
        }
    }
}
```
