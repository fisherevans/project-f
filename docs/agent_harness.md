# Agent harness

The harness lets a coding agent (or you) drive a running development build of the
game over HTTP: boot straight into a target state, capture rendered frames as
PNGs, inject synthetic controller input, pause and single-step game logic, and
hot-reload YAML content without a rebuild. It exists so the game can be developed
and debugged the same way the rest of the codebase is: make a change, run it,
observe the result, iterate.

The pieces:

- `cmd/development/debugapi` - the HTTP control surface (default `:8091`).
- `internal/game/runtime.Harness` - the loop-coupled state behind time control,
  input injection, and frame capture.
- `cmd/gamectl` - a CLI wrapping the API. This is the front door.

## Why it works this way

Two constraints shape the design:

- **No headless mode.** `gopxl/pixel` needs an OpenGL window on the main thread
  (a hard requirement on macOS). There is no offscreen render path. So the model
  is a long-lived windowed process you drive over HTTP and screenshot over HTTP.
  The window stays visible; you never have to touch it.
- **No Go hot reload.** Compiled code can't be swapped without a restart (plugins
  are a dead end on macOS). Iteration is therefore split into two loops:

  - **Content loop (no rebuild):** edit YAML -> `gamectl reload <kind>` -> drive
    and screenshot. Sub-second. Covers scripts, RPG data, and overlay flows.
  - **Code loop (rebuild):** edit Go -> rebuild and relaunch with a boot
    directive that lands you in the target state -> screenshot. Bounded by
    incremental `go build`, with no manual menu navigation.

The biggest iteration win is keeping one process alive across many iterations and
reloading content into it, not making builds faster.

## Running it

Build and launch a dev build that boots straight into a map:

```
go build -o /tmp/primortal-dev ./cmd/development
PRIMORTAL_BOOT_MAP=hq /tmp/primortal-dev &
go build -o /tmp/gamectl ./cmd/gamectl
/tmp/gamectl wait          # blocks until the loop reports ready
```

Boot directives (read by `cmd/development` at startup):

| Env var | Effect |
|---------|--------|
| `PRIMORTAL_BOOT_MAP` | boot directly into an adventure map (skips the device/title flow) |
| `PRIMORTAL_BOOT_WAYPOINT` | optional spawn waypoint within that map |
| `PRIMORTAL_BOOT_SCENE` | boot a named dev scene (matches the overlay SCENES list) |
| `PRIMORTAL_SAVE` | load a specific save id (default `default`) for a reproducible fixture |
| `PRIMORTAL_ASSETS_DIR` | directory hot-reload reads from (default `./assets`) |

`reset` re-issues whatever boot intent was configured, so it returns to the same
state every time.

## gamectl

`gamectl` talks to `http://localhost:8091` by default (override with `-addr` or
`$GAMECTL_ADDR`).

Status:

```
gamectl health                 # ready, active state, frame counter
gamectl state                  # active state detail (map, player pos)
gamectl entities -player       # the player entity (drop -player for all)
gamectl wait -timeout 60       # block until ready
```

Capture (PNG written to disk; the path is printed so you can open it):

```
gamectl shot frame.png             # 240x160 crisp scene (best for assertions)
gamectl shot -layer scaled out.png # window-resolution, post pixel-grid shader
gamectl shot -compare base.png -tol 0.01 cand.png   # capture and diff vs a baseline
gamectl compare base.png cand.png  # pixel-diff two PNGs, writes a .diff.png
```

Golden-image regression: while the game is `pause`d, nothing advances, so
repeated captures of the same state are byte-identical. `gamectl compare` reports
the fraction of differing pixels, writes a magenta diff image, and exits non-zero
when the diff exceeds `-tol`.

To make a *stepped* sequence reproducible too, turn on determinism with a fixed
seed: `determinism on -seed N` routes the game's shared RNG
(`internal/util/rng`) through a seeded source and re-seeds it on every `reset`.
The canonical golden recipe is:

```
gamectl determinism on -seed 42
gamectl pause
gamectl reset
gamectl step 120 -shot baseline.png
```

Run from two separate launches this produces byte-identical PNGs (verified for
adventure and combat scenes; the same HQ scene differs ~3% across launches
without determinism). Audio randomness is intentionally not seeded - it runs on
the speaker goroutine, where interleaving would break reproducibility.

