package overlay

import (
	"fmt"
	"math"

	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
)

// ---- Toggle -----------------------------------------------------------------

// Toggle draws a labelled on/off toggle. Returns (newValue, changed).
func (dc *DrawCtx) Toggle(r pixel.Rect, label string, value bool) (bool, bool) {
	clicked := dc.clicked(r)
	newValue := value
	if clicked {
		newValue = !value
	}

	dc.rowBg(r)

	// pill (rounded 2px frame, stretched)
	pillW := dc.sc(30)
	pillH := dc.sc(14)
	px := math.Floor(r.Max.X - pillW - dc.sc(8))
	py := math.Floor(r.Center().Y - pillH/2)
	pillR := pixel.R(px, py, px+pillW, py+pillH)
	pillCol := colTrack
	if newValue {
		pillCol = colAccent
	}
	dc.drawRoundedRect(pillR, dc.fade(pillCol))

	// knob — a smaller rounded square on the pill, slid left or right
	knobSize := dc.sc(10)
	innerPad := dc.sc(2)
	var kx float64
	if newValue {
		kx = pillR.Max.X - innerPad - knobSize
	} else {
		kx = pillR.Min.X + innerPad
	}
	ky := math.Floor(pillR.Center().Y - knobSize/2)
	knobCol := colors.LayerAlpha(colors.HexString("#101010"), dc.A)
	if !newValue {
		knobCol = colors.LayerAlpha(colors.HexString("#808080"), dc.A)
	}
	dc.drawRoundedRect(pixel.R(kx, ky, kx+knobSize, ky+knobSize), knobCol)

	txt := newText()
	txt.Color = dc.fade(colText)
	txt.WriteString(label)
	dc.drawText(txt, pixel.V(r.Min.X+dc.sc(8), r.Center().Y-dc.lh()/2))

	return newValue, clicked
}

// ---- Slider -----------------------------------------------------------------

// SliderOpts configures optional Slider behaviour.
type SliderOpts struct {
	// Format is a printf format string applied to the raw value.
	// If empty, the slider displays "NN%" (value * 100, rounded).
	Format string
}

// Slider draws a horizontal slider: [Label] [====●====] [XX%]
// Returns (newValue, changed).
func (dc *DrawCtx) Slider(r pixel.Rect, label string, value, min, max float64, opts ...SliderOpts) (float64, bool) {
	var o SliderOpts
	if len(opts) > 0 {
		o = opts[0]
	}
	labelColW := dc.sc(120)
	valColW := dc.sc(44)
	pad := dc.sc(6)
	trackX0 := r.Min.X + labelColW + pad
	trackX1 := r.Max.X - valColW - pad
	trackLen := trackX1 - trackX0
	trackY := r.Center().Y

	changed := false
	newValue := value
	if dc.Ptr.Down && contains(r, dc.Ptr.Pos) {
		t := (dc.Ptr.Pos.X - trackX0) / trackLen
		t = math.Max(0, math.Min(1, t))
		newValue = min + t*(max-min)
		changed = true
	}

	dc.rowBg(r)

	trackH := dc.sc(3)
	dc.fillRect(pixel.R(trackX0, trackY-trackH/2, trackX1, trackY+trackH/2),
		dc.fade(colTrack))

	frac := (newValue - min) / (max - min)
	fillX := trackX0 + frac*trackLen
	if fillX > trackX0 {
		dc.fillRect(pixel.R(trackX0, trackY-trackH/2, fillX, trackY+trackH/2),
			dc.fade(colAccent))
	}
	dc.circle(pixel.V(fillX, trackY), dc.sc(5), colors.LayerAlpha(colors.Alpha(1), dc.A))

	txt := newText()
	txt.Color = dc.fade(colText)
	txt.WriteString(label)
	dc.drawText(txt, pixel.V(r.Min.X+dc.sc(8), r.Center().Y-dc.lh()/2))

	txt.Clear()
	txt.Color = dc.fade(colDim)
	var valStr string
	if o.Format != "" {
		valStr = fmt.Sprintf(o.Format, newValue)
	} else {
		valStr = fmt.Sprintf("%.0f%%", newValue*100)
	}
	txt.WriteString(valStr)
	tw := txt.Bounds().W() * dc.ts()
	dc.drawText(txt, pixel.V(r.Max.X-valColW+(valColW-tw)/2, r.Center().Y-dc.lh()/2))

	return newValue, changed
}

// ---- VerticalSlider ---------------------------------------------------------

// VerticalSlider draws a slim vertical slider with the knob at the current value.
// Bottom of the rect is min, top is max. Returns (newValue, changed).
func (dc *DrawCtx) VerticalSlider(r pixel.Rect, value, min, max float64) (float64, bool) {
	pad := dc.sc(10)
	trackY0 := r.Min.Y + pad
	trackY1 := r.Max.Y - pad
	trackLen := trackY1 - trackY0
	trackX := r.Center().X

	changed := false
	newValue := value
	if dc.Ptr.Down && contains(r, dc.Ptr.Pos) {
		t := (dc.Ptr.Pos.Y - trackY0) / trackLen
		t = math.Max(0, math.Min(1, t))
		newValue = min + t*(max-min)
		changed = true
	}

	// panel background
	dc.fillRect(r, dc.fade(colBg))

	// track groove
	trackW := dc.sc(3)
	dc.fillRect(pixel.R(trackX-trackW/2, trackY0, trackX+trackW/2, trackY1),
		dc.fade(colTrack))

	// filled portion from bottom up to current value
	frac := (newValue - min) / (max - min)
	fillY := trackY0 + frac*trackLen
	if fillY > trackY0 {
		dc.fillRect(pixel.R(trackX-trackW/2, trackY0, trackX+trackW/2, fillY),
			dc.fade(colAccent))
	}

	// knob
	dc.circle(pixel.V(trackX, fillY), dc.sc(5), colors.LayerAlpha(colors.Alpha(1), dc.A))

	return newValue, changed
}

