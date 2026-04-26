# cmd/asset_editor/CLAUDE.md

Local web tool for browsing, previewing, and editing game asset
configurations. Covers sprite YAML sidecars (tilesheets, animations,
9-slice frames, non-atlas overrides) and audio files (gain control,
playback preview, resource name reference).

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

## E2E tests (Playwright)

```bash
cd cmd/asset_editor/frontend

# Run all tests (starts Go backend + Vite automatically):
npm run test:e2e

# Interactive UI mode (pick/debug individual tests):
npm run test:e2e:ui

# Headed mode (watch the browser):
npm run test:e2e:headed
```

Tests live in `e2e/` and require both the Go backend (:8090) and Vite
(:5173). The Playwright config (`playwright.config.ts`) starts both
servers automatically via `webServer`. If they're already running,
Playwright reuses them (`reuseExistingServer: true`).

Test files:
- `navigation.spec.ts` - nav rail links, root redirect
- `scripts.spec.ts` - script browser listing, editor tabs, handler selection
- `expression-api.spec.ts` - direct API tests for `POST /api/v1/scripts/_validate-expr`
- `expression-editor.spec.ts` - CodeMirror expression input: autocomplete, linting, help modal

Playwright runs Chromium only. Screenshots and traces are captured on
failure (saved to `test-results/`). Add new test files to `e2e/` -
they're auto-discovered.

When writing new e2e tests:
- Use `page.locator(".cm-editor")` / `.cm-content` / `.cm-tooltip-autocomplete` for CodeMirror elements
- Expression lint errors render as `.cm-lintRange-error` (wait up to 5s for debounce + network)
- The step kind picker opens as a dialog - search for step names with the search input
- API tests can use `request.post()` directly against `http://localhost:8090`

## Architecture

### Go backend (`cmd/asset_editor/server/`)

Go 1.24 `net/http` with wildcard routing. No framework.

| File | Purpose |
|------|---------|
| `server.go` | Router, middleware, dev/prod mode switching |
| `api_sprites.go` | CRUD handlers for sprite metadata |
| `api_audio.go` | CRUD handlers for audio metadata |
| `api_images.go` | Serves PNG files directly from disk |
| `api_scaffold.go` | POST handler that shells out to `cmd/sprite_new` |
| `sprite_service.go` | Directory scanning, YAML read/write, image dimensions |
| `audio_service.go` | Audio directory scanning, YAML sidecar read/write |
| `script_service.go` | Script YAML file listing, read/write |
| `api_scripts.go` | CRUD handlers for script files + schema endpoint |
| `api_expr_validate.go` | POST endpoint to validate expr expressions via expr.Compile |
| `api_rpg.go` | CRUD handlers for RPG data (skills, primortals, combat) |
| `rpg_service.go` | RPG YAML file scanning, read/write under `assets/rpg/` |
| `api_saves.go` | CRUD handlers for game save files |
| `save_service.go` | Save file scanning, raw YAML read/write under `game_data/saves/` |
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

GET    /api/v1/audio                # list all audio files
GET    /api/v1/audio/{name...}      # full metadata + resource name + raw YAML
PUT    /api/v1/audio/{name...}      # write/remove YAML sidecar (gain config)
GET    /api/v1/audio-files/{path...}# serve audio file from disk (wav/mp3/ogg/flac)

POST   /api/v1/scripts/_validate-expr # validate an expression via expr.Compile (returns {valid, error?})
GET    /api/v1/script-schema        # complete script system schema (step kinds, actions, conditions, hooks)
GET    /api/v1/scripts              # list all script files with handler/sequence names
GET    /api/v1/scripts/{name...}    # full script detail + raw YAML
PUT    /api/v1/scripts/{name...}    # write script YAML (atomic: temp + rename)
DELETE /api/v1/scripts/{name...}    # delete script file

GET    /api/v1/rpg/skills           # list all skills (from YAML under assets/rpg/skills/)
GET    /api/v1/rpg/skills/{id}      # skill detail with tick timeline
PUT    /api/v1/rpg/skills/{id}      # write skill YAML (atomic: temp + rename)
DELETE /api/v1/rpg/skills/{id}      # delete skill YAML
GET    /api/v1/rpg/primortals       # list all primortals (from YAML under assets/rpg/primortals/)
GET    /api/v1/rpg/primortals/{type}# primortal detail with skills + archetypes
PUT    /api/v1/rpg/primortals/{type}# write primortal YAML (atomic: temp + rename)
DELETE /api/v1/rpg/primortals/{type}# delete primortal YAML
GET    /api/v1/rpg/combat           # status effects + combat stances reference

