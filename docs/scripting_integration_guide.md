# Scripting System Integration Guide

This guide shows how to integrate the scripting system with the existing Project F adventure state.

## Step 1: Add Scripting Fields to Adventure State

```go
// In internal/game/states/adventure/adventure_state.go

type State struct {
    game.BaseState
    
    // ... existing fields ...
    
    // Scripting system
    scriptRegistry   *scripting.ScriptRegistry
    scriptDispatcher *scripting.EventDispatcher
    worldVars        *scripting.WorldVarStore
    saveSystem       *scripting.SaveSystem
    debugTools       *scripting.DebugTools
}
```

## Step 2: Initialize Scripting in State Constructor

```go
// In internal/game/states/adventure/adventure_state.go

func New(i game.AdventureIntent) game.State {
    // ... existing initialization ...
    
    // Initialize scripting system
    a.scriptRegistry = scripting.NewScriptRegistry()
    a.worldVars = scripting.NewWorldVarStore()
    a.scriptDispatcher = scripting.NewEventDispatcher(a.scriptRegistry, a.worldVars)
    a.saveSystem = scripting.NewSaveSystem(a.scriptDispatcher, a.worldVars)
    a.debugTools = scripting.NewDebugTools(a.scriptDispatcher, a.worldVars)
    
    // Load scripts from assets
    loader := scripting.NewScriptLoader(a.scriptRegistry, "assets/scripts/examples")
    if err := loader.LoadDirectory("."); err != nil {
        log.Error().Err(err).Msg("Failed to load scripts")
    }
    
    // ... rest of initialization ...
    
    return a
}
```

## Step 3: Extend Entity Interface

```go
// In internal/game/states/adventure/entity.go

type Entity interface {
    // ... existing methods ...
    
    // Scripting support
    GetScriptModule() string
    HasScript() bool
}

type BaseEntity struct {
    // ... existing fields ...
    
    ScriptModule string
}

func (b *BaseEntity) GetScriptModule() string {
    return b.ScriptModule
}

func (b *BaseEntity) HasScript() bool {
    return b.ScriptModule != ""
}
```

## Step 4: Attach Scripts During Entity Creation

```go
// In internal/game/states/adventure/load.go

func createNPCEntity(obj *tiled.Object, layer *tiled.Layer) Entity {
    npc := &NPC{
        // ... existing initialization ...
    }
    
    // Check for script property
    if scriptModule, ok := obj.Properties.GetString("script"); ok {
        npc.ScriptModule = scriptModule
    }
    
    return npc
}

// After entity is added to state
func (s *State) AddEntity(e Entity) bool {
    // ... existing add logic ...
    
    // Attach script if entity has one
    if e.HasScript() {
        initialState := s.getEntityInitialState(e)
        if err := s.scriptDispatcher.AttachScript(string(e.GetEntityId()), e.GetScriptModule(), initialState); err != nil {
            log.Error().Err(err).Str("entity", string(e.GetEntityId())).Msg("Failed to attach script")
        } else {
            // Dispatch Init event
            initEvent := scripting.NewInitEvent(string(e.GetEntityId()))
            if err := s.scriptDispatcher.Dispatch(initEvent); err != nil {
                log.Error().Err(err).Str("entity", string(e.GetEntityId())).Msg("Failed to dispatch init event")
            }
        }
    }
    
    return true
}

func (s *State) getEntityInitialState(e Entity) map[string]interface{} {
    // Extract initial state from entity properties
    // This could come from Tiled custom properties
    return nil
}
```

## Step 5: Dispatch Events from Game Logic

### Interact Events

```go
// In internal/game/states/adventure/entity_npc.go

func (n *NPC) Interact(adv *State, source Entity) {
    // Dispatch to scripting system first
    if n.HasScript() {
        event := scripting.NewInteractEvent(string(n.Id), string(source.GetEntityId()))
        if err := adv.scriptDispatcher.Dispatch(event); err != nil {
            log.Error().Err(err).Msg("Script interaction failed")
        }
        
        // Check if script handled it (could set a flag)
        // If not, fall back to default behavior
    }
    
    // Default behavior
    if n.Talking {
        return
    }
    // ... existing logic ...
}
```

### Zone Events

