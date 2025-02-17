package rpg

import (
	"gopkg.in/yaml.v3"
	"strings"
)

var Skills = map[SkillId]Skill{}

type SkillId string

func (id SkillId) Get() Skill {
	if skill, exists := Skills[id]; exists {
		return skill
	}
	panic("unknown skill ID: " + id)
}

type SkillType string

const (
	SkillTypeKinetic  SkillType = "kinetic"
	SkillTypeVoltaic  SkillType = "voltaic"
	SkillTypeThermal  SkillType = "thermal"
	SkillTypeSonic    SkillType = "sonic"
	SkillTypeMagnetic SkillType = "magnetic"
	SkillTypeAcidic   SkillType = "acidic"
	SkillTypeGamma    SkillType = "gamma"
	SkillTypeAbyssal  SkillType = "abyssal"
)

type DamageMedium string

const (
	DamageMediumPhysical DamageMedium = "physical"
	DamageMediumAether   DamageMedium = "aether"
)

type Skill struct {
	Id          SkillId
	Name        string
	Description string
	Type        SkillType
	Ticks       []SkillTick
}

func (s Skill) Duration() int {
	return len(s.Ticks) - 1
}

// validate checks that the skill is correctly defined, identifying all the problems with the skull definition and then
// panicing with the full list of errors and the skill definition.
func (s Skill) validate() {
	errors := []string{}
	if s.Id == "" {
		errors = append(errors, "missing ID")
	}
	if s.Name == "" {
		errors = append(errors, "missing name")
	}
	if s.Description == "" {
		errors = append(errors, "missing description")
	}
	if s.Type == "" {
		errors = append(errors, "missing type")
	}
	if len(s.Ticks) == 0 {
		errors = append(errors, "missing ticks")
	}
	for _, tick := range s.Ticks {
		errors = append(errors, tick.validate()...)
	}
	if len(errors) > 0 {
		panic("invalid skill definition: " + strings.Join(errors, ", ") + "\n" + s.String())
	}
}

func (s Skill) register() Skill {
	s.validate()
	if _, exists := Skills[s.Id]; exists {
		panic("duplicate skill ID: " + s.Id)
	}
	Skills[s.Id] = s
	return s
}

func (s Skill) String() string {
	y, _ := yaml.Marshal(s)
	return string(y)
}
