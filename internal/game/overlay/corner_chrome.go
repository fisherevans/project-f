package overlay

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

// Base chrome dimensions at uiScale = 1.0.
const (
	chromeSize = 32.0
	chromeGap  = 6.0
	chromePad  = 12.0
	popoverGap = 6.0

	chromeIconSize = 32.0 // source sprite size; scaled to button size at render time
)

// chromeLayout holds scaled chrome dimensions computed once per frame.
type chromeLayout struct {
	sz, pad, gap, pgap float64
}

func newChromeLayout(wb pixel.Rect) chromeLayout {
	us := OverlayUIScale(wb)
	return chromeLayout{
		sz:   math.Floor(chromeSize * us),
		pad:  math.Floor(chromePad * us),
		gap:  math.Floor(chromeGap * us),
		pgap: math.Floor(popoverGap * us),
	}
}

const volumePopoverHideDelay = 0.25 // seconds of cursor-outside before popover closes

func (o *Overlay) renderCornerChrome(win *opengl.Window, dt float64) {
	wb := win.Bounds()
	cl := newChromeLayout(wb)

	dc := &DrawCtx{
		target: win,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    o.pointer,
		A:      o.alpha,
		Scale:  OverlayUIScale(wb),
	}

	// Gear button at top-right
	gearX := math.Floor(wb.Max.X - cl.pad - cl.sz)
	gearY := math.Floor(wb.Max.Y - cl.pad - cl.sz)
	gearR := pixel.R(gearX, gearY, gearX+cl.sz, gearY+cl.sz)

	// Remaining buttons at bottom-right: fullscreen, scale, volume (right to left)
	x := math.Floor(wb.Max.X - cl.pad - cl.sz)
	y := math.Floor(cl.pad)
	var fullscreenR pixel.Rect
	if SupportsFullscreen() {
		fullscreenR = pixel.R(x, y, x+cl.sz, y+cl.sz)
		x -= cl.sz + cl.gap
	}
	scaleR := pixel.R(x, y, x+cl.sz, y+cl.sz)
	x -= cl.sz + cl.gap
	volumeR := pixel.R(x, y, x+cl.sz, y+cl.sz)

	volumePopoverR := volumeSliderRect(volumeR, cl)

	// Volume popover: click/tap toggles open; while open, hovering button or
	// popover resets the close timer; moving away closes after the grace period.
	// On mobile there is no hover so the tap-to-toggle is the only mechanism.
	if dc.clicked(volumeR) {
		o.volumePopoverOpen = !o.volumePopoverOpen
		o.volumePopoverCloseTimer = 0
	} else {
		inButton := contains(volumeR, dc.Ptr.Pos)
		inPopover := o.volumePopoverOpen && contains(volumePopoverR, dc.Ptr.Pos)
		if inButton || inPopover || dc.Ptr.Down {
			o.volumePopoverCloseTimer = 0
		} else if o.volumePopoverOpen {
			o.volumePopoverCloseTimer += dt
			if o.volumePopoverCloseTimer >= volumePopoverHideDelay {
				o.volumePopoverOpen = false
				o.volumePopoverCloseTimer = 0
			}
		}
	}

	scalePopoverR := scaleListRect(scaleR, o.scaleOptionCount(win), cl)

	// 1. Buttons
	drawIconFn := func(name string) func(*DrawCtx, pixel.Rect) {
		return func(dc *DrawCtx, r pixel.Rect) { dc.drawIcon(name, r.Center(), r.W()) }
	}
	o.drawChromeButton(dc, gearR, drawIconFn("settings"))
	if dc.clicked(gearR) && !o.panelJustClosed {
		o.panelOpen = !o.panelOpen
	}
	o.panelJustClosed = false

	if SupportsFullscreen() {
		o.drawChromeButton(dc, fullscreenR, func(dc *DrawCtx, r pixel.Rect) {
			name := "fullscreen"
			if o.displayFullscreen {
				name = "fullscreen_exit"
			}
			dc.drawIcon(name, r.Center(), r.W())
		})
		if dc.clicked(fullscreenR) {
			o.toggleFullscreen()
		}
	}

	o.drawChromeButton(dc, scaleR, func(dc *DrawCtx, r pixel.Rect) {
		dc.drawIcon("scale", r.Center(), r.W())
	})
	if dc.clicked(scaleR) {
		o.scalePopoverOpen = !o.scalePopoverOpen
	}

	o.drawChromeButton(dc, volumeR, func(dc *DrawCtx, r pixel.Rect) {
		dc.drawIcon(o.volumeSpriteName(), r.Center(), r.W())
	})

	// 2. Popovers
	if o.volumePopoverOpen {
		o.renderVolumePopover(dc, volumePopoverR)
	}
	if o.scalePopoverOpen {
		if dc.Ptr.JustUp && !contains(scalePopoverR, dc.Ptr.Pos) && !contains(scaleR, dc.Ptr.Pos) {
			o.scalePopoverOpen = false
		} else {
			o.renderScalePopover(dc, scalePopoverR, win, cl)
		}
	}
}