// ---- SegmentedSelect --------------------------------------------------------

// SegmentedSelect draws a labelled row with N option buttons.
// Returns (selectedIndex, changed).
func (dc *DrawCtx) SegmentedSelect(r pixel.Rect, label string, options []string, selected int) (int, bool) {
	dc.rowBg(r)

	txt := newText()
	txt.Color = dc.fade(colDim)
	txt.WriteString(label)
	dc.drawText(txt, pixel.V(r.Min.X+dc.sc(8), r.Center().Y-dc.lh()/2))

	n := len(options)
	if n == 0 {
		return selected, false
	}

	btnAreaW := r.W() * 0.55
	btnW := (btnAreaW - float64(n-1)*dc.sc(2)) / float64(n)
	btnH := r.H() - dc.sc(8)
	bx := r.Max.X - btnAreaW
	by := r.Min.Y + dc.sc(4)

	newSelected := selected
	changed := false
	for i, opt := range options {
		btnR := pixel.R(bx, by, bx+btnW, by+btnH)
		hover := contains(btnR, dc.Ptr.Pos)
		isSelected := i == selected

		var bg pixel.RGBA
		switch {
		case isSelected:
			bg = dc.fade(colAccent)
		case hover && dc.Ptr.Down:
			bg = dc.fade(colPress)
		case hover:
			bg = dc.fade(colHover)
		default:
			bg = colors.LayerAlpha(colors.HexString("#262626"), dc.A)
		}
		dc.fillRect(btnR, bg)

		txt.Clear()
		if isSelected {
			txt.Color = colors.LayerAlpha(colors.HexString("#101010"), dc.A)
		} else {
			txt.Color = dc.fade(colText)
		}
		txt.WriteString(opt)
		tw := txt.Bounds().W() * dc.ts()
		cx := btnR.Min.X + (btnR.W()-tw)/2
		cy := btnR.Min.Y + (btnR.H()-dc.lh())/2
		dc.drawText(txt, pixel.V(cx, cy))

		if hover && dc.Ptr.JustUp {
			newSelected = i
			changed = true
		}
		bx += btnW + dc.sc(2)
	}
	return newSelected, changed
}

// ---- DropdownSelectRow ------------------------------------------------------

// DropdownSelectRow draws a labelled row showing the current selection and an
// open/close indicator ("v"/"^"). Returns true when clicked; the caller manages
// open/close state and renders the actual popover separately.
func (dc *DrawCtx) DropdownSelectRow(r pixel.Rect, label, currentValue string, isOpen bool) bool {
	if isOpen {
		dc.fillRect(r, dc.fade(colHover))
	} else {
		dc.rowBg(r)
	}

	txt := newText()
	txt.Color = dc.fade(colDim)
	txt.WriteString(label)
	dc.drawText(txt, pixel.V(r.Min.X+dc.sc(8), r.Center().Y-dc.lh()/2))

	indicator := "v"
	if isOpen {
		indicator = "^"
	}
	txt.Clear()
	txt.Color = dc.fade(colText)
	txt.WriteString(currentValue + " " + indicator)
	tw := txt.Bounds().W() * dc.ts()
	dc.drawText(txt, pixel.V(math.Floor(r.Max.X-tw-dc.sc(8)), r.Center().Y-dc.lh()/2))

	return dc.clicked(r)
}

// ---- Button -----------------------------------------------------------------

// Button draws a full-row text button. Returns true if clicked.
func (dc *DrawCtx) Button(r pixel.Rect, label string) bool {
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

	txt := newText()
	txt.Color = dc.fade(colText)
	txt.WriteString(label)
	tw := txt.Bounds().W() * dc.ts()
	cx := math.Floor(r.Min.X + (r.W()-tw)/2)
	cy := math.Floor(r.Min.Y + (r.H()-dc.lh())/2)
	dc.drawText(txt, pixel.V(cx, cy))

	return dc.clicked(r)
}

// ---- Section header ---------------------------------------------------------

func (dc *DrawCtx) SectionHeader(r pixel.Rect, label string) {
	dc.SectionHeaderWithReset(r, label, nil)
}

// SectionHeaderWithReset renders the section label, plus a small reset icon
// on the right if onReset is non-nil. Clicking the icon calls onReset.
func (dc *DrawCtx) SectionHeaderWithReset(r pixel.Rect, label string, onReset func()) {
	txt := newText()
	txt.Color = dc.fade(colDim)
	txt.WriteString(label)
	dc.drawText(txt, pixel.V(r.Min.X+dc.sc(8), r.Center().Y-dc.lh()/2))

	if onReset == nil {
		return
	}

	btnSize := dc.sc(16)
	btnYMin := math.Floor(r.Center().Y - btnSize/2)
	btnR := pixel.R(r.Max.X-btnSize-dc.sc(2), btnYMin, r.Max.X-dc.sc(2), btnYMin+btnSize)

	hover := contains(btnR, dc.Ptr.Pos)
	if hover {
		bg := dc.fade(colHover)
		if dc.Ptr.Down {
			bg = dc.fade(colPress)
		}
		dc.fillRect(btnR, bg)
	}
	dc.drawIcon("reset", btnR.Center(), btnSize)

	if hover && dc.Ptr.JustUp {
		onReset()
	}
}
