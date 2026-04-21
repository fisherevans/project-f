package overlay

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
)

// base panel dimensions at uiScale = 1.0
const (
	basePanelW        = 360.0
	basePanelPad      = 14.0
	baseRowH          = 30.0
	baseRowGap        = 3.0
	baseHeaderH       = 22.0
	baseScenesRowH    = 22.0
	baseScrollbarW    = 8.0
	baseScrollbarGap  = 4.0
	baseScrollWheelPx = 30.0
	baseMinThumbH     = 20.0
	baseDropdownItemH = 24.0
	baseDropdownPad   = 4.0
	baseDropdownGap   = 4.0
)

// pl holds all panel geometry pre-scaled for the current frame.
type pl struct {
	W, Pad, RowH, RowGap, HeaderH, ScenesRowH                     float64
	ScrollbarW, ScrollbarGap, ScrollWheelPx, MinThumbH             float64
	DropdownItemH, DropdownPad, DropdownGap                        float64
}

func newPL(uiScale float64) pl {
	ps := func(v float64) float64 { return math.Floor(v * uiScale) }
	return pl{
		W: ps(basePanelW), Pad: ps(basePanelPad),
		RowH: ps(baseRowH), RowGap: ps(baseRowGap), HeaderH: ps(baseHeaderH),
		ScenesRowH: ps(baseScenesRowH),
		ScrollbarW: ps(baseScrollbarW), ScrollbarGap: ps(baseScrollbarGap),
		ScrollWheelPx: ps(baseScrollWheelPx), MinThumbH: ps(baseMinThumbH),
		DropdownItemH: ps(baseDropdownItemH), DropdownPad: ps(baseDropdownPad),
		DropdownGap: ps(baseDropdownGap),
	}
}

