package types

type MoveState int

const (
	MoveStateIdle MoveState = iota
	MoveStateWalking
	MoveStateRunning
	MoveStateDashing
)

const (
	MetadataKeyMode      = "mode"
	MetadataKeyIsTalking = "isTalking"
)

type ModeBaseRenderConfig struct {
	Mode       *string                         `yaml:"mode"`
	Animations map[string][]AnimationReference `yaml:"animations"`
	Lights     map[string][]LightConfig        `yaml:"lights"`
}

type LightConfig struct {
	Color    string  `yaml:"color"`
	Size     float64 `yaml:"size"`
	Modifier *string `yaml:"modifier"`
}

type AnimationReference struct {
	Name      string  `yaml:"name"`
	ColorMask *string `yaml:"colorMask"`
}

type PresenceConfig struct {
	IsInteractable *bool `yaml:"is_interactable"`
	BlockIngress   *bool `yaml:"block_ingress"`
}
