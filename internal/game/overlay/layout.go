package overlay

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/rpg"
	"github.com/gopxl/pixel/v2"
)

// MinGamepadAreaH is the gamepad block height at uiScale = 1.0 (reference width).
const MinGamepadAreaH = 220.0

// gamepadReferenceW is the window width at which uiScale = 1.0. Below this the
// scale is clamped to 1.0; above it grows linearly up to the cap.
const gamepadReferenceW = 420.0

// IsPortrait returns true when the window is taller than wide.
func IsPortrait(wb pixel.Rect) bool { return wb.H() > wb.W() }

// OverlayUIScale returns a scale multiplier for all overlay UI elements (chrome
// buttons and, in portrait mode, the virtual gamepad). Applies in both
// landscape and portrait so the corner chrome grows on large windows too.
func OverlayUIScale(wb pixel.Rect) float64 {
	s := wb.W() / gamepadReferenceW
	return math.Max(1.0, math.Min(2.0, s))
}

// GamepadUIScale returns a scale multiplier for the virtual gamepad UI.
// Same as OverlayUIScale in portrait; always 1.0 in landscape.
func GamepadUIScale(wb pixel.Rect) float64 {
	if !IsPortrait(wb) {
		return 1.0
	}
	return OverlayUIScale(wb)
}

// ChromeAreaBottom returns the window-space Y just above the corner chrome
// buttons, accounting for the scaled button size. Used as the panel's bottom
// boundary and the portrait bottom-clearance budget.
func ChromeAreaBottom(wb pixel.Rect) float64 {
	us := OverlayUIScale(wb)
	return math.Floor((chromePad + chromeSize + chromeGap) * us)
}

// GamepadAreaH returns the scaled height of the virtual gamepad block.
func GamepadAreaH(wb pixel.Rect) float64 {
	return math.Floor(MinGamepadAreaH * GamepadUIScale(wb))
}

// portraitChromeH returns the height reserved at the bottom for the chrome in
// portrait mode (slightly less than ChromeAreaBottom - no gap needed above chrome).
func portraitChromeH(wb pixel.Rect) float64 {
	us := OverlayUIScale(wb)
	return math.Floor((chromePad + chromeSize) * us)
}

// portraitGap returns the equal gap used above the canvas and between the canvas
// and gamepad. The bottom gap is portraitGap + portraitChromeH so all three
// visible spacings look equal while still clearing the corner chrome.
func portraitGap(wb pixel.Rect, canvasH float64) float64 {
	return math.Max(0, (wb.H()-canvasH-GamepadAreaH(wb)-portraitChromeH(wb))/3)
}

// GameCanvasScale returns the integer pixel scale for the 240x160 game canvas.
func GameCanvasScale(wb pixel.Rect) float64 {
	availH := wb.H()
	if IsPortrait(wb) {
		// Reserve space for gamepad block + chrome clearance.
		availH -= GamepadAreaH(wb) + portraitChromeH(wb)
	}
	maxFit := math.Min(
		math.Floor(wb.W()/game.GameWidth),
		math.Floor(availH/game.GameHeight),
	)
	if maxFit < 1 {
		maxFit = 1
	}
	d := game.CurrentSave().SystemSettings.Display
	if d != nil && d.ScaleMode == rpg.ScaleModeFixed && d.FixedScale > 0 {
		fixed := float64(d.FixedScale)
		if fixed > maxFit {
			fixed = maxFit
		}
		return fixed
	}
	return maxFit
}

// CanvasCenter returns the window-space center for compositing the game canvas.
// In portrait mode all three visible gaps are equal; canvas center Y = H - gap - canvasH/2.
func CanvasCenter(wb pixel.Rect, scale float64) pixel.Vec {
	if IsPortrait(wb) {
		canvasH := float64(game.GameHeight) * scale
		gap := portraitGap(wb, canvasH)
		cy := math.Floor(wb.H() - gap - canvasH/2)
		return pixel.V(math.Floor(wb.W()/2), cy)
	}
	return pixel.V(math.Floor(wb.Center().X), math.Floor(wb.Center().Y))
}