func (o *Overlay) renderSettingsPanel(win *opengl.Window) {
	// ---- 1. compute panel rect in window space (integer-aligned) -------------
	wb := win.Bounds()
	uiScale := OverlayUIScale(wb)
	o.panelScale = uiScale
	layout := newPL(uiScale)
	cl := newChromeLayout(wb)
	panelX := math.Floor(wb.Max.X - layout.W - cl.pad)
	panelMaxY := math.Floor(wb.Max.Y - cl.pad)
	panel := pixel.R(panelX, ChromeAreaBottom(wb), panelX+layout.W, panelMaxY)

	// outer DC: panel bg, border, scrollbar, close-on-outside
	outerDC := &DrawCtx{
		target: win,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    o.pointer,
		A:      o.alpha,
		Scale:  uiScale,
	}

	outerDC.fillRect(panel, colors.LayerAlpha(colors.HexString("#0d0d0d"), 0.94*o.alpha))
	outerDC.fillRect(pixel.R(panel.Min.X, panel.Min.Y, panel.Min.X+1, panel.Max.Y),
		colors.LayerAlpha(colors.HexString("#333333"), o.alpha))

	// close on click outside the panel; include any open dropdown popover so clicks
	// inside it don't accidentally close the panel.
	insidePanel := contains(panel, o.pointer.Pos)
	if o.dropdownID != "" {
		insidePanel = insidePanel || contains(o.dropdownPopoverRect(), o.pointer.Pos)
	}
	if o.pointer.JustUp && !insidePanel && !o.scrollDragging {
		o.panelOpen = false
		o.panelJustClosed = true
		o.dropdownID = ""
		return
	}

	// ---- 2. split panel into content area + scrollbar rail -------------------
	// Canvas dims must be EVEN or the canvas center lands on a half-pixel when
	// blitted to the window, producing sub-pixel sampling and ghost trails
	// (especially visible during scroll).
	rawContentW := (panel.Max.X - layout.ScrollbarW - layout.ScrollbarGap) - (panel.Min.X + 1)
	rawContentH := panel.H()
	contentW := math.Floor(rawContentW/2) * 2
	visibleH := math.Floor(rawContentH/2) * 2
	contentRect := pixel.R(
		panel.Min.X+1,
		panel.Min.Y,
		panel.Min.X+1+contentW,
		panel.Min.Y+visibleH,
	)
	scrollbarTrack := pixel.R(
		panel.Max.X-layout.ScrollbarW-layout.ScrollbarGap/2,
		panel.Min.Y+layout.Pad,
		panel.Max.X-layout.ScrollbarGap/2,
		panel.Max.Y-layout.Pad,
	)

	// ---- 3. update scroll offset from input, clamped against last-frame's
	// measured content height (applied again after this frame measures) ------
	o.scrollMax = math.Max(0, o.scrollContentH-visibleH)

	if contains(panel, o.pointer.Pos) {
		if dy := win.MouseScroll().Y; dy != 0 {
			// Browsers report raw pixel deltas (~100/notch); desktop GLFW reports
			// ~1/notch. Normalize anything above 1 down to 1 unit so sensitivity
			// is consistent, while preserving sub-1 values for trackpad smoothness.
			if dy > 1 {
				dy = 1
			} else if dy < -1 {
				dy = -1
			}
			o.scrollOffset -= dy * layout.ScrollWheelPx
			o.dropdownID = "" // dropdown anchor moves with scroll; just close it
		}
	}

	if o.scrollDragging {
		if !o.pointer.Down {
			o.scrollDragging = false
		} else if o.scrollContentH > visibleH {
			thumbH := math.Max(layout.MinThumbH, scrollbarTrack.H()*visibleH/o.scrollContentH)
			travelH := scrollbarTrack.H() - thumbH
			if travelH > 0 {
				dy := o.scrollDragStartY - o.pointer.Pos.Y
				frac := dy / travelH
				o.scrollOffset = o.scrollDragStartOff + frac*o.scrollMax
			}
		}
	}

	if o.scrollOffset > o.scrollMax {
		o.scrollOffset = o.scrollMax
	}
	if o.scrollOffset < 0 {
		o.scrollOffset = 0
	}

	// ---- 5. ensure scroll canvas is sized to content area -------------------
	bounds := pixel.R(0, 0, contentW, visibleH)
	if o.scrollCanvas == nil {
		o.scrollCanvas = opengl.NewCanvas(bounds)
		o.scrollCanvas.SetSmooth(false)
	} else if o.scrollCanvas.Bounds() != bounds {
		o.scrollCanvas.SetBounds(bounds)
	}
	o.scrollCanvas.Clear(pixel.RGBA{})

	// ---- 6. inner DC renders into the canvas in canvas-local coords ---------
	innerPtr := o.pointer
	innerPtr.Pos = o.pointer.Pos.Sub(contentRect.Min)
	// Suppress clicks outside the visible content area, during scrollbar drag, or
	// while a dropdown popover is open (the popover handles its own clicks in
	// window space via outerDC).
	if o.scrollDragging || !contains(contentRect, o.pointer.Pos) || o.dropdownID != "" {
		innerPtr.Down = false
		innerPtr.JustDown = false
		innerPtr.JustUp = false
	}
	innerDC := &DrawCtx{
		target: o.scrollCanvas,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    innerPtr,
		A:      o.alpha,
		Scale:  uiScale,
	}

	// ---- 7. render sections into the canvas ---------------------------------
	// cursor is canvas-local Y, starting just inside the top edge, shifted up
	// by scrollOffset so larger offsets pull content into view from below.
	startCursor := visibleH - layout.Pad + o.scrollOffset
	cursor := startCursor
	cursor = o.renderAudioSection(innerDC, bounds, cursor, layout)
	cursor = o.renderDisplaySection(innerDC, bounds, cursor, win, contentRect, layout)
	cursor = o.renderScenesSection(innerDC, bounds, cursor, layout)
	cursor = o.renderDebugSection(innerDC, bounds, cursor, contentRect, layout)
	cursor = o.renderSystemSection(innerDC, bounds, cursor, layout)

	// Measure intrinsic content height for next-frame clamping. startCursor and
	// cursor both include +scrollOffset so the difference is scroll-independent;
	// add 2*panelPad for top + bottom breathing room around the content.
	totalConsumed := startCursor - cursor
	o.scrollContentH = totalConsumed + 2*layout.Pad
	o.scrollMax = math.Max(0, o.scrollContentH-visibleH)

	// ---- 8. blit canvas into window at content rect -------------------------
	o.scrollCanvas.Draw(win, pixel.IM.Moved(contentRect.Center()))

	// ---- 8b. render dropdown popover in window space (above the canvas clip) -
	if o.dropdownID != "" {
		o.renderDropdownPopover(outerDC)
	}

	// ---- 9. draw scrollbar (only if content overflows) ----------------------
	if o.scrollMax > 0 {
		outerDC.fillRect(scrollbarTrack,
			colors.LayerAlpha(colors.HexString("#202020"), 0.6*o.alpha))

		thumbH := math.Max(layout.MinThumbH, scrollbarTrack.H()*visibleH/o.scrollContentH)
		travelH := scrollbarTrack.H() - thumbH
		frac := 0.0
		if o.scrollMax > 0 {
			frac = o.scrollOffset / o.scrollMax
		}
		// scrollOffset=0 → thumb at top of track (Y-up: max Y end).
		thumbTop := scrollbarTrack.Max.Y - frac*travelH
		thumbRect := pixel.R(scrollbarTrack.Min.X, thumbTop-thumbH,
			scrollbarTrack.Max.X, thumbTop)

		thumbCol := colors.HexString("#606060")
		if o.scrollDragging || contains(thumbRect, o.pointer.Pos) {
			thumbCol = colors.HexString("#8a8a8a")
		}
		outerDC.fillRect(thumbRect, colors.LayerAlpha(thumbCol, o.alpha))

		// Start drag on mousedown on the thumb.
		if o.pointer.JustDown && contains(thumbRect, o.pointer.Pos) {
			o.scrollDragging = true
			o.scrollDragStartY = o.pointer.Pos.Y
			o.scrollDragStartOff = o.scrollOffset
		}
	}
}

