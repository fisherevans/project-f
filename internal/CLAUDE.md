# internal/CLAUDE.md

Rendering, asset, and graphics conventions for all code under `internal/`.

The game runs on `gopxl/pixel v2`. Rather than use pixel primitives directly,
almost all drawing goes through internal wrappers that enforce batching, atlas
usage, integer coordinates, and origin alignment. Use the internal tool if one
exists; fall back to raw pixel only when there's no wrapper.

## What to use for what

| Need | Use | Don't use |
|------|-----|-----------|
| Text | `util/textbox.Instance` + `Content` | raw `pixel/ext/text.Text` |
| Solid rectangle (fill) | `util/gfx.DrawRect` | `imdraw.IMDraw` |
| Straight line | `util/gfx.DrawLine` | `imdraw.IMDraw` |
| 9-slice panel / dialog background | `util/frames.Instance` | hand-rolled sprite math |
| Sprite draw | `atlas.GetSprite(...)` + `DrawColorMask` | `resources.LoadSprite` at render time |
| Tilesheet sprite | `atlas.GetTilesheetSprite(name, col, row)` | manual picture-data slicing |
| Named tile from sheet | `atlas.GetSprite("sheet:spritename")` | ad-hoc lookup |
| Animated sprite | `anim.Load(atlas, "sheet:anim")` | frame counters |
| Color | `util/colors` (`HexString`, `FromString`, named vars) | `pixel.RGB` inline literals |
| Int → vec | `gfx.IVec(x, y)` / `gfx.Moved(x, y)` | `pixel.V(float64(x), float64(y))` |
| Int → rect at origin | `gfx.R(w, h)` | `pixel.R(0, 0, float64(w), float64(h))` |
| Alignment math | `gfx.OriginLocation.Align*` | per-call x/y arithmetic |

`imdraw.IMDraw` is reserved for cases with no sprite representation (e.g. the
text-box's underline pass). If you're tempted to use it for anything else,
there's almost certainly a wrapper.

## The atlas, and why you never draw without one

`resources.Atlas` is a packed texture. Every draw against it can go into a
single `pixel.Batch` that flushes in one GL call. Drawing a sprite loaded
outside the atlas (e.g. via `resources.LoadSprite`) forces a texture switch and
breaks batching.

- Get the shared atlas: `resources.DefaultAtlas()`. Covers every sprite except
  those filtered out by prefix (`title/`, `startup/`).
- Private atlas for screens whose assets aren't in the default (title,
  startup): `resources.CreateAtlas(AtlasFilter{...})`.
- Create a batch: `batch := atlas.NewBatch()`. One per state, reused each
  frame with `batch.Clear()`.
- Draw sprites into the batch, not the canvas, then `batch.Draw(target)` once
  at the end of the render pass. See
  `internal/game/states/combat/combat_state.go` and
  `internal/game/states/adventure/adventure_state.go` for the standard pattern.

Multiple batches per state are fine and expected when layers need to compose
with different shaders or blend modes (e.g. adventure has `sceneBatch`,
`lightMapBatch`, `hudBatch`).

## Initialization order

Assets load from the embedded FS during `resources.Initialize()`. Atlas
creation depends on sprites being parsed first. Anything that references the
atlas at package init time must defer:

```go
var atlas *resources.Atlas

func init() {
    resources.RunOnceInitialized(func() {
        atlas = resources.DefaultAtlas()
        myFrame = frames.New("menu/background", atlas)
        myText = textbox.NewInstance(atlas.GetFont(resources.FontNameM5x7), tbcfg.NewConfig(...))
    })
}
```

`RunOnceInitialized` runs immediately if resources are already loaded,
otherwise queues until `Initialize()` finishes. Never call `atlas.GetSprite`,
`frames.New`, or `textbox.NewInstance` from an `init()` directly - only from
inside the callback.

## Text (`util/textbox`)

Textboxes are instances configured once, then reused to produce `Content`
objects per message.

```go
tb := textbox.NewInstance(atlas.GetFont(resources.FontNameM3x6),
    tbcfg.NewConfig(width, height,
        tbcfg.HAligned(tbcfg.HAlignCenter),
        tbcfg.WithExpandMode(tbcfg.ExpandFit),
        tbcfg.Foreground(colors.Black.RGBA),
    ))

plain := tb.NewSimpleContent("hello world")
styled := tb.NewComplexContent("{+o:black}{+c:warm_5}danger{-*}",
    textbox.WithTyping(0.03))

plain.Render(target, matrix, tbcfg.Foreground(colors.White.RGBA))
```

- Fonts are registered in `resources.FontMetadata` with fixed pixel sizes. Use
  the `FontName*` constants.
