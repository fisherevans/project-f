# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Maintaining these docs

There are scoped CLAUDE.md files that agents inherit automatically:

- `internal/CLAUDE.md` - rendering, asset, gfx, text, frames, color, shader, and atlas conventions. The source of truth for "what tool do I use to draw X".
- `internal/game/CLAUDE.md` - state/intent/context/controls patterns.
- `assets/CLAUDE.md` - YAML sidecar formats for sprite, frame, and tilesheet assets.
- `cmd/asset_editor/CLAUDE.md` - asset editor web tool architecture, API surface, and frontend conventions.

Treat these as living docs. When you introduce or change a shared pattern that other code should follow, update the relevant CLAUDE.md in the same change:

- New internal helper that should supplant a raw pixel call (new drawing util, new textbox option, new frame utility, new anim loader) - add or update the "what to use for what" guidance.
- Rename or remove an existing helper referenced in a CLAUDE.md - update the reference.
- New asset sidecar key or convention - document it in `assets/CLAUDE.md`.
- New state lifecycle hook, intent pattern, or context singleton - document it in `internal/game/CLAUDE.md`.

If a recommendation in one of these files turns out to be wrong or outdated while you're working, fix it before finishing the task. Don't leave stale guidance in place for the next agent. Conversely, don't churn the docs for one-off changes - these files exist to codify *shared* patterns, not track every edit.

When adding a new shared pattern that doesn't fit any existing CLAUDE.md, decide whether it deserves its own scoped file (new subsystem with enough surface area) or a section in an existing one. Prefer extending an existing file unless the new scope is clearly distinct.

## Commands

**Run (development):**
```
go run ./cmd/development
```
Starts with a state-selector menu for testing specific maps/encounters. Exposes pprof at `localhost:6060`.

**Run (release mode):**
```
go run ./cmd/release
```
Starts from the title screen with log output to `game_data/logs.txt`.

**Build release binary:**
```
go build -ldflags="-s -w" ./cmd/release
```

**Drive the game from an agent (run, screenshot, input, step, hot-reload):**
```
go build -o /tmp/primortal-dev ./cmd/development
PRIMORTAL_BOOT_MAP=hq /tmp/primortal-dev &   # boot straight into a map
go build -o /tmp/gamectl ./cmd/gamectl
/tmp/gamectl wait && /tmp/gamectl shot frame.png
```
The dev build exposes a debug HTTP API on `:8091`; `gamectl` wraps it. You can
boot into a target state, capture frames as PNGs, inject controller input, pause
and single-step game logic, and hot-reload `assets/scripts`, RPG data, and
overlay flows without rebuilding. This is the primary loop for developing and
verifying gameplay changes. Full reference: `docs/agent_harness.md`.

**Scaffold a new sprite sheet (PNG + YAML sidecar + .aseprite):**
```
go run ./cmd/sprite_new \
    -name assets/sprites/overlay/icons \
    -tile-width 16 -tile-height 16 \
    -cols 8 -rows 2 \
    -sprites volume_on,volume_off,gear,fullscreen,fullscreen_exit,scale,reset,-
```
Sprite aliases are positional, row-major. Use `-` or blank to skip a cell. Pass
`-force` to overwrite existing files. The `.aseprite` step runs automatically
if the Aseprite CLI is on PATH or at `/Applications/Aseprite.app`. See
`assets/CLAUDE.md` for what the generated YAML means.

**Run asset editor (visual sprite/animation/frame editor):**
```
# Development (two terminals):
GOWORK=off go run ./cmd/asset_editor -dev    # Go API on :8090
cd cmd/asset_editor/frontend && npm run dev  # Vite HMR on :5173

# Production (single binary):
cd cmd/asset_editor/frontend && npm run build
GOWORK=off go run ./cmd/asset_editor         # serves embedded frontend on :8090
```
Opens a browser automatically. See `cmd/asset_editor/CLAUDE.md` for details.

**Tests:**
```
go test ./...
go test ./internal/game/rpg/...   # run a single package
```

## Architecture