```go
// In internal/game/states/adventure/movement.go

func (s *State) onEntityEnterZone(entity Entity, zoneName string) {
    if entity.HasScript() {
        event := scripting.NewEnterZoneEvent(string(entity.GetEntityId()), zoneName)
        if err := s.scriptDispatcher.Dispatch(event); err != nil {
            log.Error().Err(err).Msg("Failed to dispatch zone enter event")
        }
    }
}

func (s *State) onEntityLeaveZone(entity Entity, zoneName string) {
    if entity.HasScript() {
        event := scripting.NewLeaveZoneEvent(string(entity.GetEntityId()), zoneName)
        if err := s.scriptDispatcher.Dispatch(event); err != nil {
            log.Error().Err(err).Msg("Failed to dispatch zone leave event")
        }
    }
}
```

### Timer Events

```go
// Create a timer manager in the adventure state

type TimerManager struct {
    timers map[string]*Timer
}

type Timer struct {
    EntityID  string
    Name      string
    Duration  float64
    Remaining float64
}

func (s *State) updateTimers(timeDelta float64) {
    for key, timer := range s.timers {
        timer.Remaining -= timeDelta
        if timer.Remaining <= 0 {
            // Timer expired
            event := scripting.NewTimerEvent(timer.EntityID, timer.Name)
            if err := s.scriptDispatcher.Dispatch(event); err != nil {
                log.Error().Err(err).Msg("Failed to dispatch timer event")
            }
            delete(s.timers, key)
        }
    }
}
```

## Step 6: Apply Effects to Game World

```go
// Create an effect handler that bridges scripting effects to game systems

type EffectHandler struct {
    state *State
}

func (h *EffectHandler) ApplyEffect(effect scripting.Effect) error {
    switch effect.Type {
    case scripting.EffectTypeMoveTo:
        return h.applyMoveTo(effect)
    case scripting.EffectTypeOpenDoor:
        return h.applyOpenDoor(effect)
    case scripting.EffectTypeCloseDoor:
        return h.applyCloseDoor(effect)
    case scripting.EffectTypeShowDialog:
        return h.applyShowDialog(effect)
    case scripting.EffectTypePlayMusic:
        return h.applyPlayMusic(effect)
    case scripting.EffectTypePlaySound:
        return h.applyPlaySound(effect)
    case scripting.EffectTypeStartTimer:
        return h.applyStartTimer(effect)
    case scripting.EffectTypeStartBattle:
        return h.applyStartBattle(effect)
    // ... other effect types ...
    }
    return nil
}

func (h *EffectHandler) applyMoveTo(effect scripting.Effect) error {
    x := effect.Data["x"].(float64)
    y := effect.Data["y"].(float64)
    
    entity, ok := h.state.entities[EntityId(effect.EntityID)]
    if !ok {
        return fmt.Errorf("entity not found: %s", effect.EntityID)
    }
    
    // Trigger movement (would need to adapt to your movement system)
    if moveable, ok := entity.(*AnimatedMoveableEntity); ok {
        target := MapLocation{X: int(x), Y: int(y)}
        moveable.TriggerMovement(h.state, target, MoveStateWalking)
    }
    
    return nil
}

func (h *EffectHandler) applyShowDialog(effect scripting.Effect) error {
    text := effect.Data["text"].(string)
    speaker := ""
    if s, ok := effect.Data["speaker"].(string); ok {
        speaker = s
    }
    
    // Add to dialogue system
    h.state.dialogues.Show(text, speaker)
    
    return nil
}

func (h *EffectHandler) applyOpenDoor(effect scripting.Effect) error {
    doorID := effect.Data["doorId"].(string)
    
    // Find and open door entity
    if door, ok := h.state.entities[EntityId(doorID)]; ok {
        // Assuming you have a Door entity type
        // door.Open()
    }
    
    return nil
}

func (h *EffectHandler) applyStartTimer(effect scripting.Effect) error {
    timerName := effect.Data["timerName"].(string)
    duration := effect.Data["duration"].(float64)
    
    timer := &Timer{
        EntityID:  effect.EntityID,
        Name:      timerName,
        Duration:  duration,
        Remaining: duration,
    }
    
    key := fmt.Sprintf("%s_%s", effect.EntityID, timerName)
    h.state.timers[key] = timer
    
    return nil
}
```

## Step 7: Update Game Loop

```go
// In internal/game/states/adventure/adventure_state.go

func (s *State) OnTick(target *shaders.Canvas, targetBounds pixel.Rect, timeDelta float64) {
    // ... existing update logic ...
    
    // Update timers
    s.updateTimers(timeDelta)
    
    // Update plan runner
    if err := s.scriptDispatcher.Update(timeDelta); err != nil {
        log.Error().Err(err).Msg("Script update failed")
    }
    
    // ... rest of update ...
}
```

