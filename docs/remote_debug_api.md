# Debug REST API + SPA for Game State Management

## Context

The in-game CommandConsole (`/` key overlay) is limited: text-only, no state browsing, no structured output, hard to use for complex operations. Replacing the console's role with a REST API served from the running game process lets a browser-based SPA provide a richer debug UI - searchable globals, entity maps, teleport buttons, save data inspection, etc. The console itself stays in the game for quick one-off commands; the API is an additional interface, not a removal.

## Architecture

### Separation between game runtime and debug API

All debug API code (HTTP server, route handlers, frontend SPA) lives under `cmd/development/debugapi/`. This keeps `cmd/release` completely unaware of it - no build tags, no conditional imports, no embedded frontend bloating the release binary. The only change to `internal/` is a one-function hook on the runtime `Instance`.

### Thread safety via command queue

The game loop is single-threaded (OpenGL). HTTP handlers run on separate goroutines. No game state is safe to read or write from an HTTP goroutine - even reads can race with the game thread mutating the same struct mid-frame.

The solution is a command queue: every API handler packages its work into a closure and enqueues it onto a buffered channel. The game loop drains this channel once per frame (between `ApplyIntent` and `UpdateControls`), executing each closure on the game thread. The HTTP goroutine blocks on a per-request result channel until its closure completes.

```
HTTP handler goroutine                    Game thread (main loop)
    |                                         |
    |-- Enqueue(closure) -----> channel ----> DrainOnGameThread()
    |        |                                   |
    |   [blocks on result chan]              closure() executes
    |        |                                   |
    |   <--- result <--------- result chan <---- return (jsonBytes, err)
    |
    v write jsonBytes to ResponseWriter
```

Closures must serialize their return data to `[]byte` (JSON) on the game thread before sending it back. The HTTP handler receives raw bytes and writes them to the response - no shared memory crosses the goroutine boundary. No mutexes needed.

Latency: worst case is one frame (~16ms at 60fps). `Enqueue` takes a `context.Context` so HTTP requests with a deadline don't block forever if the game loop hangs; timeout returns HTTP 503.

### State-dependent actions

Different game states expose different capabilities. `adventure.State` has entities, maps, and teleport targets; `combat.State` has combatants and turns; `title.State` has nothing useful. The API handles this at two levels:

**1. Typed endpoints with runtime checks.** State-specific handlers (entities, teleports, map info) call `game.GetActiveState()` inside their closure, type-assert to the expected state (e.g. `*adventure.State`), and return a structured 409 error if the assertion fails. The response includes the actual state type name so the frontend knows what's available:

```go
// Inside a closure running on the game thread
advState, ok := game.GetActiveState().(*adventure.State)
if !ok {
    return nil, &StateError{
        Want: "adventure",
        Got:  reflect.TypeOf(game.GetActiveState()).String(),
    }
}
// ... use advState safely
```

**2. `POST /command` as a universal escape hatch.** This endpoint forwards raw text to `activeState.HandleConsoleInput(cmd)`, which already dispatches per-state (only `adventure.State` implements it; `BaseState` returns false). Every console command that works today works through the API immediately, regardless of which state is active or what typed endpoints exist.

**Frontend adaptation.** `GET /state` returns the active state's type name. The frontend polls or subscribes to this and conditionally shows/hides sections - e.g. the entity inspector only renders when `state.type == "adventure"`. No capability negotiation protocol needed; the frontend just matches on a known set of type strings.

## Files to create/modify

### New: `cmd/development/debugapi/queue.go`

Command queue with two methods:
- `Enqueue(ctx context.Context, fn func() (any, error)) (any, error)` - called from HTTP goroutines, blocks until game thread executes `fn` or context deadline. Returns 503 on timeout.
- `DrainOnGameThread()` - called once per frame from the game loop. Executes all pending closures and sends results back.

Channel buffer size: 64 (more than enough for a dev tool).

### New: `cmd/development/debugapi/server.go`

HTTP server modeled on `cmd/asset_editor/server/server.go`:
- Go stdlib `net/http.ServeMux` with Go 1.22+ wildcard routing
- CORS middleware in dev mode (same pattern as asset editor)
- Constructor takes `*CommandQueue`, `devMode bool`, `frontendFS fs.FS`
- Dev mode: Vite reverse proxy to `:5174` (5173 is asset editor)
- Prod mode: `http.FileServer` over embedded frontend/dist

### New: `cmd/development/debugapi/api_state.go`

```
GET /api/v1/debug/state     - active state type + capabilities (used by frontend to show/hide sections)
GET /api/v1/debug/save      - full GameSave serialized to JSON
```

`/state` response includes `type` (e.g. `"*adventure.State"`), `capabilities` (list of strings like `"entities"`, `"teleport"`, `"map"`), and state-specific fields (map name, player position when in adventure). The capabilities list drives frontend section visibility.

### New: `cmd/development/debugapi/api_globals.go`

```
GET    /api/v1/debug/globals           - all globals, optional ?prefix= filter
GET    /api/v1/debug/globals/{key...}  - single global value
POST   /api/v1/debug/globals/{key...}  - set global: {"type": "int", "value": 42}
DELETE /api/v1/debug/globals/{key...}  - delete global
```

Uses `game.CurrentSave().Globals` interface methods: `KeysWithPrefix`, `Get`, `Set`, `Delete`.

