package overlay

import (
	"math"

	"fisherevans.com/project/f/internal/game"
	"fisherevans.com/project/f/internal/game/input"
	"fisherevans.com/project/f/internal/game/rpg"
	"fisherevans.com/project/f/internal/resources"
	"fisherevans.com/project/f/internal/util/colors"
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/text"
)

// All gamepad sprites rendered at the same pixel scale: source px * gamepadRenderScale.
//   D-pad source tiles: 32x32 → 128px on screen
//   Button source tiles: 16x16 → 64px on screen
const (
	gamepadRenderScale = 4.0
	dpadRendered       = 32.0 * gamepadRenderScale // 128
	btnRendered        = 16.0 * gamepadRenderScale // 64
	labelScale         = 2.0                       // m3x6 font rendered at 2x

	// Portrait-specific spacing — more breathing room for thumbs.
	portraitEdgePad = 40.0 // from screen edges to dpad/A centres
	portraitBBtnGap = 28.0 // gap between A right edge and B left edge (wider than landscape)

	// Landscape spacing
	landscapeEdgePad = 20.0
	landscapeBtnGap  = 8.0

	labelGap    = 6.0  // px between button bottom and top of label text
	pillBtnDrop = 14.0 // how many px lower the pill buttons sit relative to their label anchor
	dpadHitPad  = 28.0 // extra px beyond sprite edge added to dpad hit rect on each side
)

// gamepadLabelFont is the m3x6 font atlas used for START/SELECT labels.
var gamepadLabelFont *text.Atlas

func init() { initGamepadFont() }

func initGamepadFont() {
	resources.RunOnceInitialized(func() {
		fi := resources.DefaultAtlas().GetFont(resources.FontNameM3x6)
		gamepadLabelFont = fi.Atlas
	})
}

// gamepadOpacity returns the user-configured opacity for the virtual gamepad.
func (o *Overlay) gamepadOpacity() float64 {
	save := game.CurrentSave()
	if save == nil || save.SystemSettings.Display == nil {
		return 0.85
	}
	op := save.SystemSettings.Display.VirtualGamepadOpacity
	if op <= 0 {
		return 0.85
	}
	return op
}

// gamepadCenters holds the computed button centres for one frame.
// dpadSize / btnSize are the scaled rendered pixel sizes used for both drawing
// and hit detection so both stay in sync.
type gamepadCenters struct {
	dpad, btnA, btnB, start, sel pixel.Vec
	startLabelY, selLabelY       float64
	dpadSize, btnSize            float64
}

// isGamepadVisible returns whether the virtual gamepad should be shown.
// "auto" shows in portrait windows (H > W) and hides in landscape.
func (o *Overlay) isGamepadVisible(wb pixel.Rect) bool {
	if o.panelOpen {
		return false
	}
	setting := rpg.VirtualGamepadAuto
	if save := game.CurrentSave(); save != nil && save.SystemSettings.Display != nil {
		setting = save.SystemSettings.Display.VirtualGamepad
	}
	switch setting {
	case rpg.VirtualGamepadOn:
		return true
	case rpg.VirtualGamepadOff:
		return false
	default: // auto
		return IsPortrait(wb)
	}
}

// gamepadLayout computes button centres for the current window and canvas scale.
func gamepadLayout(wb pixel.Rect, scale float64) gamepadCenters {
	if IsPortrait(wb) {
		return portraitLayout(wb, scale)
	}
	return landscapeLayout(wb)
}

// landscapeLayout positions buttons in the lower corners above the corner chrome.
func landscapeLayout(wb pixel.Rect) gamepadCenters {
	base := gamepadBaseY(wb)
	halfDpad := dpadRendered / 2
	halfBtn := btnRendered / 2

	dpadC := pixel.V(
		math.Floor(landscapeEdgePad+halfDpad),
		math.Floor(base+halfDpad),
	)
	// A is the primary right button — higher (larger Y = higher on screen).
	aCenter := pixel.V(
		math.Floor(wb.Max.X-landscapeEdgePad-halfBtn),
		math.Floor(base+halfBtn+halfBtn*0.25),
	)
	// B is left of A and lower.
	bCenter := pixel.V(
		math.Floor(aCenter.X-btnRendered-landscapeBtnGap),
		math.Floor(base+halfBtn-halfBtn*0.25),
	)
	startC := pixel.V(
		math.Floor(wb.Center().X+8+halfBtn),
		math.Floor(base+halfBtn),
	)
	selC := pixel.V(
		math.Floor(wb.Center().X-8-halfBtn),
		math.Floor(base+halfBtn),
	)
	labelY := base + halfBtn
	return gamepadCenters{
		dpad: dpadC, btnA: aCenter, btnB: bCenter, start: startC, sel: selC,
		startLabelY: labelY, selLabelY: labelY,
		dpadSize: dpadRendered, btnSize: btnRendered,
	}
}

