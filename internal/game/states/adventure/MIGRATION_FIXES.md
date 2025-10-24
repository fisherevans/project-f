# Controller Migration - Bug Fixes

## Issues Found and Fixed

### 1. ❌ Missing Animation Updates
**Problem:** Animations were not being updated after removing `entity.Update()` calls.

**Impact:** Characters would be frozen in a single frame.

**Fix:** Added `updateAnimationsAndLights()` method that:
- Updates player animations with speed multiplier
- Updates NPC animations with speed multiplier  
- Updates DynamicEntity and LightEntity animations/lights
- Resets animations when direction or move state changes

**Code:**
```go
// In main loop
s.updateAnimationsAndLights(timeDelta)

// Separate update methods for Player and NPC
func (s *State) updatePlayerAnimations(timeDelta float64)
func (s *State) updateNPCAnimations(npc *NPC, timeDelta float64)
```

### 2. ❌ Missing Animation Reset Logic
**Problem:** Animations didn't reset when direction or move state changed.

**Impact:** Walking animation would continue from wrong frame when changing direction.

**Fix:** Added tracking fields to Player and NPC:
```go
type Player struct {
    // ...
    lastAnimationDirection input.Direction
    lastAnimationState     MoveState
}
```

Then check and reset in update:
```go
if s.player.lastAnimationDirection != movement.FacingDirection() || 
   s.player.lastAnimationState != movement.MoveState() {
    animation.Reset()
}
```

### 3. ❌ Missing Y-Offset for Character Rendering
**Problem:** Player and NPC `RenderMapLocation()` didn't add the 0.25 Y offset.

**Impact:** Characters would render at wrong vertical position (feet not aligned with tile).

**Fix:**
```go
func (p *Player) RenderMapLocation() pixel.Vec {
    location := p.PreciseMapLocation()
    // Add Y offset for character rendering (feet position)
    return location.Add(pixel.V(0, 0.25))
}
```

### 4. ❌ Missing Import
**Problem:** `adventure_state.go` was missing `input` package import.

**Impact:** Compilation error when referencing `input.Down`.

**Fix:** Added import:
```go
import (
    "fisherevans.com/project/f/internal/game/input"
    // ...
)
```

### 5. ❌ Animation Speed Multiplier
**Problem:** Animations need to update faster when character is moving faster.

**Impact:** Walking animation would look wrong at different speeds.

**Fix:** Multiply timeDelta by movement speed when moving:
```go
if movement.IsMoving() {
    animation.Update(timeDelta * movement.GetCurrentSpeed())
} else {
    animation.Update(timeDelta)
}
```

## Entities Still Using Update()

These entities were NOT migrated to the controller system (they don't move on grid):

### DynamicEntity
- Updates animations and lights per mode
- Used for: torches, lights, props, etc.
- **Still has `Update()` method** - called in `updateAnimationsAndLights()`

### LightEntity  
- Updates light modifiers and animations
- Used for: static light sources
- **Still has `Update()` method** - called in `updateAnimationsAndLights()`

### ShadowMob
- Has custom movement logic (not grid-based)
- **Still has `Update()` method** - called separately in main loop

## Testing Checklist

- [x] Build succeeds
- [ ] Player animations play correctly
- [ ] Player animations reset when changing direction
- [ ] Player animations speed up when running
- [ ] NPC animations play correctly
- [ ] NPC animations reset when changing direction
- [ ] Character feet align with tiles (Y offset)
- [ ] Torches and lights animate
- [ ] Time delta chaining works (multi-tile movements in one frame)
- [ ] Camera follows player smoothly
- [ ] Interactions work (A button)
- [ ] Teleportation works
- [ ] Zone events fire correctly

## Potential Future Issues

### Animation State Desync
If movement state changes outside the controller (via effects), animations might not reset properly.

**Solution:** Ensure all movement changes go through MovementController or trigger animation reset.

### Performance
Iterating all entities twice (once for NPCs, once for other types) could be slow with many entities.

**Solution:** Consider caching entity types or using a component system.

### Missing Light Updates for Player/NPC
Player and NPC don't have lights in the new system (removed from AnimatedMoveableEntity).

**Status:** Intentional - player light was removed. If needed, add back via separate system.

## Summary

✅ **All critical bugs fixed**
✅ **Build succeeds**  
✅ **Animation system restored**
✅ **Rendering offsets preserved**

The migration is functionally complete. Runtime testing needed to verify gameplay.
