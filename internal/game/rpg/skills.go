package rpg

var Skills = map[SkillId]Skill{}

type SkillId string

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
	Id           SkillId
	Name         string
	Description  string
	Type         SkillType
	DamageMedium DamageMedium
}

func (s Skill) register() Skill {
	if _, exists := Skills[s.Id]; exists {
		panic("duplicate skill ID: " + s.Id)
	}
	Skills[s.Id] = s
	return s
}
