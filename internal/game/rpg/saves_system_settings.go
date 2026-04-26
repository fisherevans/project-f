package rpg

type SystemSettings struct {
	Lighting   *SystemSettingsLighting   `yaml:"lighting"`
	Combat     *SystemSettingsCombat     `yaml:"combat"`
	RetroFrame *SystemSettingsRetroFrame `yaml:"retro_frame"`
	Debugging  *SystemSettingsDebugging  `yaml:"debugging"`
	Audio      *SystemSettingsAudio      `yaml:"audio"`
	Display    *SystemSettingsDisplay    `yaml:"display"`
}

func (s *SystemSettings) FillDefaults() {
	if s.Lighting == nil {
		s.Lighting = &SystemSettingsLighting{}
	}
	s.Lighting.FillDefaults()
	if s.Combat == nil {
		s.Combat = &SystemSettingsCombat{}
	}
	s.Combat.FillDefaults()
	if s.RetroFrame == nil {
		s.RetroFrame = &SystemSettingsRetroFrame{}
	}
	if s.Debugging == nil {
		s.Debugging = &SystemSettingsDebugging{}
	}
	s.Debugging.FillDefaults()
	s.RetroFrame.FillDefaults()
	if s.Audio == nil {
		s.Audio = &SystemSettingsAudio{}
	}
	s.Audio.FillDefaults()
	if s.Display == nil {
		s.Display = &SystemSettingsDisplay{}
	}
	s.Display.FillDefaults()
}

type SystemSettingsLighting struct {
	LightingComposition LightingComposition `yaml:"lighting_composition"`
	LightingMode        LightingMode        `yaml:"lighting_mode"`

	BloomMode BloomMode `yaml:"bloom_mode"`
}

func (s *SystemSettingsLighting) FillDefaults() {
	if s.LightingComposition == "" {
		s.LightingComposition = LightingCompositionFull
	}
	if s.LightingMode == "" {
		s.LightingMode = LightingModeBlended
	}
	if s.BloomMode == "" {
		s.BloomMode = BloomModeBlended
	}
}

type SystemSettingsCombat struct {
	CombatSpeed float64 `yaml:"combat_speed"`
}

func (s *SystemSettingsCombat) FillDefaults() {
	if s.CombatSpeed == 0 {
		s.CombatSpeed = 1.0
	}
}

type SystemSettingsRetroFrame struct {
	DisablePixelGrid    bool    `yaml:"disable_pixel_grid"`
	OverrideRetroOverlay bool   `yaml:"override_retro_overlay"` // false = derive from render scale
	ScanlineDarken    float64 `yaml:"scanline_darken"`
	GridDarkenX       float64 `yaml:"grid_darken_x"`
	GridDarkenY       float64 `yaml:"grid_darken_y"`
	SubpixelTint      float64 `yaml:"subpixel_tint"`
}

// EffectiveGridValues returns the grid darken and subpixel tint to pass to the
// shader. When OverrideRetroOverlay is false, values are derived from physScale so
// the grid stays visually consistent as the window resizes. When true, the
// stored values are used as-is.
func (s *SystemSettingsRetroFrame) EffectiveGridValues(physScale int) (darkenX, darkenY, subpixel float64) {
	if s.OverrideRetroOverlay {
		return s.GridDarkenX, s.GridDarkenY, s.SubpixelTint
	}
	t := clampF(float64(physScale-2)/4.0, 0, 1) // 0 at scale=2, 1 at scale=6
	darkenX = 0.08 + 0.12*t
	darkenY = 0.08 + 0.12*t
	subpixel = 0.048 + (0.0666-0.048)*t
	return
}

func (s *SystemSettingsRetroFrame) FillDefaults() {
	s.FillDefaultsForScale(0)
}

