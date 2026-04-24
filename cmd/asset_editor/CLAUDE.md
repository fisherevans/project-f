# cmd/asset_editor/CLAUDE.md

Local web tool for browsing, previewing, and editing sprite YAML sidecar
configurations. Covers tilesheets (grid + named aliases + animations),
9-slice frames, plain sprites, and non-atlas overrides. This is the first
tool in what will grow into a broader game dev toolkit - the routing and
layout are designed to support parallel tool domains (entities, events, etc).

## Running

```bash
# Development - two terminals:
GOWORK=off go run ./cmd/asset_editor -dev    # Go API on :8090
cd cmd/asset_editor/frontend && npm run dev  # Vite HMR on :5173

# Production - single binary with embedded frontend:
cd cmd/asset_editor/frontend && npm run build
GOWORK=off go run ./cmd/asset_editor         # serves on :8090
```

`GOWORK=off` is required because the asset editor's Go dependencies
(gorilla/websocket, fsnotify) are not in the workspace go.work file.

Flags: `-dev` enables CORS + Vite reverse proxy, `-port N` changes the
listen port (default 8090), `-assets-dir` overrides the assets path
(defaults to `assets.LocalFolderPath()`).

## Architecture

### Go backend (`cmd/asset_editor/server/`)

Go 1.24 `net/http` with wildcard routing. No framework.

| File | Purpose |
|------|---------|
| `server.go` | Router, middleware, dev/prod mode switching |
| `api_sprites.go` | CRUD handlers for sprite metadata |
| `api_images.go` | Serves PNG files directly from disk |
| `api_scaffold.go` | POST handler that shells out to `cmd/sprite_new` |
| `sprite_service.go` | Directory scanning, YAML read/write, image dimensions |
| `ws.go` | WebSocket hub + fsnotify watcher for live reload |
| `types.go` | JSON response types (wraps `internal/schema` with computed fields) |

**API surface:**

```
GET    /api/v1/sprites              # list all sprites
GET    /api/v1/sprites/{name...}    # full metadata + raw YAML
PUT    /api/v1/sprites/{name...}    # write YAML sidecar (atomic: temp + rename)
DELETE /api/v1/sprites/{name...}    # remove YAML sidecar
POST   /api/v1/sprites/_scaffold    # shell out to cmd/sprite_new
GET    /api/v1/images/{path...}     # serve PNG from disk (not embedded)
GET    /api/v1/ws                   # WebSocket file-change events
```

PNGs are served from disk so Aseprite re-exports appear immediately. The
WebSocket pushes `{type: "change"|"create"|"remove", path: "..."}` events
debounced at 200ms per path.

YAML writes use `yaml.NewEncoder` with `SetIndent(2)` for 2-space
indentation, and write to a temp file then rename for atomicity.

### Frontend (`cmd/asset_editor/frontend/`)

React 19 + TypeScript + Vite + Tailwind CSS v4 + shadcn/ui.

**Key libraries:**
- shadcn/ui components use `@base-ui/react` (not Radix). This affects some
  component APIs - notably `DialogTrigger` uses a `render` prop instead of
  `asChild`, and `Slider` needs `data-horizontal:` variants for sizing.
- TanStack Query for all server state (sprite lists, metadata, mutations).
- React Router for routing.

**Structure:**

```
src/
  api/
    client.ts            # fetch wrapper, apiImageUrl helper
    sprites.ts           # TanStack Query hooks
    websocket.tsx         # WebSocket provider with auto-reconnect + cache invalidation
  types/
    sprites.ts           # TS interfaces mirroring Go JSON types
  layouts/
    AppLayout.tsx         # left nav rail + Outlet
  pages/
    sprites/
      SpriteBrowser.tsx   # directory tree + thumbnail grid + search
      SpriteEditor.tsx    # detail view with type switching, save/delete
  components/
    sprites/
      TilesheetViewer.tsx      # canvas: PNG + grid overlay + zoom + tile selection
      SpriteAliasEditor.tsx    # named sprite CRUD, click-to-assign
      AnimationModal.tsx       # full-screen modal for animation editing
      AnimationConfigForm.tsx  # type selector + sequence/tile fields + playback options
      AnimationPreview.tsx     # live canvas playback + frame strip + speed controls
      FrameEditor.tsx          # 9-slice config + canvas preview
      ScaffoldDialog.tsx       # new tilesheet form (wraps cmd/sprite_new)
    ui/                        # shadcn/ui primitives (button, dialog, input, etc.)
  lib/
    animationEngine.ts   # TypeScript port of Go's animation accumulator
```

**Routing:**
```
/              -> redirect to /sprites
/sprites       -> SpriteBrowser (with ?dir= for directory navigation)
/sprites/:path -> SpriteEditor
```

### Animation engine (`src/lib/animationEngine.ts`)

TypeScript port of the Go animation accumulator from
`internal/game/anim/animation.go:301-329`. `resolveFrames()` expands
h_sequence/v_sequence/tiles into a flat frame list, applies `reverse` then
`pingPong` (appends reversed frames minus endpoints, matching Go's
`for i := len-2; i > 0; i--`). `AnimationPlayer` handles the frame
progression with weight-based duration, jitter, randomize, and repeat.

This must stay in sync with the Go implementation. If the Go animation
logic changes, update `animationEngine.ts` to match.

### Type alignment

Go types live in `internal/schema/`. The frontend mirrors them in
`src/types/sprites.ts`. JSON field names use the Go struct tags. If a
schema type changes, update both the Go struct and the TS interface.

## Common gotchas

- **`GOWORK=off`**: The asset editor's dependencies aren't in the workspace.
  Forgetting this causes build failures.
- **DialogTrigger**: Uses `render` prop, not `asChild`:
  `<DialogTrigger render={<Button>...</Button>} />`
- **Slider width**: Base-ui Slider's `data-horizontal:w-full` overrides
  width classes. Wrap in a sized container div instead.
- **YAML omitempty**: All optional fields in `internal/schema/animation.go`
  and `sprite.go` have `omitempty` yaml tags to avoid writing null/empty
  values.