## Step 8: Save/Load Integration

```go
// In your save/load system

func (s *State) Save() error {
    // ... existing save logic ...
    
    // Save scripting state
    scriptSave, err := s.saveSystem.Save()
    if err != nil {
        return err
    }
    
    scriptData, err := s.saveSystem.SerializeToJSON(scriptSave)
    if err != nil {
        return err
    }
    
    // Write to save file
    os.WriteFile("save_scripts.json", scriptData, 0644)
    
    return nil
}

func (s *State) Load() error {
    // ... existing load logic ...
    
    // Load scripting state
    scriptData, err := os.ReadFile("save_scripts.json")
    if err != nil {
        return err
    }
    
    scriptSave, err := s.saveSystem.DeserializeFromJSON(scriptData)
    if err != nil {
        return err
    }
    
    if err := s.saveSystem.Load(scriptSave); err != nil {
        return err
    }
    
    return nil
}
```

## Step 9: Debug UI Integration

```go
// Add debug overlay to show scripting state

func (s *State) renderDebugScripting(target pixel.Target) {
    if !game.DebugEnabled {
        return
    }
    
    // Show world variables
    vars := s.worldVars.GetAll()
    y := 100.0
    for key, value := range vars {
        text := fmt.Sprintf("%s = %v", key, value)
        // Render text at y position
        y += 20
    }
    
    // Show active plans
    plans := s.scriptDispatcher.GetPlanRunner().GetActivePlans()
    for planID, cursor := range plans {
        text := fmt.Sprintf("Plan %s: %s", planID, cursor.EntityID)
        // Render text
    }
}
```

## Step 10: Tiled Integration

In Tiled, add custom properties to entities:

```xml
<object id="1" name="guard_001" type="npc" x="320" y="240">
  <properties>
    <property name="script" value="npc_gatekeeper"/>
    <property name="scriptState" value='{"greeted": false}'/>
  </properties>
</object>
```

## Example: Complete NPC with Script

### Tiled Object
```xml
<object id="5" name="elder_npc" type="npc" x="480" y="320">
  <properties>
    <property name="script" value="quest_npc"/>
    <property name="sprite" value="elder"/>
  </properties>
</object>
```

### Script File (assets/scripts/examples/quest_npc.js)
Already created in previous steps.

### Loading Code
```go
func loadNPC(obj *tiled.Object) Entity {
    npc := &NPC{
        BaseEntity: BaseEntity{
            Id:           EntityId(obj.Name),
            ScriptModule: obj.Properties.GetString("script"),
        },
        // ... other properties ...
    }
    return npc
}
```

## Testing the Integration

```go
func TestScriptingIntegration(t *testing.T) {
    // Create adventure state
    state := New(game.AdventureIntent{MapName: "test_map"})
    
    // Create NPC with script
    npc := &NPC{
        BaseEntity: BaseEntity{
            Id:           "test_npc",
            ScriptModule: "npc_gatekeeper",
        },
    }
    
    state.AddEntity(npc)
    
    // Interact with NPC
    player := state.player
    npc.Interact(state, player)
    
    // Verify world vars were set
    visits := state.worldVars.GetInt("global.test.interactions")
    if visits != 1 {
        t.Errorf("Expected 1 interaction, got %d", visits)
    }
}
```

## Performance Considerations

1. **Script Caching**: Scripts are compiled once and cached by hash
2. **Event Filtering**: Only entities with matching subscriptions receive events
3. **Lazy Execution**: Plans execute incrementally, not all at once
4. **Batch Dispatching**: Use `BatchDispatcher` for multiple events

## Debugging Tips

1. Enable event logging:
   ```go
   s.debugTools.PrintEventLog(10)
   ```

2. Inspect world variables:
   ```go
   s.debugTools.PrintWorldVars()
   ```

3. View entity states:
   ```go
   s.debugTools.PrintEntityStates()
   ```

4. Export full state:
   ```go
   jsonData, _ := s.debugTools.ExportToJSON()
   os.WriteFile("debug_state.json", jsonData, 0644)
   ```

## Next Steps

1. Create more example scripts for your specific game mechanics
2. Add custom effect types for game-specific actions
3. Implement hot-reload for development
4. Add script profiling to identify performance bottlenecks
5. Create visual debugging tools in the game UI
