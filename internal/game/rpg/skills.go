package rpg

import (
	"strings"

	"gopkg.in/yaml.v3"
)

var Skills = map[SkillId]Skill{}

type SkillId string

func (id SkillId) Get() Skill {
	if skill, exists := Skills[id]; exists {
		return skill
	}
	panic("unknown skill ID: " + id)
}

type Skill struct {
	Id          SkillId
	Name        string
	Description string
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
	if len(s.Ticks) == 0 {
		errors = append(errors, "missing ticks")
	}
	lastStance := TickStanceNone
	stanceDuration := 0
	for _, tick := range s.Ticks {
		errors = append(errors, tick.validate()...)
		if tick.StanceType != lastStance {
			if stanceDuration <= 1 && lastStance != TickStanceNone {
				errors = append(errors, "skill stance duration must be greater than 1 tick")
			}
			stanceDuration = 0
			lastStance = tick.StanceType
		}
		stanceDuration++
		lastStance = tick.StanceType
	}
	if stanceDuration <= 1 && lastStance != TickStanceNone {
		errors = append(errors, "skill stance duration must be greater than 1 tick")
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

var Skill_Tackle = Skill{
	Id:          "tackle",
	Name:        "Tackle",
	Description: "Tackle an enemy, dealing kinetic damage.",
	Ticks:       simpleDamageSkillTicks(3, 3),
}.register()

var Skill_Crush = Skill{
	Id:          "crush",
	Name:        "Crush",
	Description: "Slam down with immense force, dealing heavy kinetic damage.",
	Ticks: skillTicks().
		tick(damageTick().damage(6, 4)).
		tick(nothingTick().repeat(2)...),
}.register()

var Skill_Shunt = Skill{
	Id:          "shunt",
	Name:        "Shunt",
	Description: "Shunt the foe and enter a defencive stance.",
	Ticks: skillTicks().
		tick(damageTick().damage(6, 4)).
		tick(nothingTick().repeat(2)...).
		tick(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

var Skill_Block = Skill{
	Id:          "block",
	Name:        "Block",
	Description: "Raise your defences briefly",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

var Skill_Riposte = Skill{
	Id:          "riposte",
	Name:        "Riposte",
	Description: "Reflect damage after being exposed",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(2)...).
		tick(stanceTick(TickStanceReflecting).repeat(4)...),
}.register()

var Skill_DrawnBlow = Skill{
	Id:          "drawn_blow",
	Name:        "Drawn Blow",
	Description: "Make a huge hit after being exposed",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(4)...).
		tick(damageTick().damage(15, 4)),
}.register()
