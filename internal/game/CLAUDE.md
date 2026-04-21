# internal/game/CLAUDE.md

State machine, intents, and runtime context. Read `internal/CLAUDE.md` first
for rendering conventions - this file covers only game-specific patterns.

## State interface

Every screen implements `game.State` (defined in `game.go`):

```go
type State interface {
    ClearColor() pixel.RGBA
    OnTick(target pixel.ComposeTarget, targetBounds pixel.Rect, timeDelta float64)
    OnEnter(data any)
    OnExit()
    HandleConsoleInput(string) bool
}
```

Embed `game.BaseState` to get no-op defaults for `OnEnter`, `OnExit`,
`HandleConsoleInput`, and a black `ClearColor`. Override what you need.

The runtime calls `OnTick` once per frame with the scene canvas as target.
States are responsible for clearing their own batches, drawing, and flushing.
They do not own the canvas - the runtime clears it to `ClearColor()` before
`OnTick`.

## Intents and state transitions

States don't construct other states directly. They request a transition by
publishing an intent:

```go
game.SetActiveStateIntent(game.CombatIntent{
    Player:     game.NewCombatPlayer(animech),
    Opponent:   game.NewCombatOpponent(...),
    OnComplete: func(prev game.State, r game.CombatIntentResult) { ... },
})
```

The runtime applies the intent on the next frame (`ApplyIntent` → `OnExit` old
state → `OnEnter` new state). Intent types are mapped to factories in one
central place: `runtime/register_intents.go`. Each line is
`game.RegisterStateFactory(stateconstructor)`, where the constructor has
signature `func(IntentType) game.State`.

Adding a new state:
1. Define the intent struct in `game/intents.go`.
2. Implement `New(intent) game.State` in the state package.
3. Add `game.RegisterStateFactory(mystate.New)` to
   `runtime/register_intents.go`.

`SwapStateIntent{State: ...}` is a passthrough that reuses an existing state
instance - used when returning from menus back to the state that opened them.

## Controls

`game.Controls[*MyState]()` returns the controls instance, but only if
`*MyState` is the active state. Otherwise you get `ControlsNoop`, which reads
as neutral. This is how background states (menu's parent, adventure under
dialogue) naturally stop responding to input.

```go
ctrls := game.Controls[*State]()
if ctrls.ButtonA().JustPressed() { ... }
if ctrls.DPad().JustPressedOrRepeatedDirection() == input.Up { ... }
```

A state that has its own modal input capture (highlighter, dialogue) should
return `ControlsNoop` from its own controls accessor when the modal is
active, mirroring `adventure.State.Controls()`.

## Context singletons

`internal/game/context.go` exposes process-wide accessors. These are
package-level functions, not a struct you pass around:

- `game.CurrentSave() *rpg.GameSave` - active save; persist with `.Save()`.
- `game.TimeElapsed() float64` - seconds since `Initialize`. Use for
  time-based animations that should survive pauses.
- `game.Utils().TimeCycleSin(hz)` - shared oscillator helpers for UI pulses.
- `game.Flags()` - runtime flags; `JustChanged(name)` for edge detection.
- `game.Console()` - dev command console; `IsActive()` gates state input.
- `game.GetAudioSystem().PlayMusic(path, opts)`, `.PlaySFX(name, volume)`,
  `.PlayUI(name)`, `.PlaySoundOnBus(name, bus, volume, opts)`.

## State package layout

Typical per-state layout (see `states/combat/`, `states/adventure/`):

- `state.go` (or `*_state.go`) - the `State` struct, `New`, `OnTick`.
- `registration.go` - `init()` that registers factories.
- `controls.go` - input handling helpers if complex.
- `load.go` - map/asset setup invoked from `New`.
- Feature files (`combat_damage.go`, `entities.go`, ...) - split by subsystem.

Larger states (adventure, combat) have an event/effect bus. When adding to
those, prefer a new `Effect` implementation over ad-hoc logic in `OnTick`.

## Canvases and compositing

Simple states draw everything to `target` (the shared 240x160 scene canvas)
through a single batch. More complex states allocate their own intermediate
canvases for layered effects:

```go
s.sceneCanvas     = opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
s.lightMapCanvas  = opengl.NewCanvas(pixel.R(0, 0, game.GameWidth, game.GameHeight))
```

The `SetComposeMethod(pixel.ComposeMultiply | ComposeScreen | ComposeOver)`
pattern is how layers blend; see the adventure state's lighting pipeline.

The runtime handles the pixel-grid canvas scaling and shader, the debug
overlay, and the dev-mode HUD. Don't re-implement window-level compositing in
a state.

## Save data

`rpg.GameSave` is the persisted root. Load is automatic in `Initialize`;
writes happen when you call `save.Save()`. Don't call `Save` every frame -
save at meaningful boundaries (menu close, combat end, map transition). The
YAML format lives in `game_data/saves/<saveId>.yaml`.

## Audio

`game.GetAudioSystem()` is the entry point. `PlaybackControl` returned from
`PlayMusic`/`PlaySound` is how you fade, pause, or resume later - hold onto
it in your state struct if you need to stop the sound in `OnExit`.