GET    /api/v1/saves                # list all saves (from game_data/saves/)
GET    /api/v1/saves/{id}           # save detail + raw YAML
PUT    /api/v1/saves/{id}           # write save YAML (raw, validated as parseable YAML)
DELETE /api/v1/saves/{id}           # delete save file
POST   /api/v1/saves/{id}/clone     # duplicate save with new ID

GET    /api/v1/ws                   # WebSocket file-change events
```

PNGs and audio files are served from disk so re-exports appear immediately.
The WebSocket pushes `{type: "change"|"create"|"remove", path: "..."}` events
debounced at 200ms per path, watching `sprites/`, `audio/`, and `scripts/` directories.

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
    client.ts            # fetch wrapper, apiImageUrl/apiAudioUrl helpers
    sprites.ts           # TanStack Query hooks for sprites
    audio.ts             # TanStack Query hooks for audio
    scripts.ts           # TanStack Query hooks for scripts + schema
    rpg.ts               # TanStack Query hooks for RPG data (skills, primortals, combat)
    saves.ts             # TanStack Query hooks for game saves
    websocket.tsx         # WebSocket provider with auto-reconnect + cache invalidation
  types/
    sprites.ts           # TS interfaces mirroring Go sprite JSON types
    audio.ts             # TS interfaces mirroring Go audio JSON types
    scripts.ts           # TS interfaces mirroring Go script schema types
    rpg.ts               # TS interfaces for RPG data (skills, primortals, combat)
    saves.ts             # TS interfaces for game save data
  layouts/
    AppLayout.tsx         # left nav rail (Sprites, Audio, Scripts, RPG, Saves) + Outlet
  pages/
    sprites/
      SpriteBrowser.tsx   # directory tree + thumbnail grid + search
      SpriteEditor.tsx    # detail view with type switching, save/delete
    audio/
      AudioBrowser.tsx    # directory tree + table list + inline playback
      AudioEditor.tsx     # detail view with player, gain slider, path info
    scripts/
      ScriptBrowser.tsx   # script file listing with handler names
      ScriptEditor.tsx    # structured editor + raw YAML tabs, handler list + detail
    rpg/
      RpgLayout.tsx       # sub-nav wrapper (Skills, Primortals, Combat)
      RpgNav.tsx          # tab navigation for RPG sub-pages
      SkillBrowser.tsx    # skill table + editor panel with tick timeline, save/delete
      PrimortalBrowser.tsx # primortal table + editor panel with skill tree + archetypes, save/delete
      CombatBrowser.tsx   # status effects + combat stances reference cards
    saves/
      SaveBrowser.tsx    # save file listing with clone/delete
      SaveEditor.tsx     # raw YAML editor for save files
  components/
    sprites/
      TilesheetViewer.tsx      # canvas: PNG + grid overlay + zoom + tile selection
      SpriteAliasEditor.tsx    # named sprite CRUD, click-to-assign
      AnimationModal.tsx       # full-screen modal for animation editing
      AnimationConfigForm.tsx  # type selector + sequence/tile fields + playback options
      AnimationPreview.tsx     # live canvas playback + frame strip + speed controls
      FrameEditor.tsx          # 9-slice config + canvas preview
      ScaffoldDialog.tsx       # new tilesheet form (wraps cmd/sprite_new)
    scripts/
      HandlerList.tsx          # left panel: handler names, add/rename/delete + sequences/consts/custom actions
      HandlerDetail.tsx        # right panel: var section + hook sections for selected handler
      HookSection.tsx          # collapsible section per event hook with rules
      RuleEditor.tsx           # single rule: filter + condition + set_state + steps
      StepList.tsx             # ordered step list with add/remove/reorder
      StepEditor.tsx           # single step: kind badge + inline/expanded params + switch cases
      StepKindPicker.tsx       # schema-driven step kind selector grouped by category
      StepParamForm.tsx        # dynamic form from ParamDef[] (inputs, selects, toggles, expr inputs)
      ConditionEditor.tsx      # recursive condition tree builder (all/any/not/expr/leaf)
      ExprContext.tsx           # React context for expression editor (var keys, const keys, help modal)
      ExpressionHelpModal.tsx   # tabbed modal: env vars, functions, operators, examples
      inputs/
        ExpressionInput.tsx    # CodeMirror 6 editor with syntax highlighting, autocomplete, server-side linting
    ui/                        # shadcn/ui primitives (button, dialog, input, etc.)
  lib/
    animationEngine.ts   # TypeScript port of Go's animation accumulator
    scriptUtils.ts       # YAML parse/stringify, step/condition conversion, tree helpers
```

