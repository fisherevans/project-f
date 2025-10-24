# Controller-Based Architecture

## Overview

The adventure state has been refactored from an entity-centric update model to a **controller-based architecture** that separates concerns:

- **Movement Controller**: Handles tile-to-tile physics and time delta chaining
- **Behavior System**: Manages entity-specific logic (AI, input handling)
- **Entities**: Pure data + rendering, no update logic

## Key Components

### 1. Movement Controller (`controller_movement.go`)

Centralizes all movement logic for entities moving between tiles.

**Features:**
- ✅ **Time delta chaining**: When movement completes mid-frame, remaining time is used for next movement
- ✅ **Tile transition handling**: Manages entity registration at 50% progression
- ✅ **Zone event emission**: Automatically fires zone enter/exit events
- ✅ **Interpolated positioning**: Smooth rendering between tiles with easing

**Key Methods:**
```go
func (mc *MovementController) Update(s *State, timeDelta float64)
func (mc *MovementController) TriggerMovement(s *State, entityId EntityId, target MapLocation, moveState MoveState) bool
func (mc *MovementController) TeleportTo(s *State, entityId EntityId, location MapLocation) bool
```

**MovementState per entity:**
- Current/target location
- Move progression (0.0 to 1.0)
- Facing direction
- Move speeds per state (walking, running, dashing)
- Constant movement tracking (for easing)

### 2. Behavior System (`controller_behavior.go`)

Manages entity-specific update logic without coupling to movement.

**EntityBehavior Interface:**
```go
type EntityBehavior interface {
    Update(s *State, timeDelta float64)
    OnMovementComplete(s *State, movement *MovementState) bool
}
```

**Implementations:**

#### PlayerBehavior (`behavior_player.go`)
- Handles input (WASD, running, dashing)
- Intent tracking with 75ms delay before movement
- Speed changes during movement
- Interaction handling (A button)
- Does NOT auto-continue movement

#### NPCBehavior (`behavior_npc.go`)
- Random wandering with idle periods
- Horizontal-only movement option
- Configurable speeds
- Faces talking target when in conversation
- Does NOT auto-continue movement

### 3. Simplified Entity Interface

**Before (bloated):**
```go
type Entity interface {
    Move(adv *State, timeDelta float64) float64
    Update(adv *State, timeDelta float64)
    IsMoving() bool
    TeleportTo(s *State, location MapLocation) bool
    // ... 15+ methods
}
```

**After (focused):**
```go
type Entity interface {
    events.EntityContext
    Passable
    
    // Location (delegates to MovementState)
    PreciseMapLocation() pixel.Vec
    Location() MapLocation
    
    // State
    IsMoving() bool
    TeleportTo(s *State, location MapLocation) bool
    
    // Rendering
    RenderScene(target pixel.Target, matrix pixel.Matrix)
    RenderLight(target pixel.Target, matrix pixel.Matrix)
}
```

## Main Update Loop

**Before:**
```go
for _, entity := range s.entities {
    remaining := timeDelta
    for remaining > 0 {
        nextRemaining := entity.Move(s, remaining)
        elapsed := remaining - nextRemaining
        entity.Update(s, elapsed)
        remaining = nextRemaining
    }
}
```

**After:**
```go
// 1. Update behaviors (input handling, AI decisions)
for _, behavior := range s.behaviors {
    behavior.Update(s, timeDelta)
}

// 2. Update movement controller (handles all tile transitions)
s.movementController.Update(s, timeDelta)

// 3. Update other systems
s.timers.Update(timeDelta, s.eventDispatcher, s)

// 4. Process effects
s.processEffects(...)
```

## Time Delta Chaining

The movement controller properly handles time delta chaining to prevent dropped frames:

```go
func (mc *MovementController) Update(s *State, timeDelta float64) {
    for id, state := range mc.states {
        remaining := timeDelta
        for remaining > 0 {
            elapsed := mc.updateSingleMovement(s, state, remaining)
            remaining -= elapsed
            
            // If movement completed with time left, check if entity wants to continue
            if remaining > 0 && state.moveState == MoveStateIdle {
                if behavior, ok := s.behaviors[id]; ok {
                    if !behavior.OnMovementComplete(s, state) {
                        break // No more movement this frame
                    }
                }
            }
        }
    }
}
```