Time control:

```
gamectl pause                  # freeze game logic (rendering continues)
gamectl run                    # resume (optional -speed 2.0)
gamectl step 30 -shot s.png    # advance exactly 30 fixed-dt frames, then capture
gamectl determinism on -seed 42   # seed the shared RNG (reapplied on reset)
gamectl determinism off
```

`step` is synchronous: it blocks until all frames have been applied, so the state
and the captured frame reflect the step when the command returns. One frame is
1/60s by default; override with `-dt`.

Synthetic input:

```
gamectl tap A                       # momentary press (A|B|Start|Select)
gamectl move Up -frames 40 -shot s.png   # pause, hold a direction, step, capture
gamectl input -a -dir Up -frames 10      # arbitrary combo for N frames
gamectl clearinput                  # cancel held input
```

Input is injected as virtual controller state and merged into the normal control
pipeline, so it behaves exactly like a player holding the button. The standard
frame-by-frame debugging loop is: `pause`, `input ...`, `step 1 -shot a.png`,
`step 1 -shot b.png`, ...

State control:

```
gamectl reset                       # back to the boot state
gamectl save load 1                 # swap the active save to id "1" and reset
gamectl reload scripts              # re-parse assets/scripts/** from disk
gamectl reload rpg                  # re-load skills/primortals
gamectl reload overlays             # re-load overlay flows/rects
gamectl cmd "tp 10 12"              # run any console command, print output
gamectl tp 245 265                  # teleport player
gamectl map hq                      # load a map
```

## Hot reload details

Reload reads from disk (`PRIMORTAL_ASSETS_DIR`, default `./assets`), not the
embedded asset snapshot the binary was built with. So edits to YAML take effect
without rebuilding.

- **scripts**: re-parses every file under `assets/scripts/`, then reloads the
  current map so live entities pick up fresh handler instances. The player's tile
  position is preserved across the reload. If an
  edit fails to parse or validate, the reload is rejected with the error and the
  last known-good embedded scripts are restored, so the running game is never left
  half-loaded.
- **rpg**: clears and reloads the skill and primortal registries.
- **overlays**: reloads overlay flows and named rects.

A reload that panics (e.g. a registry invariant the loader enforces) is recovered
and returned as an error rather than crashing the process.

## HTTP API

`gamectl` is a thin wrapper; the raw endpoints are available for ad-hoc use. All
under `/api/v1/debug`. State-mutating calls that touch the game run on the game
thread via a command queue, so they are safe to call concurrently with the loop.

| Method | Path | Body / query | Purpose |
|--------|------|--------------|---------|
| GET | `/health` | | ready, state, frame |
| GET | `/state` | | active state detail |
| GET | `/entities`, `/entities/{id}` | | entity inspection |
| GET | `/zones`, `/teleports` | | map metadata |
| GET | `/screenshot` | `?layer=scene\|scaled` | PNG of the latest frame |
| GET | `/globals`, `/globals/{key}` | | world/run state |
| POST/DELETE | `/globals/{key}` | `{type,value}` | set/delete a global |
| POST | `/input` | `{a,b,start,select,dir,frames}` | inject input |
| DELETE | `/input` | | clear injected input |
| POST | `/time` | `{mode:run\|pause, speed}` | pause/resume |
| POST | `/step` | `{frames, dt}` | advance N fixed-dt frames (blocking) |
| POST | `/reset` | | re-issue boot intent |
| POST | `/reload/{kind}` | | reload scripts/rpg/overlays |
| POST | `/command` | `{command}` | run a console command |
| POST | `/teleport` | `{target}` or `{x,y}` | teleport player |
| POST | `/map` | `{name, waypoint}` | load a map |
| POST/DELETE | `/highlight` | | overlay highlight control |

## Limitations

- Frame capture and stepping require the dev build; release builds do not expose
  the harness.
- Determinism covers game-thread randomness routed through `internal/util/rng`.
  Anything that consumes the standard library `math/rand` directly (or runs off
  the game thread, like audio) is not covered; add new gameplay randomness via
  the `rng` package to keep it reproducible.
- Swapping the active save at runtime (`gamectl save load <id>`) resets into the
  boot state, so any unsaved progress in the current run is discarded. To pin a
  fixture from the start instead, launch with `PRIMORTAL_SAVE`.