// FillDefaultsForScale fills zero fields with scale-aware defaults.
// physScale is the physical pixels per game pixel (e.g. 2 on desktop, 5 on
// iPhone 16 Pro). Pass 0 to use a mid-range default (physScale=3).
// Grid darken values scale up with physScale because at higher scales each
// grid line covers a smaller fraction of the screen, requiring stronger
// darkening to remain visible.
func (s *SystemSettingsRetroFrame) FillDefaultsForScale(physScale int) {
	if physScale <= 0 {
		physScale = 3
	}
	if s.ScanlineDarken == 0 {
		s.ScanlineDarken = 0.015
	}
	t := clampF(float64(physScale-2)/4.0, 0, 1) // 0 at scale=2, 1 at scale=6
	if s.GridDarkenX == 0 {
		s.GridDarkenX = 0.08 + 0.12*t
	}
	if s.GridDarkenY == 0 {
		s.GridDarkenY = 0.08 + 0.12*t
	}
	if s.SubpixelTint == 0 {
		s.SubpixelTint = 0.048 + (0.0666-0.048)*t
	}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

type LightingComposition string

const (
	LightingCompositionFull    LightingComposition = "full"
	LightingCompositionAmbient LightingComposition = "ambient"
)

var LightingCompositions = []LightingComposition{
	LightingCompositionFull,
	LightingCompositionAmbient,
}

type LightingMode string

const (
	LightingModeBlended LightingMode = "Blended"
	LightingModeOver    LightingMode = "Over"
	LightingModeOff     LightingMode = "Off"
)

var LightingModes = []LightingMode{
	LightingModeBlended,
	LightingModeOver,
	LightingModeOff,
}

type BloomMode string

const (
	BloomModeBlended       BloomMode = "Blended"
	BloomModeOverBlurred   BloomMode = "OverBlended"
	BloomModeOverThreshold BloomMode = "OverThreshold"
	BloomModeOff           BloomMode = "Off"
)

var BloomModes = []BloomMode{
	BloomModeBlended,
	BloomModeOverBlurred,
	BloomModeOverThreshold,
	BloomModeOff,
}

type SystemSettingsDebugging struct {
	ShowPathfindingDebugging bool    `yaml:"show_pathfinding_debugging"`
	GameTimeSpeed            float64 `yaml:"game_time_speed"`
}

func (s *SystemSettingsDebugging) FillDefaults() {
	if s.GameTimeSpeed == 0 {
		s.GameTimeSpeed = 1.0
	}
}

type SystemSettingsAudio struct {
	Muted        bool    `yaml:"muted"`
	MasterVolume float64 `yaml:"master_volume"` // linear 0..1
}

func (s *SystemSettingsAudio) FillDefaults() {
	if s.MasterVolume == 0 {
		s.MasterVolume = 1.0
	}
}

// ScaleModeAuto lets the runtime pick the largest integer scale that fits.
// ScaleModeFixed locks to FixedScale (letterboxed if window is larger).
const (
	ScaleModeAuto  = "auto"
	ScaleModeFixed = "fixed"
)

// VirtualGamepadAuto shows the on-screen pad the first time a touch event
// is seen. On and Off force it regardless.
const (
	VirtualGamepadAuto = "auto"
	VirtualGamepadOn   = "on"
	VirtualGamepadOff  = "off"
)

type SystemSettingsDisplay struct {
	ScaleMode             string  `yaml:"scale_mode"`              // "auto" | "fixed"
	FixedScale            int     `yaml:"fixed_scale"`
	Fullscreen            bool    `yaml:"fullscreen"`
	VirtualGamepad        string  `yaml:"virtual_gamepad"`         // "auto" | "on" | "off"
	VirtualGamepadOpacity float64 `yaml:"virtual_gamepad_opacity"` // 0.2 – 1.0
	UIScale               float64 `yaml:"ui_scale"`                // multiplier on auto-computed overlay scale
}

func (s *SystemSettingsDisplay) FillDefaults() {
	if s.ScaleMode == "" {
		s.ScaleMode = ScaleModeAuto
	}
	if s.VirtualGamepad == "" {
		s.VirtualGamepad = VirtualGamepadAuto
	}
	if s.VirtualGamepadOpacity <= 0 {
		s.VirtualGamepadOpacity = 0.85
	}
	if s.UIScale == 0 {
		s.UIScale = 1.0
	}
}