**Example:** If a movement takes 0.8 seconds and completes 0.3 seconds into a frame, the remaining 0.7 seconds is available for the next movement in the same frame.

## Entity Registration

Entities with movement/behavior are registered via helper functions:

```go
// NPCs (in entity constructors)
speeds := map[MoveState]float64{
    MoveStateWalking: walkSpeed,
}
movementState := s.movementController.Register(entityId, location, speeds)
npc.SetMovementState(movementState)

behavior := NewNPCBehavior(entityId, doesMove, horizOnly)
s.behaviors[entityId] = behavior

// Player (in load.go)
speeds := map[MoveState]float64{
    MoveStateWalking: characterSpeed,
    MoveStateRunning: characterSpeed * 1.75,
    MoveStateDashing: characterSpeed * 1.5,
}
movementState := a.movementController.Register(entityId, location, speeds)
a.player.SetMovementState(movementState)

behavior := NewPlayerBehavior(entityId)
a.behaviors[entityId] = behavior
```

## Benefits

### ✅ Separation of Concerns
- Movement physics: MovementController
- Entity logic: Behaviors
- Rendering: Entities
- No cross-contamination

### ✅ Simplified Entity Interface
- Entities are data + rendering
- No complex update loops
- Easy to add new entity types

### ✅ Testable
- Movement controller is pure logic
- Behaviors can be tested independently
- No tight coupling to State

### ✅ Performance
- Single loop over moving entities
- No virtual dispatch per entity
- Cache-friendly data layout

### ✅ Time Delta Accuracy
- No dropped frames
- Smooth multi-tile movements
- Proper frame time accounting

### ✅ Extensible
- Easy to add new behaviors
- Movement logic is centralized
- Clear extension points

## Migration Notes

### Entity Constructor Signature Changed
All entity constructors now receive `s *State` as first parameter:

```go
// Before
func(entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler)

// After
func(s *State, entityId EntityId, location MapLocation, mapEntity *resources.Entity) (Entity, events.EventHandler)
```

This allows constructors to register movement and behavior directly.

### Entities Store MovementState Reference
Entities that move store a reference to their MovementState:

```go
type Player struct {
    BaseEntity
    Passable
    Animations      map[MoveState]map[input.Direction]*anim.AnimatedSprite
    movementState   *MovementState  // ← Reference to controller state
}

func (p *Player) SetMovementState(state *MovementState) {
    p.movementState = state
}
```

### Removed Entity Methods
- `Move(adv *State, timeDelta float64) float64` - Now in MovementController
- `Update(adv *State, timeDelta float64)` - Now in Behaviors
- `Interact(adv *State, source Entity)` - Replaced by event system

## Future Enhancements

### Potential Additions
- **CombatBehavior**: For combat-specific AI
- **PatrolBehavior**: For guard NPCs with waypoints
- **FlockingBehavior**: For group movement
- **PhysicsController**: For non-grid movement (projectiles, etc.)
- **AnimationController**: Centralized animation state management

### Optimization Opportunities
- Spatial partitioning for movement queries
- Behavior pooling for common patterns
- Movement prediction for smooth camera following
- Parallel behavior updates (if needed)

## Files Modified/Created

### Created
- `controller_movement.go` - Movement controller
- `controller_behavior.go` - Behavior system
- `behavior_player.go` - Player behavior
- `behavior_npc.go` - NPC behavior
- `entity_helpers.go` - Registration helpers

### Modified
- `adventure_state.go` - Added controllers, updated main loop
- `entity.go` - Simplified Entity interface
- `entity_player.go` - Removed update logic, delegates to MovementState
- `entity_npc.go` - Removed update logic, delegates to MovementState
- `entity_constructors.go` - Updated signature to include State
- `load.go` - Updated player registration
- `effect_executor.go` - Updated to use movement controller
- All entity constructors - Updated signatures

### Deleted Methods
- `MoveableEntity.Move()` - Logic moved to MovementController
- `MoveableEntity.Update()` - No longer needed
- `Player.Update()` - Logic moved to PlayerBehavior
- `NPC.Update()` - Logic moved to NPCBehavior
- `Entity.Interact()` - Replaced by event system
