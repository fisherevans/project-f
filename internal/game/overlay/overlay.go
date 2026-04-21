package overlay

import (
	"github.com/gopxl/pixel/v2/backends/opengl"
	"github.com/gopxl/pixel/v2/ext/imdraw"
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

	panelOpen      bool
	scalePopoverOpen bool // vertical scale list above the scale button
	settingsDirty  bool   // any setting changed; save on next pointer release

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
	}

	if o.idleTimer < idleHideSeconds {
		o.alpha = clamp01(o.alpha + dt/fadeDuration)
	} else {
		o.alpha = clamp01(o.alpha - dt/fadeDuration)
	}
}

func (o *Overlay) Render(win *opengl.Window, dt float64) {
	if o.alpha <= 0 {
		return
	}
	if o.panelOpen {
		o.renderSettingsPanel(win)
	}
	o.renderCornerChrome(win)

	// Persist settings on pointer release. Captures click-releases from sliders,
	// toggles, and segmented selects without spamming the save during drags.
	if o.settingsDirty && o.pointer.JustUp {
		if o.hooks.SaveSettings != nil {
			o.hooks.SaveSettings()
		}
		o.settingsDirty = false
	}
}

func (o *Overlay) markDirty() { o.settingsDirty = true }

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
