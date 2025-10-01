package rpg

import (
	"strings"

	"gopkg.in/yaml.v3"
)

var Skills = map[SkillId]Skill{}

type SkillId string

const UnsetSkillId SkillId = ""

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

type Loadout struct {
	Name     string
	SkillSet SkillSet `yaml:"skill_set"`
}

type SkillSet struct {
	Skill1 SkillId `yaml:"1"`
	Skill2 SkillId `yaml:"2"`
	Skill3 SkillId `yaml:"3"`
	Skill4 SkillId `yaml:"4"`
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

// ==============================================================
// ===== BASIC SKILLS
// ==============================================================

// Defending

var Skill_Brace = Skill{
	Id:          "brace",
	Name:        "Brace",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().repeat(2)...).
		tick(stanceTick(TickStanceDefending).repeat(2)...),
}.register()

var Skill_Guard = Skill{
	Id:          "guard",
	Name:        "Guard",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick()).
		tick(stanceTick(TickStanceDefending).repeat(3)...).
		tick(tick()),
}.register()

// Attacking

var Skill_Jab = Skill{
	Id:          "jab",
	Name:        "Jab",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().damage(5, 2)).
		tick(tick().damage(2, 1)),
}.register()

var Skill_Strike = Skill{
	Id:          "strike",
	Name:        "Strike",
	Description: "todo",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(4)...).
		tick(tick().damage(15, 0)),
}.register()

// ==============================================================
// ===== KINETIC SKILLS
// ==============================================================

// skill name ideas: Riposte

var Skill_ShoulderRoll = Skill{
	Id:          "shoulder_roll",
	Name:        "Shoulder Roll",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick()).
		tick(tick().damage(10, 0)),
}.register()

var Skill_CurlUp = Skill{
	Id:          "curl_up",
	Name:        "Curl Up",
	Description: "todo",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceDefending).repeat(4)...), // todo add defence status
}.register()

// ==============================================================
// ===== THERMAL SKILLS
// ==============================================================

// skill name ideas: ignite

var Skill_Cinder = Skill{
	Id:          "cinder",
	Name:        "Cinder",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().status(StatusBurning, 2)).
		tick(tick()),
}.register()

var Skill_Searline = Skill{
	Id:          "searline",
	Name:        "Searline",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().damage(3, 0).repeat(3)...). // todo increase burn status condition if there
		tick(stanceTick(TickStanceVulnerable).repeat(3)...),
}.register()

// ==============================================================
// ===== VOLTAIC SKILLS
// ==============================================================

var Skill_ArcDart = Skill{
	Id:          "arc_dart",
	Name:        "Arc Dart",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().status(StatusBurning, 5)). // todo apply ionize
		tick(stanceTick(TickStanceExposed).repeat(3)...),
}.register()

var Skill_ZapWrap = Skill{
	Id:          "zap_wrap",
	Name:        "Zap Wrap",
	Description: "todo",
	Ticks: skillTicks().
		tick().
		tick(stanceTick(TickStanceDefending).damage(3, 3).repeat(4)...).
		tick(),
}.register()

// ==============================================================
// ===== CORROSIVE SKILLS
// ==============================================================

// skill name ideas: bile surge

var Skill_Molt = Skill{
	Id:          "molt",
	Name:        "Molt",
	Description: "todo",
	Ticks: skillTicks().
		tick(stanceTick(TickStanceExposed).repeat(3)...).
		tick(), // TODO remove statuses
}.register()

var Skill_AcidSting = Skill{
	Id:          "acid_string",
	Name:        "Acid String",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().damage(2, 0).status(StatusPoisoned, 3)).
		tick(),
}.register()

// ==============================================================
// ===== GROWTH SKILLS
// ==============================================================

var Skill_MendSpores = Skill{
	Id:          "mend_spores",
	Name:        "Mend Spores",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().repeat(6)...), // TODO regen
}.register()

var Skill_PhotoSurge = Skill{
	Id:          "photo_surge",
	Name:        "Photo Surge",
	Description: "todo",
	Ticks: skillTicks().
		tick(tick().damage(5, 0)). // TODO scale up over time
		tick(stanceTick(TickStanceExposed).repeat(2)...),
}.register()