// ---- AUDIO ------------------------------------------------------------------

func (o *Overlay) renderAudioSection(dc *DrawCtx, panel pixel.Rect, cursorY float64, l pl) float64 {
	x0, w := panel.Min.X+l.Pad, panel.W()-l.Pad*2

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-l.HeaderH, x0+w, cursorY), "AUDIO", func() {
		*game.CurrentSave().SystemSettings.Audio = rpg.SystemSettingsAudio{}
		game.CurrentSave().SystemSettings.Audio.FillDefaults()
		s := game.CurrentSave().SystemSettings.Audio
		o.audioMuted = s.Muted
		o.audioVolume = s.MasterVolume
		if o.hooks.SetAudio != nil {
			o.hooks.SetAudio(o.audioMuted, o.audioVolume)
		}
		o.markDirty()
	})
	cursorY -= l.HeaderH + l.RowGap

	newMuted, changed := dc.Toggle(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY), "Mute", o.audioMuted)
	if changed {
		o.audioMuted = newMuted
		if o.hooks.SetAudio != nil {
			o.hooks.SetAudio(o.audioMuted, o.audioVolume)
		}
		o.markDirty()
	}
	cursorY -= l.RowH + l.RowGap

	sliderDC := *dc
	if o.audioMuted {
		sliderDC.A = dc.A * 0.35
	}
	newVol, volChanged := sliderDC.Slider(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY), "Volume", o.audioVolume, 0, 1)
	if volChanged && !o.audioMuted {
		o.setVolume(newVol)
	}
	cursorY -= l.RowH + l.RowGap*3

	return cursorY
}

// ---- DISPLAY ----------------------------------------------------------------

func (o *Overlay) renderDisplaySection(dc *DrawCtx, panel pixel.Rect, cursorY float64, win *opengl.Window, contentRect pixel.Rect, l pl) float64 {
	x0, w := panel.Min.X+l.Pad, panel.W()-l.Pad*2

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-l.HeaderH, x0+w, cursorY), "DISPLAY", func() {
		// Reset scale to auto. Leave fullscreen as-is (toggling here would hit
		// the OS in a confusing way — users can toggle it explicitly).
		o.setScale(0)
	})
	cursorY -= l.HeaderH + l.RowGap

	newFS, fsChanged := dc.Toggle(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY), "Fullscreen", o.displayFullscreen)
	if fsChanged {
		o.displayFullscreen = newFS
		if o.hooks.ToggleFullscreen != nil {
			o.hooks.ToggleFullscreen()
		}
		o.markDirty()
	}
	cursorY -= l.RowH + l.RowGap

	wb := win.Bounds()
	maxFit := int(math.Min(math.Floor(wb.W()/240), math.Floor(wb.H()/160)))
	if maxFit < 1 {
		maxFit = 1
	}

	opts := make([]string, 0, maxFit+1)
	opts = append(opts, "auto")
	for n := 1; n <= maxFit; n++ {
		opts = append(opts, fmt.Sprintf("x%d", n))
	}

	selectedIdx := 0
	if o.displayScale > 0 && o.displayScale <= maxFit {
		selectedIdx = o.displayScale
	}

	scaleRow := pixel.R(x0, cursorY-l.RowH, x0+w, cursorY)
	if dc.DropdownSelectRow(scaleRow, "Scale", opts[selectedIdx], o.dropdownID == "scale") {
		o.toggleDropdown("scale", o.canvasToWindow(scaleRow, contentRect), opts, selectedIdx, func(i int) {
			o.setScale(i)
		})
	}
	cursorY -= l.RowH + l.RowGap

	// Virtual Gamepad
	vgOpts := []string{"Auto", "On", "Off"}
	vgVals := []string{rpg.VirtualGamepadAuto, rpg.VirtualGamepadOn, rpg.VirtualGamepadOff}
	d := game.CurrentSave().SystemSettings.Display
	vgSel := indexOf(vgVals, d.VirtualGamepad)
	vgRow := pixel.R(x0, cursorY-l.RowH, x0+w, cursorY)
	if dc.DropdownSelectRow(vgRow, "Virtual Gamepad", vgOpts[vgSel], o.dropdownID == "vgpad") {
		o.toggleDropdown("vgpad", o.canvasToWindow(vgRow, contentRect), vgOpts, vgSel, func(i int) {
			d.VirtualGamepad = vgVals[i]
			o.markDirty()
		})
	}
	cursorY -= l.RowH + l.RowGap

	// Opacity slider — only when the gamepad can be visible
	if d.VirtualGamepad != rpg.VirtualGamepadOff {
		newOp, opChanged := dc.Slider(
			pixel.R(x0, cursorY-l.RowH, x0+w, cursorY),
			"Gamepad Opacity", d.VirtualGamepadOpacity, 0.2, 1.0,
		)
		if opChanged {
			d.VirtualGamepadOpacity = newOp
			o.markDirty()
		}
		cursorY -= l.RowH + l.RowGap
	}

	cursorY -= l.RowGap * 2

	return cursorY
}

