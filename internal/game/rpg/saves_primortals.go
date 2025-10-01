package rpg

type PrimortalVisibility string

const (
	PrimortalVisibilityUnknown PrimortalVisibility = ""
	PrimortalVisibilitySeen    PrimortalVisibility = "seen"
)

type PrimortalProgress struct {
	Visibility     PrimortalVisibility `yaml:"visibility"`
	ResearchPoints int                 `yaml:"research_points"`
}
