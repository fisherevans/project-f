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

const (
	panelW   = 320.0
	panelPad = 14.0
	rowH     = 30.0
	rowGap   = 3.0
	headerH  = 22.0
)

const (
	scrollbarW     = 8.0
	scrollbarGap   = 4.0
	scrollWheelPx  = 30.0 // pixels per wheel tick
	minThumbH      = 20.0
)

func (o *Overlay) renderSettingsPanel(win *opengl.Window) {
	// ---- 1. compute panel rect in window space (integer-aligned) -------------
	wb := win.Bounds()
	panelX := math.Floor(wb.Max.X - panelW - chromePad)
	panelMaxY := math.Floor(wb.Max.Y - chromePad)
	panel := pixel.R(panelX, chromePad+chromeSize+chromeGap, panelX+panelW, panelMaxY)

	// outer DC: panel bg, border, scrollbar, close-on-outside
	outerDC := &DrawCtx{
		target: win,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    o.pointer,
		A:      o.alpha,
	}

	outerDC.fillRect(panel, colors.LayerAlpha(colors.HexString("#0d0d0d"), 0.94*o.alpha))
	outerDC.fillRect(pixel.R(panel.Min.X, panel.Min.Y, panel.Min.X+1, panel.Max.Y),
		colors.LayerAlpha(colors.HexString("#333333"), o.alpha))

	// close on click outside the panel (scrollbar included in panel rect)
	if o.pointer.JustUp && !contains(panel, o.pointer.Pos) && !o.scrollDragging {
		o.panelOpen = false
		return
	}

	// ---- 2. split panel into content area + scrollbar rail -------------------
	// Canvas dims must be EVEN or the canvas center lands on a half-pixel when
	// blitted to the window, producing sub-pixel sampling and ghost trails
	// (especially visible during scroll).
	rawContentW := (panel.Max.X - scrollbarW - scrollbarGap) - (panel.Min.X + 1)
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
		panel.Max.X-scrollbarW-scrollbarGap/2,
		panel.Min.Y+panelPad,
		panel.Max.X-scrollbarGap/2,
		panel.Max.Y-panelPad,
	)

	// ---- 3. update scroll offset from input, clamped against last-frame's
	// measured content height (applied again after this frame measures) ------
	o.scrollMax = math.Max(0, o.scrollContentH-visibleH)

	if contains(panel, o.pointer.Pos) {
		if dy := win.MouseScroll().Y; dy != 0 {
			o.scrollOffset -= dy * scrollWheelPx
		}
	}

	if o.scrollDragging {
		if !o.pointer.Down {
			o.scrollDragging = false
		} else if o.scrollContentH > visibleH {
			thumbH := math.Max(minThumbH, scrollbarTrack.H()*visibleH/o.scrollContentH)
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
	// Suppress clicks outside the visible content area (hits the scrollbar or
	// escapes the panel) and while scrollbar drag is active, so widgets don't
	// accidentally react to a held pointer that belongs to the scrollbar.
	if o.scrollDragging || !contains(contentRect, o.pointer.Pos) {
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
	}

	// ---- 7. render sections into the canvas ---------------------------------
	// cursor is canvas-local Y, starting just inside the top edge, shifted up
	// by scrollOffset so larger offsets pull content into view from below.
	startCursor := visibleH - panelPad + o.scrollOffset
	cursor := startCursor
	cursor = o.renderAudioSection(innerDC, bounds, cursor)
	cursor = o.renderDisplaySection(innerDC, bounds, cursor, win)
	cursor = o.renderScenesSection(innerDC, bounds, cursor)
	cursor = o.renderDebugSection(innerDC, bounds, cursor)
	cursor = o.renderSystemSection(innerDC, bounds, cursor)

	// Measure intrinsic content height for next-frame clamping. startCursor and
	// cursor both include +scrollOffset so the difference is scroll-independent;
	// add 2*panelPad for top + bottom breathing room around the content.
	totalConsumed := startCursor - cursor
	o.scrollContentH = totalConsumed + 2*panelPad
	o.scrollMax = math.Max(0, o.scrollContentH-visibleH)

	// ---- 8. blit canvas into window at content rect -------------------------
	o.scrollCanvas.Draw(win, pixel.IM.Moved(contentRect.Center()))

	// ---- 9. draw scrollbar (only if content overflows) ----------------------
	if o.scrollMax > 0 {
		outerDC.fillRect(scrollbarTrack,
			colors.LayerAlpha(colors.HexString("#202020"), 0.6*o.alpha))

		thumbH := math.Max(minThumbH, scrollbarTrack.H()*visibleH/o.scrollContentH)
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

func (o *Overlay) renderAudioSection(dc *DrawCtx, panel pixel.Rect, cursorY float64) float64 {
	x0, w := panel.Min.X+panelPad, panel.W()-panelPad*2

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-headerH, x0+w, cursorY), "AUDIO", func() {
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
	cursorY -= headerH + rowGap

	newMuted, changed := dc.Toggle(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Mute", o.audioMuted)
	if changed {
		o.audioMuted = newMuted
		if o.hooks.SetAudio != nil {
			o.hooks.SetAudio(o.audioMuted, o.audioVolume)
		}
		o.markDirty()
	}
	cursorY -= rowH + rowGap

	sliderDC := *dc
	if o.audioMuted {
		sliderDC.A = dc.A * 0.35
	}
	newVol, volChanged := sliderDC.Slider(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Volume", o.audioVolume, 0, 1)
	if volChanged && !o.audioMuted {
		o.setVolume(newVol)
	}
	cursorY -= rowH + rowGap*3

	return cursorY
}

// ---- DISPLAY ----------------------------------------------------------------

func (o *Overlay) renderDisplaySection(dc *DrawCtx, panel pixel.Rect, cursorY float64, win *opengl.Window) float64 {
	x0, w := panel.Min.X+panelPad, panel.W()-panelPad*2

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-headerH, x0+w, cursorY), "DISPLAY", func() {
		// Reset scale to auto. Leave fullscreen as-is (toggling here would hit
		// the OS in a confusing way — users can toggle it explicitly).
		o.setScale(0)
	})
	cursorY -= headerH + rowGap

	newFS, fsChanged := dc.Toggle(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Fullscreen", o.displayFullscreen)
	if fsChanged {
		o.displayFullscreen = newFS
		if o.hooks.ToggleFullscreen != nil {
			o.hooks.ToggleFullscreen()
		}
		o.markDirty()
	}
	cursorY -= rowH + rowGap

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

	newIdx, scaleChanged := dc.SegmentedSelect(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Scale", opts, selectedIdx)
	if scaleChanged {
		o.setScale(newIdx)
	}
	cursorY -= rowH + rowGap*3

	return cursorY
}

// ---- SCENES -----------------------------------------------------------------

const scenesRowH = 22.0

func (o *Overlay) renderScenesSection(dc *DrawCtx, panel pixel.Rect, cursorY float64) float64 {
	if len(o.hooks.DevScenes) == 0 {
		return cursorY
	}
	x0, w := panel.Min.X+panelPad, panel.W()-panelPad*2

	dc.SectionHeader(pixel.R(x0, cursorY-headerH, x0+w, cursorY), "SCENES")
	cursorY -= headerH + rowGap

	for _, scene := range o.hooks.DevScenes {
		r := pixel.R(x0, cursorY-scenesRowH, x0+w, cursorY)
		if dc.Button(r, scene.Name) {
			game.SetActiveStateIntent(scene.Factory())
			o.panelOpen = false
		}
		cursorY -= scenesRowH + rowGap
	}
	cursorY -= rowGap * 2
	return cursorY
}

// ---- DEBUG ------------------------------------------------------------------

func (o *Overlay) renderDebugSection(dc *DrawCtx, panel pixel.Rect, cursorY float64) float64 {
	x0, w := panel.Min.X+panelPad, panel.W()-panelPad*2
	s := game.CurrentSave().SystemSettings

	dc.SectionHeaderWithReset(pixel.R(x0, cursorY-headerH, x0+w, cursorY), "DEBUG", func() {
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
	cursorY -= headerH + rowGap

	// Pathfinding debug
	newVal, changed := dc.Toggle(pixel.R(x0, cursorY-rowH, x0+w, cursorY),
		"Pathfinding", s.Debugging.ShowPathfindingDebugging)
	if changed {
		s.Debugging.ShowPathfindingDebugging = newVal
		o.markDirty()
	}
	cursorY -= rowH + rowGap

	// Lighting Mode
	lightingOpts := []string{"Blended", "Over", "Off"}
	lightingVals := []rpg.LightingMode{rpg.LightingModeBlended, rpg.LightingModeOver, rpg.LightingModeOff}
	lSel := indexOf(lightingVals, s.Lighting.LightingMode)
	newL, lChanged := dc.SegmentedSelect(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Lighting", lightingOpts, lSel)
	if lChanged {
		s.Lighting.LightingMode = lightingVals[newL]
		o.markDirty()
	}
	cursorY -= rowH + rowGap

	// Lighting Composition
	compOpts := []string{"Full", "Ambient"}
	compVals := []rpg.LightingComposition{rpg.LightingCompositionFull, rpg.LightingCompositionAmbient}
	cSel := indexOf(compVals, s.Lighting.LightingComposition)
	newC, cChanged := dc.SegmentedSelect(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Light Comp", compOpts, cSel)
	if cChanged {
		s.Lighting.LightingComposition = compVals[newC]
		o.markDirty()
	}
	cursorY -= rowH + rowGap

	// Bloom Mode
	bloomOpts := []string{"Blended", "Blur", "Threshold", "Off"}
	bloomVals := []rpg.BloomMode{rpg.BloomModeBlended, rpg.BloomModeOverBlurred, rpg.BloomModeOverThreshold, rpg.BloomModeOff}
	bSel := indexOf(bloomVals, s.Lighting.BloomMode)
	newB, bChanged := dc.SegmentedSelect(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Bloom", bloomOpts, bSel)
	if bChanged {
		s.Lighting.BloomMode = bloomVals[newB]
		o.markDirty()
	}
	cursorY -= rowH + rowGap

	// Pixel Grid (note: field is DisablePixelGrid, so UI shows inverse)
	pixelGridOn := !s.RetroFrame.DisablePixelGrid
	newPG, pgChanged := dc.Toggle(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Pixel Grid", pixelGridOn)
	if pgChanged {
		s.RetroFrame.DisablePixelGrid = !newPG
		game.Flags().Set("retro_frame_reset")
		o.markDirty()
	}
	cursorY -= rowH + rowGap

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
		nv, ch := dc.Slider(pixel.R(x0, cursorY-rowH, x0+w, cursorY),
			sl.label, *sl.ptr, 0, 0.25, retroOpts)
		if ch {
			*sl.ptr = nv
			game.Flags().Set("retro_frame_reset")
			o.markDirty()
		}
		cursorY -= rowH + rowGap
	}

	cursorY -= rowGap * 2
	return cursorY
}

// ---- SYSTEM -----------------------------------------------------------------

func (o *Overlay) renderSystemSection(dc *DrawCtx, panel pixel.Rect, cursorY float64) float64 {
	x0, w := panel.Min.X+panelPad, panel.W()-panelPad*2

	dc.SectionHeader(pixel.R(x0, cursorY-headerH, x0+w, cursorY), "SYSTEM")
	cursorY -= headerH + rowGap

	if dc.Button(pixel.R(x0, cursorY-rowH, x0+w, cursorY), "Reset Device") {
		if o.hooks.Reset != nil {
			o.hooks.Reset()
		}
		o.panelOpen = false
	}
	cursorY -= rowH + rowGap

	return cursorY
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