**Routing:**
```
/              -> redirect to /sprites
/sprites       -> SpriteBrowser (with ?dir= for directory navigation)
/sprites/:path -> SpriteEditor
/audio         -> AudioBrowser (with ?dir= for directory navigation)
/audio/:path   -> AudioEditor (gain control, playback, resource name display)
/scripts       -> ScriptBrowser (file listing with handler names)
/scripts/:path -> ScriptEditor (YAML editor + schema reference)
/rpg           -> SkillBrowser (editable, from YAML under assets/rpg/skills/)
/rpg/primortals -> PrimortalBrowser
/rpg/combat    -> CombatBrowser (statuses + stances reference)
/saves         -> SaveBrowser (list + clone/delete from game_data/saves/)
/saves/:id     -> SaveEditor (raw YAML editor with field reference)
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

### Script editor architecture

The script editor uses client-side YAML parsing (the `yaml` npm package).
The backend serves raw YAML strings; the frontend parses them into a typed
tree (`ParsedScript > HandlerDef > RuleDef > StepNode/ConditionNode`),
renders editable components, and serializes back to YAML on save.

Two editing modes share state via bidirectional sync:
- **Structured tab** - split layout with handler list (left) and handler
  detail (right). Components nest: HandlerDetail > HookSection > RuleEditor
  > StepList/ConditionEditor.
- **Raw YAML tab** - plain textarea, same as before.

Switching tabs converts between representations. Parse errors block the
switch to structured mode.

`scriptUtils.ts` contains all parse/serialize logic. Sub-steps within
container steps (focused_sequence effects, teleport_player interstitial,
if/switch/while control flow) are kept as raw YAML objects in
`StepNode.params`. The `getStepSubSteps()` and `setStepSubSteps()`
helpers convert them to/from `StepNode[]` at edit boundaries.

**Handler variables:** HandlerDetail shows a collapsible "Variables"
section for editing `var` defaults. Values are typed (string, number,
bool) based on input parsing.

**Expression-aware editing:** Steps that accept expressions (`if.when`,
`while.when`, `switch.on`, `set_var.value`) use a CodeMirror-based
`ExpressionInput` component with syntax highlighting (token-based:
keywords, functions, env roots, strings, numbers, operators),
autocomplete (env variables, functions, handler var keys, const keys,
save paths), and server-side linting via `POST /api/v1/scripts/_validate-expr`
(debounced 500ms, calls `expr.Compile` on the backend). The `expr`
condition type uses the same input. A help modal (`ExpressionHelpModal`)
documents the full expression environment, functions, operators, and
examples - accessed via an "expr reference" link next to expression fields.

Expression context (handler var keys, const keys) flows via
`ExprContext.tsx` React context, provided at the structured editor root
in ScriptEditor.

**Control flow rendering:** `if` and `while` steps render their sub-step
branches (then/else, steps) inline. `switch` has a dedicated cases
editor with per-case value expression inputs and step lists.

**Custom actions:** The `custom_action` step renders with a dynamic
key-value map editor for the `with` parameter (expression map).

**Left panel sections:** HandlerList shows handlers (editable),
sequences, custom actions, and constants defined in the file.

### Type alignment

Go types live in `internal/schema/`. The frontend mirrors them in
`src/types/sprites.ts` and `src/types/audio.ts`. JSON field names use
the Go struct tags. If a schema type changes, update both the Go struct
and the TS interface.

## Design system

Light mode only. Flat colors, no gradients. Compact density with clear
section hierarchy. Power-user focus - information density over whitespace,
but organized with accents and structure so data doesn't read as a wall.

### Font

Plus Jakarta Sans (variable weight). Imported via
`@fontsource-variable/plus-jakarta-sans` in `index.css`.

| Role | Weight | Size |
|------|--------|------|
| Page header | 700 | `text-base` (16px) |
| Section header | 600 | `text-sm` (14px) |
| Body / form labels | 500 | `text-xs` (12px) |
| Metadata / badges | 500 | `text-[11px]` or `text-[10px]` |
| Code values, YAML, paths | mono 400 | `text-xs font-mono` |

Monospace uses the system stack (`font-mono`). Used for YAML content,
entity IDs, script handler names, file paths, and any value the user
might copy into code.

### Colors

All colors are OKLch, defined as CSS custom properties in `index.css`.
The base palette (background, foreground, muted, border, etc.) comes from
shadcn tokens. Semantic accents are the main tool for organizing data.

**Accent palette** - 8 hues, each with 3 variants:

| Suffix | Purpose | Example class |
|--------|---------|---------------|
| `accent-{hue}` | Text, icons | `text-accent-blue` |
| `accent-{hue}-tint` | Subtle background wash | `bg-accent-blue-tint` |
| `accent-{hue}-edge` | Borders, left-bar indicators | `border-accent-blue-edge` |

**Hue assignments** - each hue maps to a consistent semantic domain:

| Hue | Domain | Used for |
|-----|--------|----------|
| `blue` | Structure / flow | Composite conditions, camera steps, defending stance, tilesheet badges |
| `violet` | References / meta | Flow steps, template variables, action refs, reflecting stance |
| `amber` | State / values | State steps, key-value conditions, vulnerable stance |
| `teal` | Entities / objects | Entity steps, named conditions, sequence refs |
| `red` | Danger / combat | RPG/combat steps, exposed stance, destructive actions, diff removals |
| `orange` | Transitions / alerts | Transition steps, non-atlas badges, burning status |
| `green` | Success / live data | Live-game indicators, diff additions, confirmations |
| `yellow` | Attention / highlight | Ionized status, selected items, warnings |

**Rules:**

- Never use raw Tailwind color classes (`text-blue-400`, `bg-zinc-900`).
  Use semantic tokens: `text-accent-blue`, `bg-accent-blue-tint`,
  `bg-canvas`, `text-muted-foreground`, etc.
- Asset preview backgrounds (sprites, animations, tile grids) use
  `bg-canvas` - a dark neutral that makes pixel art readable.
- `text-muted-foreground` for secondary/descriptive text.
  `text-foreground` for primary text. No raw gray classes.
- Status colors map to the accent palette: green for success, red for
  error, amber for warning. No separate status tokens.

### Spacing and density

Compact layout. Consistent gap/padding values:

| Context | Value |
|---------|-------|
| Items in a list | `gap-1` (4px) or `gap-1.5` (6px) |
| Fields in a form | `gap-2` (8px) |
| Between sections | `gap-4` (16px) |
| Panel internal padding | `p-2` (8px) compact, `p-3` (12px) content areas |
| Page-level padding | `p-4` (16px) |
| Inline element spacing | `gap-1` (4px) |

### Borders and radius

- Base radius: 6px (`rounded-md`). Small badges: 4px (`rounded`).
- Standard dividers: `border-border`.
- Category-colored borders (step cards, condition cards):
  `border-l-2 border-l-accent-{hue}-edge`.
- Section dividers: `border-b border-border`.
- No box shadows except floating elements (popovers, dropdowns, modals)
  where `shadow-lg` is appropriate.

### Component patterns

**Badges/tags** - tinted background + accent text:
`bg-accent-{hue}-tint text-accent-{hue} text-[10px] font-medium px-1.5 py-0.5 rounded`

**Cards** - `border border-border rounded-md`. Category cards get a left
accent border: `border-l-2 border-l-accent-{hue}-edge`.

**Form inputs** - use shadcn primitives (`Input`, `Select`, `Switch`).
Compact height: `h-7` for standard inputs, `h-6` for inline/embedded.

**Combobox/autocomplete** - several custom implementations exist
(SoundPicker, GlobalKeyInput, EntityRefInput, HandlerRefInput,
ZoneIdInput). They share a pattern: text input with filtered dropdown,
click-outside close, keyboard navigation. These should be consolidated
into a shared `SearchableSelect` component. Until then, follow the
pattern in `GlobalKeyInput.tsx`.

**Section headers** - `text-sm font-semibold text-foreground` with
`border-b border-border pb-1`.

**Empty states** - centered `text-sm text-muted-foreground` with an
actionable message.

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