// ---- SCENES -----------------------------------------------------------------

func (o *Overlay) renderScenesSection(dc *DrawCtx, panel pixel.Rect, cursorY float64, l pl) float64 {
	if len(o.hooks.DevScenes) == 0 {
		return cursorY
	}
	x0, w := panel.Min.X+l.Pad, panel.W()-l.Pad*2

	dc.SectionHeader(pixel.R(x0, cursorY-l.HeaderH, x0+w, cursorY), "SCENES")
	cursorY -= l.HeaderH + l.RowGap

	for _, scene := range o.hooks.DevScenes {
		r := pixel.R(x0, cursorY-l.ScenesRowH, x0+w, cursorY)
		if dc.Button(r, scene.Name) {
			game.SetActiveStateIntent(scene.Factory())
			o.panelOpen = false
		}
		cursorY -= l.ScenesRowH + l.RowGap
	}
	cursorY -= l.RowGap * 2
	return cursorY
}

// ---- DEBUG ------------------------------------------------------------------

func (o *Overlay) renderDebugSection(dc *DrawCtx, panel pixel.Rect, cursorY float64, contentRect pixel.Rect, l pl) float64 {
	x0, w := panel.Min.X+l.Pad, panel.W()-l.Pad*2
	s := game.CurrentSave().SystemSettings

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-l.HeaderH, x0+w, cursorY), "DEBUG", func() {
		ss := game.CurrentSave().SystemSettings
		*ss.Debugging = rpg.SystemSettingsDebugging{}
		ss.Debugging.FillDefaults()
		*ss.Lighting = rpg.SystemSettingsLighting{}
		ss.Lighting.FillDefaults()
		*ss.RetroFrame = rpg.SystemSettingsRetroFrame{}
		ss.RetroFrame.FillDefaults()
		game.Flags().Set("retro_frame_reset")
		o.markDirty()
	})
	cursorY -= l.HeaderH + l.RowGap

	// Debug overlay (FPS, memory, etc.) - also toggled with backslash key
	debugVal, debugChanged := dc.Toggle(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY),
		"Show Overlay", game.ShowDebug())
	if debugChanged {
		game.SetShowDebug(debugVal)
	}
	cursorY -= l.RowH + l.RowGap

	// Pathfinding debug
	newVal, changed := dc.Toggle(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY),
		"Pathfinding", s.Debugging.ShowPathfindingDebugging)
	if changed {
		s.Debugging.ShowPathfindingDebugging = newVal
		o.markDirty()
	}
	cursorY -= l.RowH + l.RowGap

	// Lighting Mode
	lightingOpts := []string{"Blended", "Over", "Off"}
	lightingVals := []rpg.LightingMode{rpg.LightingModeBlended, rpg.LightingModeOver, rpg.LightingModeOff}
	lSel := indexOf(lightingVals, s.Lighting.LightingMode)
	lRow := pixel.R(x0, cursorY-l.RowH, x0+w, cursorY)
	if dc.DropdownSelectRow(lRow, "Lighting", lightingOpts[lSel], o.dropdownID == "lighting") {
		o.toggleDropdown("lighting", o.canvasToWindow(lRow, contentRect), lightingOpts, lSel, func(i int) {
			s.Lighting.LightingMode = lightingVals[i]
			o.markDirty()
		})
	}
	cursorY -= l.RowH + l.RowGap

	// Lighting Composition
	compOpts := []string{"Full", "Ambient"}
	compVals := []rpg.LightingComposition{rpg.LightingCompositionFull, rpg.LightingCompositionAmbient}
	cSel := indexOf(compVals, s.Lighting.LightingComposition)
	cRow := pixel.R(x0, cursorY-l.RowH, x0+w, cursorY)
	if dc.DropdownSelectRow(cRow, "Light Comp", compOpts[cSel], o.dropdownID == "lightcomp") {
		o.toggleDropdown("lightcomp", o.canvasToWindow(cRow, contentRect), compOpts, cSel, func(i int) {
			s.Lighting.LightingComposition = compVals[i]
			o.markDirty()
		})
	}
	cursorY -= l.RowH + l.RowGap

	// Bloom Mode
	bloomOpts := []string{"Blended", "Blur", "Over Threshold", "Off"}
	bloomVals := []rpg.BloomMode{rpg.BloomModeBlended, rpg.BloomModeOverBlurred, rpg.BloomModeOverThreshold, rpg.BloomModeOff}
	bSel := indexOf(bloomVals, s.Lighting.BloomMode)
	bRow := pixel.R(x0, cursorY-l.RowH, x0+w, cursorY)
	if dc.DropdownSelectRow(bRow, "Bloom", bloomOpts[bSel], o.dropdownID == "bloom") {
		o.toggleDropdown("bloom", o.canvasToWindow(bRow, contentRect), bloomOpts, bSel, func(i int) {
			s.Lighting.BloomMode = bloomVals[i]
			o.markDirty()
		})
	}
	cursorY -= l.RowH + l.RowGap

	// Pixel Grid (note: field is DisablePixelGrid, so UI shows inverse)
	pixelGridOn := !s.RetroFrame.DisablePixelGrid
	newPG, pgChanged := dc.Toggle(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY), "Pixel Grid", pixelGridOn)
	if pgChanged {
		s.RetroFrame.DisablePixelGrid = !newPG
		game.Flags().Set("retro_frame_reset")
		o.markDirty()
	}
	cursorY -= l.RowH + l.RowGap

	// Retro-frame float sliders (range 0 - 0.25, three-decimal display)
	retroOpts := SliderOpts{Format: "%.3f"}
	sliders := []struct {
		label string
		ptr   *float64
	}{
		{"Scanline", &s.RetroFrame.ScanlineDarken},
		{"Grid X", &s.RetroFrame.GridDarkenX},
		{"Grid Y", &s.RetroFrame.GridDarkenY},
		{"Subpixel", &s.RetroFrame.SubpixelTint},
	}
	for _, sl := range sliders {
		nv, ch := dc.Slider(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY),
			sl.label, *sl.ptr, 0, 0.25, retroOpts)
		if ch {
			*sl.ptr = nv
			game.Flags().Set("retro_frame_reset")
			o.markDirty()
		}
		cursorY -= l.RowH + l.RowGap
	}

	cursorY -= l.RowGap * 2
	return cursorY
}

