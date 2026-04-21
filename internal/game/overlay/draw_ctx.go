package overlay

import (
	"math"

	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"fisherevans.com/project/f/internal/util/frames"
	"fisherevans.com/project/f/internal/util/gfx"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/ext/imdraw"
	"github.com/gopxl/pixel/v2/ext/text"
	"golang.org/x/image/font/basicfont"
)

// DrawCtx bundles everything widget functions need for one render pass.
// Passed by pointer so widgets can share and flush the single imd instance.
//
// target is a pixel.Target (window or canvas) so the panel can redirect
// draws into an off-screen canvas for clipping / scrolling.
type DrawCtx struct {
	target pixel.Target
	atlas  *resources.Atlas
	imd    *imdraw.IMDraw
	Ptr    Pointer
	A      float64 // fade alpha 0..1
}

// ---- colors (muted greys) ---------------------------------------------------

var (
	colAccent  = colors.HexString("#c8c8c8") // light grey — used for filled/active elements
	colTrack   = colors.HexString("#404040") // darker grey for inactive track/pill
	colText    = colors.HexString("#d0d0d0")
	colDim     = colors.HexString("#7a7a7a")
	colBg      pixel.RGBA // assigned in initColors; needs alpha so can't use HexString
	colHover   pixel.RGBA
	colPress   pixel.RGBA
)

func init() {
	colBg    = colors.LayerAlpha(colors.HexString("#141414"), 0.82)
	colHover = colors.LayerAlpha(colors.HexString("#505050"), 0.22)
	colPress = colors.LayerAlpha(colors.HexString("#505050"), 0.40)
}

// ---- atlas init -------------------------------------------------------------

var (
	overlayAtlas     *resources.Atlas
	overlayRounded2  *frames.Instance // filled 2px rounded frame for pills/buttons
)

func initAtlas() {
	resources.RunOnceInitialized(func() {
		overlayAtlas = resources.DefaultAtlas()
		overlayRounded2 = frames.New("common/rounded_frame_2px", overlayAtlas,
			frames.WithRenderOrigin(gfx.BottomLeft))
	})
}

// drawRoundedRect stretches the 2px rounded frame to r, color-masked.
// Falls back to a solid fillRect if the frame asset hasn't initialised yet.
func (dc *DrawCtx) drawRoundedRect(r pixel.Rect, c pixel.RGBA) {
	if overlayRounded2 == nil {
		dc.fillRect(r, c)
		return
	}
	w := math.Round(r.W())
	h := math.Round(r.H())
	if w <= 0 || h <= 0 {
		return
	}
	overlayRounded2.Draw(dc.target,
		pixel.R(0, 0, w, h),
		pixel.IM.Moved(r.Min),
		frames.WithColor(c))
}

// ---- rect / shape drawing ---------------------------------------------------

// fillRect draws a solid rectangle using the atlas 2×2 sprite (avoids imdraw).
func (dc *DrawCtx) fillRect(r pixel.Rect, c pixel.RGBA) {
	if overlayAtlas == nil {
		return
	}
	w := int(math.Round(r.W()))
	h := int(math.Round(r.H()))
	if w <= 0 || h <= 0 {
		return
	}
	gfx.DrawRect(overlayAtlas, dc.target, pixel.IM.Moved(r.Min), gfx.BottomLeft, w, h, c)
}

// drawIcon renders an atlas sprite centred on `center` scaled uniformly to
// `sizePx` on the longest edge. Colour-masked with `colText` dimmed by the
// current fade alpha so icons pick up the overlay's idle fade.
func (dc *DrawCtx) drawIcon(spriteName string, center pixel.Vec, sizePx float64) {
	if overlayAtlas == nil {
		return
	}
	sprite := overlayAtlas.GetSprite("overlay/icons:" + spriteName)
	if sprite == nil {
		return
	}
	b := sprite.Bounds()
	longest := b.W()
	if b.H() > longest {
		longest = b.H()
	}
	scale := sizePx / longest
	m := pixel.IM.Scaled(pixel.ZV, scale).Moved(center)
	sprite.DrawColorMask(dc.target, m, dc.fade(colText))
}

// circle draws a filled circle using imd, flushing immediately.
func (dc *DrawCtx) circle(center pixel.Vec, radius float64, c pixel.RGBA) {
	dc.imd.Color = c
	dc.imd.Push(center)
	dc.imd.Circle(radius, 0)
	dc.imd.Draw(dc.target)
	dc.imd.Clear()
}

// ---- text -------------------------------------------------------------------

var widgetAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)

func newText() *text.Text { return text.New(pixel.ZV, widgetAtlas) }

// drawText draws txt at pos, flooring both axes to integer pixels. The 7x13
// bitmap font renders blurry if placed at a half-pixel Y (easy to hit with
// `LineHeight/2`, since LineHeight=13 → 6.5).
func (dc *DrawCtx) drawText(txt *text.Text, pos pixel.Vec) {
	txt.Draw(dc.target, pixel.IM.Moved(pixel.V(math.Floor(pos.X), math.Floor(pos.Y))))
}

// ---- alpha helpers ----------------------------------------------------------

func (dc *DrawCtx) fade(c pixel.RGBA) pixel.RGBA {
	return colors.LayerAlpha(c, dc.A)
}

// ---- hit test ---------------------------------------------------------------

func contains(r pixel.Rect, p pixel.Vec) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X && p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

// ---- row background ---------------------------------------------------------

func (dc *DrawCtx) rowBg(r pixel.Rect) {
	hover := contains(r, dc.Ptr.Pos)
	var c pixel.RGBA
	switch {
	case hover && dc.Ptr.Down:
		c = dc.fade(colPress)
	case hover:
		c = dc.fade(colHover)
	default:
		c = dc.fade(colBg)
	}
	dc.fillRect(r, c)
}

// ---- clicked ----------------------------------------------------------------

func (dc *DrawCtx) clicked(r pixel.Rect) bool {
	return contains(r, dc.Ptr.Pos) && dc.Ptr.JustUp
}
