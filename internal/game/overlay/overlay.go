package overlay

import (
	"github.com/gopxl/pixel/v2"
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"

	"fisherevans.com/project/f/internal/game/input"
)

func init() { initAtlas() }

const (
	idleHideSeconds = 1.8
	fadeDuration    = 0.4
)

// DevScene is one entry in the dev scene-select list.
type DevScene struct {
	Name    string
	Factory func() any
}

// Hooks are runtime callbacks the overlay invokes to apply settings.
type Hooks struct {
	Reset            func()
	SetAudio         func(muted bool, linearVolume float64)
	SetScale         func(mode string, fixedN int) // "auto"/0 or "fixed"/N
	ToggleFullscreen func()
	SaveSettings     func() // persist SystemSettings to disk; called when pointer released + dirty
	DevScenes        []DevScene
}

// Overlay is the hi-res UI layer rendered above the pixel-grid composite.
type Overlay struct {
	hooks Hooks
	imd   *imdraw.IMDraw

	pointer   Pointer
	idleTimer float64
	alpha     float64
	touchSeen bool

	panelOpen           bool
	panelJustClosed     bool // true for the one frame renderSettingsPanel closed the panel
	scalePopoverOpen    bool
	volumePopoverOpen      bool
	volumePopoverCloseTimer float64 // counts up while cursor is outside button+popover zone
	settingsDirty     bool
	settingsSaveTimer float64 // counts down to zero, then flushes to disk

	// settings panel layout scale (set each frame from OverlayUIScale)
	panelScale float64

	// settings panel scroll state
	scrollCanvas       *opengl.Canvas
	scrollOffset       float64 // 0 = content top flush with panel top
	scrollMax          float64 // updated each frame from measured content height
	scrollContentH     float64 // measured content height last frame
	scrollDragging     bool
	scrollDragStartY   float64 // window-space pointer Y when drag started
	scrollDragStartOff float64 // scrollOffset when drag started

	// audio state
	audioMuted  bool
	audioVolume float64

	// display state
	displayScale      int // 0 = auto, >0 = fixed scale N
	displayFullscreen bool

	// dropdown popover state (panel rows that need more options than SegmentedSelect fits)
	dropdownID          string     // non-empty = a dropdown is open
	dropdownAnchor      pixel.Rect // window-space anchor row (used for close-on-anchor-click)
	dropdownRect        pixel.Rect // window-space popover rect, computed at open time
	dropdownOpts        []string
	dropdownSel         int
	dropdownOnSel       func(int)
	dropdownJustOpened   bool // true for the one frame the dropdown was opened; skips same-frame close
	dropdownScrollIdx    int  // first visible item index when dropdown has overflow
	dropdownVisibleCount int  // number of items visible in dropdown (capped)

	// content drag-scroll state (touch drag anywhere in content area)
	contentDragging     bool
	contentScrolling    bool
	contentDragStartY   float64
	contentDragStartOff float64

	// fn keys section expand state
	fnKeysOpen bool

	// virtual gamepad state; computed in UpdateInput, consumed by runtime before UpdateControls
	virtualState input.VirtualState
}

func New(hooks Hooks) *Overlay {
	return &Overlay{
		hooks:       hooks,
		imd:         imdraw.New(nil),
		audioVolume: 1.0,
	}
}

// InitAudio seeds the overlay's audio state from persisted settings and applies
// the volume to the audio system immediately.
func (o *Overlay) InitAudio(muted bool, linearVolume float64) {
	o.audioMuted = muted
	o.audioVolume = linearVolume
}

// InitDisplay seeds the overlay's display state from persisted settings.
func (o *Overlay) InitDisplay(scale int, fullscreen bool) {
	o.displayScale = scale
	o.displayFullscreen = fullscreen
}

func (o *Overlay) UpdateInput(win *opengl.Window, dt float64) {
	prev := o.pointer
	o.pointer = pollPointer(win, prev)

	moved := o.pointer.Pos != prev.Pos || o.pointer.JustDown
	if moved {
		o.idleTimer = 0
	} else {
		o.idleTimer += dt
	}

	if o.panelOpen {
		o.idleTimer = 0
		o.alpha = 1 // panel must always be fully opaque; don't wait for fade-in
	}

	if o.idleTimer < idleHideSeconds {
		o.alpha = clamp01(o.alpha + dt/fadeDuration)
	} else {
		o.alpha = clamp01(o.alpha - dt/fadeDuration)
	}

	o.virtualState = o.computeGamepadInput(win)

	setCursorVisible(win, o.alpha > 0)
}

func (o *Overlay) Render(win *opengl.Window, dt float64) {
	// Persist settings even when the overlay is faded out.
	if o.settingsDirty {
		o.settingsSaveTimer -= dt
		if o.settingsSaveTimer <= 0 {
			if o.hooks.SaveSettings != nil {
				o.hooks.SaveSettings()
			}
			o.settingsDirty = false
		}
	}

	// Gamepad renders at its own persisted opacity, never suppressed by the idle fade.
	o.renderGamepad(win)

	if o.alpha <= 0 {
		return
	}
	if o.panelOpen {
		o.renderSettingsPanel(win)
	}
	o.renderCornerChrome(win, dt)
}

const settingsDebounceSeconds = 1.0

func (o *Overlay) markDirty() {
	o.settingsDirty = true
	o.settingsSaveTimer = settingsDebounceSeconds
}

// VirtualControls returns the virtual gamepad state computed during UpdateInput.
// Call this after UpdateInput and before game.UpdateControls each frame.
func (o *Overlay) VirtualControls() input.VirtualState { return o.virtualState }

func (o *Overlay) PointerConsumed() bool { return o.panelOpen }

// ---- audio actions ----------------------------------------------------------

func (o *Overlay) toggleMute() {
	o.audioMuted = !o.audioMuted
	if o.hooks.SetAudio != nil {
		o.hooks.SetAudio(o.audioMuted, o.audioVolume)
	}
	o.markDirty()
}

func (o *Overlay) setVolume(linear float64) {
	o.audioVolume = linear
	if o.audioMuted && linear > 0 {
		o.audioMuted = false
	}
	if o.hooks.SetAudio != nil {
		o.hooks.SetAudio(o.audioMuted, o.audioVolume)
	}
	o.markDirty()
}

// ---- display actions --------------------------------------------------------

func (o *Overlay) setScale(fixedN int) {
	o.displayScale = fixedN
	mode := "auto"
	if fixedN > 0 {
		mode = "fixed"
	}
	if o.hooks.SetScale != nil {
		o.hooks.SetScale(mode, fixedN)
	}
	o.markDirty()
}

func (o *Overlay) toggleFullscreen() {
	o.displayFullscreen = !o.displayFullscreen
	if o.hooks.ToggleFullscreen != nil {
		o.hooks.ToggleFullscreen()
	}
	o.markDirty()
}

// ---- helpers ----------------------------------------------------------------

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
