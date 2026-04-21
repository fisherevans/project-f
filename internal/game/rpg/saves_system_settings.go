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
	DisablePixelGrid bool    `yaml:"disable_pixel_grid"`
	ScanlineDarken   float64 `yaml:"scanline_darken"`
	GridDarkenX      float64 `yaml:"grid_darken_x"`
	GridDarkenY      float64 `yaml:"grid_darken_y"`
	SubpixelTint     float64 `yaml:"subpixel_tint"`
}

func (s *SystemSettingsRetroFrame) FillDefaults() {
	if s.ScanlineDarken == 0 {
		s.ScanlineDarken = 0.015
	}
	if s.GridDarkenX == 0 {
		s.GridDarkenX = 0.05
	}
	if s.GridDarkenY == 0 {
		s.GridDarkenY = 0.05
	}
	if s.SubpixelTint == 0 {
		s.SubpixelTint = 0.05
	}
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
	ShowPathfindingDebugging bool `yaml:"show_pathfinding_debugging"`
}

func (s *SystemSettingsDebugging) FillDefaults() {
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
}