- `NewSimpleContent` for plain strings; `NewComplexContent` for the `{+/-}`
  template grammar (color, shadow, outline, underline, rumble, typing weight -
  full grammar in `util/textbox/parsing.go`).
- Render options passed to `Content.Render` override the instance config for
  that call only (common uses: `Foreground`, `ColorMask`, `RenderFrom`).
- `tbcfg.RenderFrom(origin)` controls where the text box's reference point
  sits - you pass the matrix to that origin, the textbox positions itself
  accordingly. Don't pre-offset the matrix.
- Typed content must be updated each frame via `Content.Update(dt, listener)`.
  Static content can skip update.

## Frames (`util/frames`) - 9-slice

Used for panels, dialogs, buttons, anything with a resizable border. Frame
assets are sprites with a YAML sidecar declaring `cutMargin` (where the slices
are cut) and `padding` (inset for children). See `assets/CLAUDE.md`.

```go
f := frames.New("menu/background", atlas, frames.WithRenderOrigin(gfx.TopLeft))
f.Draw(target, pixel.R(0, 0, 100, 60), matrix,
    frames.WithColor(colors.White.RGBA))
```

- `matrix` is the reference point per the render origin.
- `rect` is the target size in pixels (not a world rect).
- `FrameMode` per side controls stretch vs repeat - configured in the
  sidecar, not code.

## Shapes and lines (`util/gfx`)

Both helpers stretch a 1x1 or 2x2 sprite out of the atlas, so they participate
in batching.

```go
gfx.DrawRect(atlas, batch, matrix, gfx.BottomLeft, width, height, color)
gfx.DrawLine(batch, atlas, from, to, thickness, color)
```

- Fills only. For borders, use a frame asset.
- Pass a batch, not the canvas, unless you know you want the immediate draw.
- `DrawRect`'s origin parameter means "where does `matrix` sit relative to the
  rectangle" - same convention as frames and textbox.

## Coordinate and origin conventions

- Game logic works in integer pixels on a 240x160 canvas. Always use
  `gfx.IVec` / `gfx.Moved` for integer positions so floats don't sneak in and
  cause subpixel jitter.
- `gfx.OriginLocation` (Centered, TopLeft, ... BottomCenter) is the shared
  origin enum used by `frames`, `textbox.tbcfg`, and `gfx.DrawRect`. Use
  `origin.Align*` to compute offsets when placing one element inside another
  rather than re-deriving the math.

## Animations (`game/anim`)

`anim.AnimatedSprite` wraps frame playback for tilesheet-defined animations.

```go
a := anim.Load(atlas, "combat/combatant_stats/tempo:level_3_border")
// each frame:
a.Update(dt)
a.Sprite().DrawColorMask(batch, matrix, mask)
```

- `name` is `"tilesheet"` (uses animation `default`) or `"tilesheet:anim"`.
- `NewStaticAnimation(drawable)` wraps a single sprite when you need the
  `AnimatedSprite` interface.
- Animation playback metadata (FPS, jitter, ping-pong, reverse, randomize)
  lives in the sprite's YAML sidecar, not code. See root CLAUDE.md for the
  sidecar grammar.

## Colors (`util/colors`)

- Named palette vars: `colors.White`, `colors.Warm5`, `colors.SkillTypeKinetic`
  etc. Each is a `NamedColor` with `.RGBA`. Prefer these over ad-hoc hex when
  one fits.
- `colors.HexString("#rrggbb")` / `colors.FromString("name_or_hex")` for
  arbitrary values.
- `WithAlpha(c, a)`: rescale to a given alpha (treats source alpha as
  pre-multiplier). `LayerAlpha(c, a)`: straight multiply.
- `Lerp` (linear) / `GammaLerp` (sRGB-aware) for transitions.

## Shaders (`game/shaders`)

`shaders.Canvas` wraps `opengl.Canvas` with cached fragment shader swaps. The
runtime owns the final pixel-grid canvas; states render into a regular
`opengl.Canvas` or `shaders.Canvas` and let the runtime composite.

- Use `pixel.ComposeMethod` (Over, Multiply, Screen) for blend effects rather
  than shader tricks where possible - see adventure's lighting pipeline.
- Post-process shaders (bloom, swirl, glitch, flush) are parameterized
  through the `shaders.Options` interface and applied through
  `game.AppliedShader`.

## Debug helpers

From `game/context.go`:

- `game.DebugTLf(...)` / `DebugTRf` / `DebugBLf` / `DebugBRf`: per-frame debug
  text in a screen corner. Cheap, leave them in during development.
- `game.DebugNotificationf(...)`: transient toast.
- `game.DebugToggles().F1() ...F12()`: keys for toggling behaviors in dev
  builds.
