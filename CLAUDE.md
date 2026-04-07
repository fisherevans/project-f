# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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

- `primortals.go` — creature definitions and learnable skills
- `skills.go` — combat skill mechanics
- `damage.go` — damage calculation
- `saves.go` / `saves_animech.go` / `saves_primortals.go` — YAML-backed persistence in `game_data/saves/`

### Rendering pipeline

1. State renders to a 240×160 `pixel.Canvas`
2. Canvas is composited to the window at integer scale
3. Pixel-grid GLSL shader overlaid (`internal/game/shaders/`)
4. Optional bloom/effect shaders
5. Debug overlay drawn last

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

### Entry points

- `cmd/development/dev_launcher.go` — dev launcher with hardcoded test scenarios
- `cmd/release/release_launcher.go` — release launcher with panic recovery and zenity error dialogs
