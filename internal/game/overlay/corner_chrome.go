package overlay

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

const (
	chromeSize = 32.0
	chromeGap  = 6.0
	chromePad  = 12.0

	popoverGap = 6.0
)

func (o *Overlay) renderCornerChrome(win *opengl.Window) {
	dc := &DrawCtx{
		target: win,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    o.pointer,
		A:      o.alpha,
	}

	wb := win.Bounds()
	x := wb.Max.X - chromePad - chromeSize
	y := chromePad

	// Button rects, right-to-left: gear, fullscreen, scale, volume
	gearR := pixel.R(x, y, x+chromeSize, y+chromeSize)
	x -= chromeSize + chromeGap
	fullscreenR := pixel.R(x, y, x+chromeSize, y+chromeSize)
	x -= chromeSize + chromeGap
	scaleR := pixel.R(x, y, x+chromeSize, y+chromeSize)
	x -= chromeSize + chromeGap
	volumeR := pixel.R(x, y, x+chromeSize, y+chromeSize)

	// Volume popover hit area (used for hover expansion)
	volumePopoverR := volumeSliderRect(volumeR)
	volumeExpanded := contains(volumeR, dc.Ptr.Pos) || contains(volumePopoverR, dc.Ptr.Pos)

	// Scale popover rect (only rendered when open)
	scalePopoverR := scaleListRect(scaleR, o.scaleOptionCount(win))

	// ---- 1. render buttons (bg + icon) ----
	o.drawChromeButton(dc, gearR, drawIconSettings)
	if dc.clicked(gearR) {
		o.panelOpen = !o.panelOpen
	}

	o.drawChromeButton(dc, fullscreenR, o.drawIconFullscreen)
	if dc.clicked(fullscreenR) {
		o.toggleFullscreen()
	}

	o.drawChromeButton(dc, scaleR, o.drawScaleIcon)
	if dc.clicked(scaleR) {
		o.scalePopoverOpen = !o.scalePopoverOpen
	}

	o.drawChromeButton(dc, volumeR, o.drawVolumeIcon)
	if dc.clicked(volumeR) {
		o.toggleMute()
	}

	// ---- 2. popovers on top ----
	if volumeExpanded {
		o.renderVolumePopover(dc, volumePopoverR)
	}
	if o.scalePopoverOpen {
		// close on click outside both the popover and the scale button
		if dc.Ptr.JustUp && !contains(scalePopoverR, dc.Ptr.Pos) && !contains(scaleR, dc.Ptr.Pos) {
			o.scalePopoverOpen = false
		} else {
			o.renderScalePopover(dc, scalePopoverR, win)
		}
	}
}

// drawChromeButton renders one button's bg + its icon/label callback.
func (o *Overlay) drawChromeButton(dc *DrawCtx, r pixel.Rect, drawFg func(dc *DrawCtx, r pixel.Rect)) {
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

// ---- button foreground draws ------------------------------------------------

// chromeIconSize is the rendered size of the 16×16 source sprites (integer 2x).
const chromeIconSize = 32.0

func drawIconSettings(dc *DrawCtx, r pixel.Rect) {
	dc.drawIcon("settings", r.Center(), chromeIconSize)
}

func (o *Overlay) drawIconFullscreen(dc *DrawCtx, r pixel.Rect) {
	name := "fullscreen"
	if o.displayFullscreen {
		name = "fullscreen_exit"
	}
	dc.drawIcon(name, r.Center(), chromeIconSize)
}

func (o *Overlay) drawScaleIcon(dc *DrawCtx, r pixel.Rect) {
	dc.drawIcon("scale", r.Center(), chromeIconSize)
}

func (o *Overlay) drawResetIcon(dc *DrawCtx, r pixel.Rect) {
	dc.drawIcon("reset", r.Center(), chromeIconSize)
}

func (o *Overlay) drawVolumeIcon(dc *DrawCtx, r pixel.Rect) {
	dc.drawIcon(o.volumeSpriteName(), r.Center(), chromeIconSize)
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

const (
	volumePopoverW = 28.0
	volumePopoverH = 110.0
)

func volumeSliderRect(btnR pixel.Rect) pixel.Rect {
	cx := btnR.Center().X
	bottom := btnR.Max.Y + popoverGap
	return pixel.R(cx-volumePopoverW/2, bottom, cx+volumePopoverW/2, bottom+volumePopoverH)
}

func (o *Overlay) renderVolumePopover(dc *DrawCtx, r pixel.Rect) {
	// disable slider drag when muted (but still show it dimmed)
	sliderDC := *dc
	if o.audioMuted {
		sliderDC.A = dc.A * 0.4
	}
	newVol, changed := sliderDC.VerticalSlider(r, o.audioVolume, 0, 1)
	if changed && !o.audioMuted {
		o.setVolume(newVol)
	}
}

// ---- scale click popover ----------------------------------------------------

const (
	scalePopoverW    = 40.0
	scaleOptionH     = 24.0
	scalePopoverPad  = 4.0
)

func (o *Overlay) scaleOptionCount(win *opengl.Window) int {
	wb := win.Bounds()
	maxFit := int(math.Min(math.Floor(wb.W()/240), math.Floor(wb.H()/160)))
	if maxFit < 1 {
		maxFit = 1
	}
	return maxFit + 1 // auto + x1..xMax
}

func scaleListRect(btnR pixel.Rect, nOptions int) pixel.Rect {
	h := float64(nOptions)*scaleOptionH + scalePopoverPad*2
	cx := btnR.Center().X
	bottom := btnR.Max.Y + popoverGap
	return pixel.R(cx-scalePopoverW/2, bottom, cx+scalePopoverW/2, bottom+h)
}

func (o *Overlay) renderScalePopover(dc *DrawCtx, r pixel.Rect, win *opengl.Window) {
	dc.fillRect(r, dc.fade(colBg))

	nOpts := o.scaleOptionCount(win)
	// render from the TOP down: auto is on top, xMax at the bottom (closest to the button)
	// matches the web wrapper which shows higher scales at the bottom of its popover.
	for i := 0; i < nOpts; i++ {
		// index 0 = auto, 1 = x1, ...
		optIdx := i
		// position: top option at r.Max.Y - padding, going down
		top := r.Max.Y - scalePopoverPad - float64(i)*scaleOptionH
		optR := pixel.R(r.Min.X+scalePopoverPad, top-scaleOptionH,
			r.Max.X-scalePopoverPad, top)

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
		default:
			bg = pixel.RGBA{} // transparent
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