// gamepadBaseY clears the corner chrome for the landscape layout.
func gamepadBaseY(wb pixel.Rect) float64 {
	us := OverlayUIScale(wb)
	return math.Floor((chromePad+chromeSize)*us + landscapeEdgePad)
}

// portraitLayout distributes buttons within the gamepad block.
// All sizes are scaled by GamepadUIScale so the buttons remain proportional
// on large windows (e.g. desktop browser) as well as phone-size viewports.
func portraitLayout(wb pixel.Rect, scale float64) gamepadCenters {
	canvasH := float64(game.GameHeight) * scale
	topMargin := portraitGap(wb, canvasH) + portraitChromeH(wb)
	areaH := GamepadAreaH(wb)
	uiScale := GamepadUIScale(wb)

	// Snap to exact integer multiples of source tile sizes (32px dpad, 16px buttons)
	// so drawGamepadSprite's floor-divided scale produces zero sub-pixel error.
	dpad := 32.0 * math.Max(1, math.Floor(dpadRendered*uiScale/32.0))
	btn := 16.0 * math.Max(1, math.Floor(btnRendered*uiScale/16.0))
	halfDpad := dpad / 2
	halfBtn := btn / 2
	edgePad := math.Floor(portraitEdgePad * uiScale)
	bGap := math.Floor(portraitBBtnGap * uiScale)
	pDrop := math.Floor(pillBtnDrop * uiScale)

	mainY := math.Floor(topMargin + areaH*0.62)
	menuY := math.Floor(topMargin + areaH*0.25)

	dpadC := pixel.V(math.Floor(edgePad+halfDpad), mainY)

	aCenter := pixel.V(
		math.Floor(wb.Max.X-edgePad-halfBtn),
		math.Floor(mainY+halfBtn*0.25),
	)
	bCenter := pixel.V(
		math.Floor(aCenter.X-btn-bGap),
		math.Floor(mainY-halfBtn*0.35),
	)

	startC := pixel.V(math.Floor(wb.Center().X+8+halfBtn), math.Floor(menuY-pDrop))
	selC := pixel.V(math.Floor(wb.Center().X-8-halfBtn), math.Floor(menuY-pDrop))

	return gamepadCenters{
		dpad: dpadC, btnA: aCenter, btnB: bCenter, start: startC, sel: selC,
		startLabelY: menuY, selLabelY: menuY,
		dpadSize: dpad, btnSize: btn,
	}
}

func squareHitRect(center pixel.Vec, size float64) pixel.Rect {
	h := size / 2
	return pixel.R(center.X-h, center.Y-h, center.X+h, center.Y+h)
}

// computeGamepadInput reads the current pointer and returns the virtual button
// state. Called from UpdateInput so it is available before game.UpdateControls.
func (o *Overlay) computeGamepadInput(win *opengl.Window) input.VirtualState {
	wb := win.Bounds()
	if !o.isGamepadVisible(wb) || !o.pointer.Down {
		return input.VirtualState{}
	}

	scale := GameCanvasScale(wb)
	pos := gamepadLayout(wb, scale)
	ptr := o.pointer.Pos
	var vs input.VirtualState

	// D-pad: dominant-axis quadrant, no dead zone. Hit area extends dpadHitPad
	// beyond the sprite edge so a sliding thumb stays registered.
	if contains(squareHitRect(pos.dpad, pos.dpadSize+dpadHitPad*2), ptr) {
		off := ptr.Sub(pos.dpad)
		if math.Abs(off.X) >= math.Abs(off.Y) {
			if off.X >= 0 {
				vs.Dir = input.Right
			} else {
				vs.Dir = input.Left
			}
		} else {
			if off.Y >= 0 {
				vs.Dir = input.Up
			} else {
				vs.Dir = input.Down
			}
		}
	}

	vs.A = contains(squareHitRect(pos.btnA, pos.btnSize), ptr)
	vs.B = contains(squareHitRect(pos.btnB, pos.btnSize), ptr)
	vs.Start = contains(squareHitRect(pos.start, pos.btnSize), ptr)
	vs.Select = contains(squareHitRect(pos.sel, pos.btnSize), ptr)

	return vs
}

