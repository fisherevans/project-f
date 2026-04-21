# assets/CLAUDE.md

How the embedded FS under `assets/` maps to runtime resources. Loaded by
`internal/resources` at startup - see `resources/resources.go` for the root
registry and `resources/resource_*.go` for per-type loaders.

Roots, loaders, and naming:

| Root | Extension | Resource key | Loader |
|------|-----------|--------------|--------|
| `sprites/` | `.png` (+ optional `.yaml`) | path without `.png` | sprite loader; YAML configures frame/tilesheet/animations/named sprites |
| `fonts/` | `.ttf` | basename | font loader; metadata in Go (`FontName*` constants + `fontMetadata`) |
| `maps/` | `.json` | path without `.json` | custom map format |
| `tiled_maps/` | `.tmx` | path without `.tmx` | Tiled |
| `audio/sounds/` | `.wav .mp3 .ogg` | path | audio loader |
| `songs/` | (handled separately) | - | see `game/audio` |

Filenames must be lowercase - the loader skips any file with an upper-case
character.

## Sprite YAML sidecars

Every PNG under `sprites/` may have a YAML at the same path. The YAML decides
how the sprite is treated in the atlas. Pick exactly one of:

### Tilesheet

For grid-indexed sprite sheets. Each cell becomes a `TilesheetSpriteId`
addressable by `atlas.GetTilesheetSprite(name, col, row)` (1-indexed; row 1
is the top row).

```yaml
tilesheet:
  tileWidth: 16
  tileHeight: 16
```

Add `animations` to the same sidecar to declare named animations - see root
CLAUDE.md for the animation grammar.

Add `sprites` to alias specific cells with names; they become accessible via
`atlas.GetSprite("sheet_name:alias")`:

```yaml
sprites:
  door_closed:
    row: 3
    column: 7
```

### Frame (9-slice)

For panels, dialogs, buttons. `cutMargin` is where the sheet is sliced; the
remaining center stretches or repeats. `padding` is the inset applied by
callers that honor it (textbox inside a frame uses it).

```yaml
frame:
  defaults:
    cutMargin: 2
    padding: 2
    frameMode: stretch    # or: repeat
```

Override per side if the slice isn't symmetric:

```yaml
frame:
  padding:
    top: 2
    bottom: 3
    left: 3
    right: 3
  defaults:
    cutMargin: 3
    frameMode: stretch
```

`frameMode` is per-side: `stretch` scales the slice to fit, `repeat` tiles
it. Corners never scale.

### Plain sprite

No YAML, or an empty one. The whole PNG becomes one sprite, addressable by
`atlas.GetSprite("path/without/extension")`.

### Non-atlas sprite

```yaml
nonAtlasSprite: true
```

Skips atlas packing - the sprite lives in its own texture. Use only when the
sprite genuinely can't share the atlas (oversized, e.g. full-screen
backgrounds sampled by shaders).

## Scaffolding new sprite sheets

`cmd/sprite_new` generates a blank tilesheet - PNG at the right grid
dimensions, YAML sidecar using the grammar above, and a matching `.aseprite`
if the Aseprite CLI is available. Invoke from the project root:

```
go run ./cmd/sprite_new -name assets/sprites/<path>/<basename> \
    -tile-width N -tile-height N -cols N -rows N \
    -sprites alias1,alias2,-,alias4
```

Sprite aliases are positional (row-major); `-` or blank leaves a cell
unnamed. The tool refuses to overwrite existing files unless you pass
`-force`. Open the `.aseprite` file to start drawing; the PNG is what the
game loads and what the atlas consumes.

## What to edit vs regenerate

- `.aseprite` files are the source of truth for art; `.png` siblings are
  exports. Don't edit the PNG directly if an aseprite exists.
- The YAML sidecar is hand-written. The sprite loader never rewrites it.
- Adding a new sprite: drop the PNG, add a sidecar if it needs structure,
  and reference it with `atlas.GetSprite(path_without_extension)`. Restart
  the dev binary to re-run `resources.Initialize`.