func (o *Overlay) drawChromeButton(dc *DrawCtx, r pixel.Rect, drawFg func(*DrawCtx, pixel.Rect)) {
	hover := contains(r, dc.Ptr.Pos)
	var bg pixel.RGBA
	switch {
	case hover && dc.Ptr.Down:
		bg = dc.fade(colPress)
	case hover:
		bg = dc.fade(colHover)
	default:
		bg = dc.fade(colBg)
	}
	dc.fillRect(r, bg)
	drawFg(dc, r)
}

func (o *Overlay) volumeSpriteName() string {
	switch {
	case o.audioMuted || o.audioVolume <= 0:
		return "volume_muted"
	case o.audioVolume <= 1.0/3.0:
		return "volume_low"
	case o.audioVolume <= 2.0/3.0:
		return "volume_med"
	default:
		return "volume_high"
	}
}

// ---- volume hover popover ---------------------------------------------------

func volumeSliderRect(btnR pixel.Rect, cl chromeLayout) pixel.Rect {
	w := math.Floor(cl.sz * 0.875) // ~28px at 1x
	h := math.Floor(cl.sz * 3.4)   // ~110px at 1x
	cx := btnR.Center().X
	bot := btnR.Max.Y + cl.pgap
	return pixel.R(math.Floor(cx-w/2), bot, math.Floor(cx+w/2), bot+h)
}

func (o *Overlay) renderVolumePopover(dc *DrawCtx, r pixel.Rect) {
	sliderDC := *dc
	if o.audioMuted {
		sliderDC.A = dc.A * 0.4
	}
	newVol, changed := sliderDC.VerticalSlider(r, o.audioVolume, 0, 1)
	if changed {
		o.setVolume(newVol)
	}
}

// ---- scale click popover ----------------------------------------------------

func (o *Overlay) scaleOptionCount(win *opengl.Window) int {
	wb := win.Bounds()
	physFit := int(math.Min(math.Floor(wb.W()/240), math.Floor(wb.H()/160)))
	if physFit < 1 {
		physFit = 1
	}
	return physFit + 1 // auto + x1..xPhysFit
}

func scaleListRect(btnR pixel.Rect, nOptions int, cl chromeLayout) pixel.Rect {
	optH := math.Floor(cl.sz * 0.75) // ~24px at 1x
	pad := math.Floor(cl.sz * 0.125) // ~4px at 1x
	w := math.Floor(cl.sz * 1.25)    // ~40px at 1x
	h := float64(nOptions)*optH + pad*2
	cx := btnR.Center().X
	bot := btnR.Max.Y + cl.pgap
	return pixel.R(math.Floor(cx-w/2), bot, math.Floor(cx+w/2), bot+h)
}

func (o *Overlay) renderScalePopover(dc *DrawCtx, r pixel.Rect, win *opengl.Window, cl chromeLayout) {
	dc.fillRect(r, dc.fade(colBg))

	optH := math.Floor(cl.sz * 0.75)
	pad := math.Floor(cl.sz * 0.125)
	nOpts := o.scaleOptionCount(win)

	for i := 0; i < nOpts; i++ {
		optIdx := i
		top := r.Max.Y - pad - float64(i)*optH
		optR := pixel.R(r.Min.X+pad, top-optH, r.Max.X-pad, top)

		hover := contains(optR, dc.Ptr.Pos)
		isSelected := optIdx == o.displayScale

		var bg pixel.RGBA
		switch {
		case isSelected:
			bg = dc.fade(colAccent)
		case hover && dc.Ptr.Down:
			bg = dc.fade(colPress)
		case hover:
			bg = dc.fade(colHover)
		}
		if bg.A > 0 {
			dc.fillRect(optR, bg)
		}

		label := "auto"
		if optIdx > 0 {
			label = fmt.Sprintf("x%d", optIdx)
		}
		txt := newText()
		if isSelected {
			txt.Color = colors.LayerAlpha(colors.HexString("#101010"), dc.A)
		} else {
			txt.Color = dc.fade(colText)
		}
		txt.WriteString(label)
		tw := txt.Bounds().W()
		cx := optR.Min.X + (optR.W()-tw)/2
		cy := optR.Min.Y + (optR.H()-txt.LineHeight)/2
		dc.drawText(txt, pixel.V(cx, cy))

		if hover && dc.Ptr.JustUp {
			o.setScale(optIdx)
			o.scalePopoverOpen = false
		}
	}
}
