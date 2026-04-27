package schema

type OverlayRect struct {
	X int `yaml:"x" json:"x"`
	Y int `yaml:"y" json:"y"`
	W int `yaml:"w" json:"w"`
	H int `yaml:"h" json:"h"`
}

type OverlayTargetMessage struct {
	Text      string `yaml:"text" json:"text"`
	Placement string `yaml:"placement" json:"placement"`
	Wrap      int    `yaml:"wrap,omitempty" json:"wrap,omitempty"`
}

type OverlayTargetBadge struct {
	Placement string `yaml:"placement,omitempty" json:"placement,omitempty"`
	Label     string `yaml:"label,omitempty" json:"label,omitempty"`
}

type OverlayTarget struct {
	Rect      string                `yaml:"rect,omitempty" json:"rect,omitempty"`
	Region    *OverlayRect          `yaml:"region,omitempty" json:"region,omitempty"`
	Message   *OverlayTargetMessage `yaml:"message,omitempty" json:"message,omitempty"`
	Badge     *OverlayTargetBadge   `yaml:"badge,omitempty" json:"badge,omitempty"`
	NoPadding bool                  `yaml:"no_padding,omitempty" json:"noPadding,omitempty"`
}

type OverlayFlow struct {
	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Targets     []OverlayTarget `yaml:"targets" json:"targets"`
}