// ---- SYSTEM -----------------------------------------------------------------

func (o *Overlay) renderSystemSection(dc *DrawCtx, panel pixel.Rect, cursorY float64, l pl) float64 {
	x0, w := panel.Min.X+l.Pad, panel.W()-l.Pad*2

	dc.SectionHeader(pixel.R(x0, cursorY-l.HeaderH, x0+w, cursorY), "SYSTEM")
	cursorY -= l.HeaderH + l.RowGap

	if dc.Button(pixel.R(x0, cursorY-l.RowH, x0+w, cursorY), "Reset Device") {
		if o.hooks.Reset != nil {
			o.hooks.Reset()
		}
		o.panelOpen = false
	}
	cursorY -= l.RowH + l.RowGap

	return cursorY
}

// ---- dropdown popover -------------------------------------------------------

// canvasToWindow converts a canvas-space rect to window space given the content
// rect that the canvas is blitted into (canvas origin maps to contentRect.Min).
func (o *Overlay) canvasToWindow(r pixel.Rect, contentRect pixel.Rect) pixel.Rect {
	return pixel.R(
		contentRect.Min.X+r.Min.X,
		contentRect.Min.Y+r.Min.Y,
		contentRect.Min.X+r.Max.X,
		contentRect.Min.Y+r.Max.Y,
	)
}