### New: `cmd/development/debugapi/api_commands.go`

```
POST /api/v1/debug/command   - execute raw console command: {"command": "tp npc_1"}
POST /api/v1/debug/teleport  - {"target": "entity_id"} or {"x": 5, "y": 10}
POST /api/v1/debug/map       - {"name": "map1", "waypoint": "default"}
```

`/command` is the universal escape hatch - it routes through `activeState.HandleConsoleInput(cmd)` (line 88 in `context.go`), so any existing console command works through the API. The handler captures console output by recording the buffer length before execution and reading new lines after. `/teleport` and `/map` are convenience wrappers that construct the equivalent console commands (`"tp <target>"`, `"map <name> <waypoint>"`) - they exist so the frontend can offer structured UIs without string-building.

Requires a small addition to `CommandConsole`:

### Modified: `internal/game/console.go`

Add `RunCapture(fn func()) []string` method - records buffer length, runs fn, returns new lines. Runs on game thread so no concurrency issue.

### New: `cmd/development/debugapi/api_entities.go`

```
GET /api/v1/debug/entities        - list entities (id, position, type) - adventure only
GET /api/v1/debug/entities/{id}   - full entity detail (behaviors, metadata, presence)
GET /api/v1/debug/teleports       - list teleport references in current map
```

All three handlers type-assert `game.GetActiveState()` to `*adventure.State` inside their game-thread closure. Returns HTTP 409 with `{"error": "requires adventure state", "active_state": "..."}` when the game is in combat, menus, or any other state. The frontend uses the `/state` endpoint to know when to even show the entities section.

### New: `cmd/development/debugapi/types.go`

JSON response structs: `StateInfo`, `GlobalEntry`, `EntitySnapshot`, `EntityDetail`, `TeleportEntry`, `CommandResponse`.

### Modified: `internal/game/runtime/runtime.go`

Add optional drain hook to `Instance`:

```go
type Instance struct {
    // ...existing fields...
    debugDrain func()
}

func (i *Instance) WithDebugDrain(fn func()) *Instance {
    i.debugDrain = fn
    return i
}
```

In the game loop, call `i.debugDrain()` after `game.ApplyIntent()` and before `game.Update()`:

```go
game.ApplyIntent()
if i.debugDrain != nil {
    i.debugDrain()
}
game.UpdateControls(i.window)
game.Update(i.window, gameDelta)
```

### Modified: `cmd/development/dev_launcher.go`

Wire up the debug API:

```go
queue := debugapi.NewCommandQueue(64)
server := debugapi.NewServer(queue, true, nil)
instance := runtime.NewInstance(...).
    WithDevScenes(devscenes.Scenes()).
    WithDebugDrain(queue.DrainOnGameThread)
go http.ListenAndServe("localhost:8091", server)
opengl.Run(instance.Run)
```

Port 8091 (pprof is 6060, asset editor is 8090).

### New: `cmd/development/debugapi/frontend/`

Same toolchain as `cmd/asset_editor/frontend/`: Vite + React 19 + TanStack Query + Tailwind v4. Vite proxy config points API to `:8091`.

Pages (in priority order):
1. **Overview** - active state, map, player position, save summary, quick-action buttons (teleport, load map, set elythium)
2. **Globals** - searchable table, inline edit, add/delete
3. **Commands** - terminal UI for raw console commands with output display
4. **Entities** - entity list with positions, detail drill-down (behaviors, metadata)

### New: `cmd/development/debugapi/embed.go`

`//go:embed frontend/dist` for production-mode serving. Only compiled when building `cmd/development`.

## Implementation order

**Phase 1 - Core plumbing + API (no frontend)**
1. `queue.go` - CommandQueue
2. `server.go` - HTTP server skeleton
3. `console.go` mod - RunCapture method
4. `runtime.go` mod - WithDebugDrain hook
5. `dev_launcher.go` mod - wire up
6. `api_state.go` - GET /state, GET /save
7. `api_globals.go` - globals CRUD
8. `api_commands.go` - POST /command, /teleport, /map
9. `types.go` - response structs

Testable with curl/Postman immediately after this phase.

**Phase 2 - Adventure-specific endpoints**
10. `api_entities.go` - entity list + detail + teleports

**Phase 3 - Frontend SPA**
11. Frontend scaffold (Vite, React, TanStack Query, Tailwind)
12. API client + types
13. Overview page
14. Globals editor page
15. Commands terminal page
16. Entities page
17. `embed.go`

**Phase 4 - WebSocket (optional, defer)**
18. WebSocket hub for live state-change push notifications
19. Frontend WebSocket provider with cache invalidation

## Verification

1. Start dev build: `go run ./cmd/development`
2. Confirm API responds: `curl http://localhost:8091/api/v1/debug/state`
3. Test globals: `curl -X POST http://localhost:8091/api/v1/debug/globals/test_key -d '{"type":"int","value":42}'`
4. Test command: `curl -X POST http://localhost:8091/api/v1/debug/command -d '{"command":"global list"}'`
5. Test teleport (while in adventure): `curl -X POST http://localhost:8091/api/v1/debug/teleport -d '{"target":"some_entity"}'`
6. Start frontend dev: `cd cmd/development/debugapi/frontend && npm run dev` - opens on :5174
7. Verify all pages load and API calls succeed through Vite proxy
8. Confirm `cmd/release` builds and runs without debug API code