**Tech stack:** Go 1.24, [gopxl/pixel](https://github.com/gopxl/pixel) (OpenGL via pixelgl), [gopxl/beep](https://github.com/gopxl/beep) (audio). Game renders to a 240×160 canvas (GBA resolution) then scales up with a pixel-grid shader.

### State machine

All game screens implement the `State` interface (`internal/game/states.go`). Transitions happen through **Intents** — a state requests an intent, the runtime applies it next frame:

```
AdventureState → requests CombatIntent → runtime creates CombatState
```

Key files:
- `internal/game/intents.go` — intent types and `RegisterStateFactory`
- `internal/game/runtime/runtime.go` — main loop: `ApplyIntent → UpdateControls → Update → renderScene → compositeToWindow`
- `internal/game/context.go` — singleton `Context` (active state, save, controls, debug)

### Resource loading

Assets are loaded lazily via `resources.RunOnceInitialized()` callbacks so startup is non-blocking. All sprite sheets, fonts, audio, and tilemap resources live in `internal/resources/`.

Sprites are configured in YAML sidecar files (e.g. `assets/sprites/.../foo.yaml`) and loaded by `resource_tilesheet_animation.go`.

### Game states

| State | Package |
|-------|---------|
| Adventure (exploration) | `internal/game/states/adventure/` |
| Combat (turn-based) | `internal/game/states/combat/` |
| Xenolog (creature log) | `internal/game/states/xenolog/` |
| Computer (research vessel) | `internal/game/states/computer/` |
| Travel (planet selection) | `internal/game/states/travel/` |
| Menu / Transition / Startup / Title | `internal/game/states/` |

### RPG data layer (`internal/game/rpg/`)

Skills and primortals are defined as YAML files under `assets/rpg/skills/`
and `assets/rpg/primortals/`. Schema types live in `internal/schema/rpg.go`.
At game startup, `rpg.LoadFromFS(assets.FS)` reads the YAML, converts to
runtime types, validates, and populates the `rpg.Skills` and
`rpg.Primortals` registries. The asset editor reads/writes the same YAML
files on disk.

- `loader.go` — YAML loading + schema-to-runtime conversion
- `skills.go` — runtime skill types (`Skill`, `SkillSet`, `SkillId`)
- `skills_ticks.go` — tick types (`SkillTick`, `CombatStance`, `SkillTickEffect`, etc.)
- `primortals.go` — runtime primortal types (`Primortal`, `PrimortalType`, registration)
- `damage.go` — damage calculation
- `saves.go` / `saves_animech.go` / `saves_primortals.go` — YAML-backed persistence in `game_data/saves/`

### Rendering pipeline

1. State renders to a 240×160 `pixel.Canvas`
2. Canvas is composited to the window at integer scale
3. Pixel-grid GLSL shader overlaid (`internal/game/shaders/`)
4. Optional bloom/effect shaders
5. Debug overlay drawn last

### Pixel alignment in hi-res UI

Hi-res layers (settings overlay, debug HUD, corner chrome) render at window
resolution with bitmap fonts and sprite primitives. Every draw position must
be on an integer pixel - subpixel placement makes bitmap text and small
sprites render blurry or drift by a fraction of a pixel, which reads as
jitter. Game-canvas (240×160) code is safe because `gfx.IVec` / `gfx.Moved`
already force integer positions; window-space UI code composes in raw floats
and has to stay disciplined.

The common mistake is dividing an odd value by 2. A few examples:

```go
// basicfont.Face7x13 has LineHeight = 13 (odd).
y := r.Center().Y - txt.LineHeight/2   // off by 0.5 whenever r.H() is even
y := r.Min.Y + (r.H()-13)/2            // off by 0.5 whenever r.H() is odd
dx := 3.0 / 2.0                        // 1.5 — any offset using this inherits the drift
```

Floor or round the final draw position before passing it to `txt.Draw` /
`DrawRect`. In the overlay package, `DrawCtx.drawText(txt, pos)` wraps
`text.Text.Draw` and floors both axes; prefer it over calling `Draw`
directly. When constructing a rect whose `Center()` feeds into a draw
position, prefer even dimensions so `Center()` lands on an integer.

### Animation system

Each sprite sheet (`assets/sprites/**/*.png`) has a YAML sidecar at the same path that declares tile dimensions and named animations. Animations are loaded in Go via `anim.Load(atlas, "path/to/sheet:animation_name")`. Omitting the colon uses the animation named `default`.

**YAML structure:**

```yaml
tilesheet:
  tileWidth: 32    # pixels per tile
  tileHeight: 32

animations:
  my_anim:
    # --- frame source (exactly one of the following) ---

    h_sequence:          # consecutive columns in one row
      row: 1             # required; 1-indexed
      fromColumn: 1      # optional; defaults to 1
      toColumn: 8        # optional; defaults to sheet width
      frameWeights: []   # optional; per-frame duration multipliers

    v_sequence:          # consecutive rows in one column
      column: 1          # required; 1-indexed
      fromRow: 1         # optional; defaults to 1
      toRow: 8           # optional; defaults to sheet height
      frameWeights: []

    tiles:               # arbitrary list of frames
      - row: 2
        column: 5
        weight: 1.0      # duration multiplier; defaults to 1

    # --- playback options ---
    framesPerSecond: 8   # required; 0 treated as 1
    jitterPercent: 0.15  # ±jitter on each frame duration (0–1); default 0
    repeat: true         # loop; default true
    pingPong: true       # append frames in reverse after forward pass
    reverse: true        # reverse frame order before pingPong
    randomize: true      # pick next frame randomly instead of sequentially
```

All row/column indices are **1-indexed**. `pingPong` is applied after `reverse`. The `frameWeights` list length must match the number of frames in the sequence.

### Shared schema (`internal/schema/`)

Pure-data YAML types extracted from `internal/resources/` so both the game
and the asset editor can import them without pulling in OpenGL dependencies.
Contains `SpriteMetadata`, `SpriteTilesheet`, `SpriteTilesheetAnimation`,
`SpriteFrame`, and related types. Also contains `RPGSkill`, `RPGPrimortal`,
and related types for the YAML-backed RPG data under `assets/rpg/`. The
game's resource loaders in `internal/resources/` and `internal/game/rpg/`
import these types rather than defining their own.

### Adventure script system

Event handlers for adventure entities are defined in YAML files under
`assets/scripts/`. Each file declares `handlers` (named event handlers
attached to entities via Tiled map properties) and optional `sequences`
(reusable step lists), `consts` (shared read-only data accessible as
`const.*` in expressions), and `custom_actions` (reusable parameterized
step sequences). Co-locate consts, custom actions, and handlers in the same
file when they're only used by that file's handlers.

Handlers have event hook fields (`on_interact_self`, `on_broadcast`, etc.)
containing rules with optional filters, conditions, and steps (effects).
Hooks default to `first_match` mode (first matching rule wins). Set
`mode: all` to run every matching rule:

```yaml
on_interact_self:
  mode: all
  rules:
    - when: ...
      steps: [...]
    - when: ...
      steps: [...]
```

The legacy list-of-rules format is equivalent to `first_match`. Handlers
can declare a `var` block with initial values for handler-local state,
mutable via the `set_var` step and accessible as `var.*` in expressions.

**Expression engine:** String interpolation uses `expr-lang/expr`
(`github.com/expr-lang/expr`). Expressions in `{{...}}` blocks are
compiled at script load time and evaluated at runtime. The expression
environment provides: `var` (handler state), `global` (world/run state),
`const` (script-defined constants), `save` (game save data), `self`,
`player`, `source`, `prop` (entity properties), `param` (custom action
parameters). Built-in functions: `len`, `min`, `max`, `clamp`, `str`,
`int`, `float`, `rand`, `randf`, `keys`, `values`, `hasKey`.

**Control flow steps:** `if` (conditional with then/else), `switch`
(multi-branch matching), `while` (loop with safety cap), `return` (exit
current scope - custom action or handler rule). Conditions are expr
expressions evaluated at runtime.

**Custom actions:** YAML-defined reusable step sequences with parameters.
Defined in `custom_actions:` blocks, invoked via the `custom_action` step
with `name` and optional `params`. Share the caller's var scope;
recursion capped at depth 10.

The complete schema for all step kinds, named actions, conditions, event
hooks, and template variables is in `internal/schema/script_schema.json`.
A human-readable reference is at `docs/script_reference.md` (regenerate
via `go run ./cmd/gen_script_docs`). Tests in
`internal/game/states/adventure/script_schema_test.go` enforce that the
schema stays in sync with runtime registrations.

Key files:
- `internal/game/states/adventure/script_expr.go` - expression engine (compile, eval, interpolation)
- `internal/game/states/adventure/script_types.go` - YAML parsing types
- `internal/game/states/adventure/script_effects.go` - step kind to Effect conversion
- `internal/game/states/adventure/script_actions.go` - named action/condition registrations
- `internal/game/states/adventure/script_adapter.go` - ScriptHandler (EventHandler implementation)
- `internal/game/states/adventure/script_loader.go` - YAML file loading, consts, custom actions
- `internal/game/states/adventure/effect_control_flow.go` - if/switch/while step effects
- `internal/game/states/adventure/effect_custom_action.go` - custom action invocation

### Entry points

- `cmd/development/dev_launcher.go` - dev launcher with hardcoded test scenarios
- `cmd/release/release_launcher.go` - release launcher with panic recovery and zenity error dialogs
- `cmd/asset_editor/main.go` - asset editor web tool (see `cmd/asset_editor/CLAUDE.md`)