// toggleDropdown opens the dropdown for the given id, or closes it if already open.
// The popover rect is computed once at open time: it prefers opening below the anchor
// row (lower Y = lower on screen) and flips upward if there is not enough room.
func (o *Overlay) toggleDropdown(id string, windowAnchor pixel.Rect, opts []string, sel int, onSel func(int)) {
	if o.dropdownID == id {
		o.dropdownID = ""
		return
	}
	ps := func(v float64) float64 { return math.Floor(v * o.panelScale) }
	n := len(opts)
	h := float64(n)*ps(baseDropdownItemH) + ps(baseDropdownPad)*2
	panelBot := chromePad + chromeSize + chromeGap // lowest Y the panel reaches

	var rect pixel.Rect
	gap := ps(baseDropdownGap)
	if windowAnchor.Min.Y-gap-h >= panelBot {
		// open downward (below anchor in screen terms)
		top := windowAnchor.Min.Y - gap
		rect = pixel.R(windowAnchor.Min.X, top-h, windowAnchor.Max.X, top)
	} else {
		// flip upward (above anchor in screen terms)
		bot := windowAnchor.Max.Y + gap
		rect = pixel.R(windowAnchor.Min.X, bot, windowAnchor.Max.X, bot+h)
	}

	o.dropdownID = id
	o.dropdownAnchor = windowAnchor
	o.dropdownRect = rect
	o.dropdownOpts = opts
	o.dropdownSel = sel
	o.dropdownOnSel = onSel
	o.dropdownJustOpened = true
}

// dropdownPopoverRect returns the precomputed window-space rect for the open popover.
func (o *Overlay) dropdownPopoverRect() pixel.Rect {
	return o.dropdownRect
}

func (o *Overlay) renderDropdownPopover(dc *DrawCtx) {
	// Consume and clear the just-opened flag so the close check below is skipped
	// for the single frame where the same JustUp event that opened the dropdown
	// would otherwise immediately close it.
	justOpened := o.dropdownJustOpened
	o.dropdownJustOpened = false

	r := o.dropdownRect

	dc.fillRect(r, dc.fade(colBg))
	borderCol := dc.fade(colTrack)
	dc.fillRect(pixel.R(r.Min.X, r.Min.Y, r.Max.X, r.Min.Y+1), borderCol)
	dc.fillRect(pixel.R(r.Min.X, r.Max.Y-1, r.Max.X, r.Max.Y), borderCol)
	dc.fillRect(pixel.R(r.Min.X, r.Min.Y, r.Min.X+1, r.Max.Y), borderCol)
	dc.fillRect(pixel.R(r.Max.X-1, r.Min.Y, r.Max.X, r.Max.Y), borderCol)

	itemH := dc.sc(baseDropdownItemH)
	pad := dc.sc(baseDropdownPad)
	txt := newText()
	for i, opt := range o.dropdownOpts {
		top := r.Max.Y - pad - float64(i)*itemH
		optR := pixel.R(r.Min.X+pad, top-itemH, r.Max.X-pad, top)
		isSelected := i == o.dropdownSel
		hover := contains(optR, dc.Ptr.Pos)

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

		txt.Clear()
		if isSelected {
			txt.Color = colors.LayerAlpha(colors.HexString("#101010"), dc.A)
		} else {
			txt.Color = dc.fade(colText)
		}
		txt.WriteString(opt)
		tw := txt.Bounds().W() * dc.ts()
		cx := math.Floor(optR.Min.X + (optR.W()-tw)/2)
		cy := optR.Min.Y + (optR.H()-dc.lh())/2
		dc.drawText(txt, pixel.V(cx, cy))

		if hover && dc.Ptr.JustUp {
			o.dropdownOnSel(i)
			o.dropdownSel = i
			o.dropdownID = ""
		}
	}

	// Close on any click outside the popover. Skip on the opening frame so the
	// JustUp that triggered toggleDropdown doesn't immediately re-close it.
	if !justOpened && o.dropdownID != "" && dc.Ptr.JustUp && !contains(r, dc.Ptr.Pos) {
		o.dropdownID = ""
	}
}

// ---- helpers ----------------------------------------------------------------

func indexOf[T comparable](slice []T, v T) int {
	for i, s := range slice {
		if s == v {
			return i
		}
	}
	return 0
}
