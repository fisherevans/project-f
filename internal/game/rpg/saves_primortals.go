package rpg

import (
	"time"
)

type PrimortalVisibility string

const (
	PrimortalVisibilityUnknown  PrimortalVisibility = ""
	PrimortalVisibilitySeen     PrimortalVisibility = "seen"
	PrimortalVisibilityDefeated PrimortalVisibility = "defeated"
)

type PrimortalProgress struct {
	Visibility     PrimortalVisibility `yaml:"visibility"`
	ResearchPoints int                 `yaml:"research_points"`
	LastSeen       time.Time           `yaml:"last_seen"`
}