// renderGamepad draws the on-screen gamepad using the virtualState computed
// during UpdateInput. Uses its own opacity from settings, independent of the
// overlay's idle-fade alpha.
func (o *Overlay) renderGamepad(win *opengl.Window) {
	wb := win.Bounds()
	if !o.isGamepadVisible(wb) {
		return
	}

	scale := GameCanvasScale(wb)
	pos := gamepadLayout(wb, scale)
	vs := o.virtualState

	dc := &DrawCtx{
		target: win,
		atlas:  overlayAtlas,
		imd:    o.imd,
		Ptr:    o.pointer,
		A:      o.gamepadOpacity(),
	}

	drawGamepadSprite(dc, "overlay/gamepad_dpad:"+dpadSpriteName(vs.Dir), pos.dpad, pos.dpadSize)

	aSprite := "btn_a"
	if vs.A {
		aSprite = "btn_a_press"
	}
	drawGamepadSprite(dc, "overlay/gamepad:"+aSprite, pos.btnA, pos.btnSize)

	bSprite := "btn_b"
	if vs.B {
		bSprite = "btn_b_press"
	}
	drawGamepadSprite(dc, "overlay/gamepad:"+bSprite, pos.btnB, pos.btnSize)

	startSprite := "btn_start"
	if vs.Start {
		startSprite = "btn_start_press"
	}
	drawGamepadSprite(dc, "overlay/gamepad:"+startSprite, pos.start, pos.btnSize)
	drawGamepadLabel(dc, "START", pos.start.X, pos.startLabelY, pos.btnSize)

	selectSprite := "btn_select"
	if vs.Select {
		selectSprite = "btn_select_press"
	}
	drawGamepadSprite(dc, "overlay/gamepad:"+selectSprite, pos.sel, pos.btnSize)
	drawGamepadLabel(dc, "SELECT", pos.sel.X, pos.selLabelY, pos.btnSize)
}

// drawGamepadLabel draws a small outlined label centred below a button, scaled 2x.
// anchorY is the fixed vertical anchor; btnSize is the scaled button pixel size
// used to compute clearance below the button sprite.
func drawGamepadLabel(dc *DrawCtx, label string, centerX, anchorY, btnSize float64) {
	if gamepadLabelFont == nil {
		return
	}
	txt := text.New(pixel.ZV, gamepadLabelFont)
	txt.WriteString(label)

	// Scale the bounds to account for labelScale.
	w := txt.Bounds().W() * labelScale
	lh := txt.LineHeight * labelScale
	x := math.Floor(centerX - w/2)
	baselineY := math.Floor(anchorY - btnSize/2 - labelGap - lh)

	outlineCol := colors.LayerAlpha(colors.HexString("#000000"), dc.A)
	whiteCol := colors.LayerAlpha(colors.Alpha(1), dc.A)

	for _, off := range [4]pixel.Vec{
		{X: 1}, {X: -1}, {Y: 1}, {Y: -1},
	} {
		txt.Clear()
		txt.Color = outlineCol
		txt.WriteString(label)
		txt.Draw(dc.target, pixel.IM.Scaled(pixel.ZV, labelScale).Moved(pixel.V(x+off.X, baselineY+off.Y)))
	}
	txt.Clear()
	txt.Color = whiteCol
	txt.WriteString(label)
	txt.Draw(dc.target, pixel.IM.Scaled(pixel.ZV, labelScale).Moved(pixel.V(x, baselineY)))
}

func dpadSpriteName(dir input.Direction) string {
	switch dir {
	case input.Up:
		return "dpad_up"
	case input.Down:
		return "dpad_down"
	case input.Left:
		return "dpad_left"
	case input.Right:
		return "dpad_right"
	default:
		return "dpad"
	}
}

// drawGamepadSprite renders a sprite centred on center scaled to sizePx on its
// longest edge. Falls back to a dim circle placeholder while the sheet is undrawn.
func drawGamepadSprite(dc *DrawCtx, fullName string, center pixel.Vec, sizePx float64) {
	if overlayAtlas == nil {
		return
	}
	sprite := overlayAtlas.GetSprite(fullName)
	if sprite == nil {
		dc.circle(center, sizePx/2, colors.LayerAlpha(colTrack, dc.A*0.6))
		return
	}
	b := sprite.Bounds()
	longest := math.Max(b.W(), b.H())
	if longest <= 0 {
		return
	}
	scale := math.Max(1, math.Floor(sizePx/longest))
	m := pixel.IM.Scaled(pixel.ZV, scale).Moved(center)
	sprite.DrawColorMask(dc.target, m, colors.LayerAlpha(colors.Alpha(1), dc.A))
}
