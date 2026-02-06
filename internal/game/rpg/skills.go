package rpg

import (
	"strings"

	"gopkg.in/yaml.v3"

	"fisherevans.com/project/f/internal/game/input"
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

func (s *SkillSet) DirectionalSkill(dir input.Direction) *SkillId {
	switch dir {
	case input.Up:
		return &s.Skill1
	case input.Right:
		return &s.Skill2
	case input.Down:
		return &s.Skill3
	case input.Left:
		return &s.Skill4
	default:
		panic("invalid direction: " + dir.String())
	}
}

func (s *SkillSet) DirectionOfSkill(skill SkillId) (input.Direction, bool) {
	switch skill {
	case s.Skill1:
		return input.Up, true
	case s.Skill2:
		return input.Right, true
	case s.Skill3:
		return input.Down, true
	case s.Skill4:
		return input.Left, true
	default:
		return input.NotPressed, false
	}
}

func (s *SkillSet) IsEmpty() bool {
	return s.Skill1 == "" && s.Skill2 == "" && s.Skill3 == "" && s.Skill4 == ""
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
				errors = append(errors, "skill stance duration must be greater than 1 add")
			}
			stanceDuration = 0
			lastStance = tick.StanceType
		}
		stanceDuration++
		lastStance = tick.StanceType
	}
	if stanceDuration <= 1 && lastStance != TickStanceNone {
		errors = append(errors, "skill stance duration must be greater than 1 add")
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
	Description: "Raise defenses after a rest",
	Ticks: skillTicks().
		add(stanceTick(TickStanceDefending).repeat(3)...).
		add(tick().repeat(3)...),
}.register()

var Skill_Guard = Skill{
	Id:          "guard",
	Name:        "Guard",
	Description: "Defend yourself for a short time",
	Ticks: skillTicks().
		add(tick()).
		add(stanceTick(TickStanceDefending).repeat(3)...).
		add(tick()),
}.register()

// Attacking

var Skill_Tackle = Skill{
	Id:          "tackle",
	Name:        "Tackle",
	Description: "A crude physical attack",
	Ticks: skillTicks().
		add(tick().repeat(2)...).
		add(tick().damageAmountVaried(5, 2).animate(newAnimation().
			sourceTransformations(pounce()).
			targetTransformations(recoil()))).
		add(tick().repeat(1)...),
}.register()

var Skill_Jab = Skill{
	Id:          "jab",
	Name:        "Jab",
	Description: "Quick strike that deals light damage",
	Ticks: skillTicks().
		add(tick().damageAmountVaried(5, 2).animate(newAnimation().
			sourceTransformations(pounce()).
			targetTransformations(recoil()))).
		add(tick().repeat(2)...),
}.register()

var Skill_Strike = Skill{
	Id:          "strike",
	Name:        "Strike",
	Description: "A heavy blow that leaves you exposed",
	Ticks: skillTicks().
		add(stanceTick(TickStanceExposed).repeat(4)...).
		add(tick().damageAmount(15).animate(newAnimation().
			sourceTransformations(pounce()).
			targetTransformations(recoil()))),
}.register()

// ==============================================================
// ===== KINETIC SKILLS
// ==============================================================

// skill name ideas: Riposte

var Skill_ShoulderRoll = Skill{
	Id:          "shoulder_roll",
	Name:        "Shoulder Roll",
	Description: "Roll over your enemy repeatedly",
	Ticks: skillTicks().
		add(tick().repeat(4)...).
		add(tick().damageAmount(10).animate(newAnimation().
			sourceTransformations(pounce()).
			targetTransformations(recoil()))),
}.register()

var Skill_CurlUp = Skill{
	Id:          "curl_up",
	Name:        "Curl Up",
	Description: "Fortify your defences",
	Ticks: skillTicks().
		add(tick().statusSelf(StatusWarded, 5).animate(newAnimation().
			sourceTransformations(wiggle()))).
		add(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

// ==============================================================
// ===== THERMAL SKILLS
// ==============================================================

// skill name ideas: ignite

var Skill_Cinder = Skill{
	Id:          "cinder",
	Name:        "Cinder",
	Description: "Ignite your foe with a burning ember",
	Ticks: skillTicks().
		add(tick().repeat(3)...).
		add(tick().statusOpponent(StatusBurning, 5).animate(newAnimation().
			sourceTransformations(wiggle()))).
		add(tick()),
}.register()

var Skill_Searline = Skill{
	Id:          "searline",
	Name:        "Searline",
	Description: "Channel a searing beam that burns intensely",
	Ticks: skillTicks().
		add(stanceTick(TickStanceVulnerable).repeat(3)...).
		add(tick().damageAmount(3).status(&SkillTickStatus{
			Status:                StatusBurning,
			Stacks:                2,
			RequireExistingStacks: SkillTickStatusRequireExistingStacks,
		}, false).animate(newAnimation().
			sourceTransformations(hop(1))).repeat(3)...),
}.register()

// ==============================================================
// ===== VOLTAIC SKILLS
// ==============================================================

var Skill_ArcDart = Skill{
	Id:          "arc_dart",
	Name:        "Arc Dart",
	Description: "Launch an electrified dart that ionizes",
	Ticks: skillTicks().
		add(stanceTick(TickStanceExposed).repeat(3)...).
		add(tick()).
		add(tick().statusOpponent(StatusIonized, 6).animate(newAnimation().
			sourceTransformations(hop(1)))),
}.register()

var Skill_ZapWrap = Skill{
	Id:          "zap_wrap",
	Name:        "Zap Wrap",
	Description: "Ionize your foe as you wrap around them",
	Ticks: skillTicks().
		add().
		add(stanceTick(TickStanceDefending).
			damageAmountVaried(2, 1).
			animate(newAnimation().sourceTransformations(hop(1))).
			repeat(3)...).
		add(stanceTick(TickStanceDefending).statusOpponent(StatusIonized, 5)).
		add(),
}.register()

// ==============================================================
// ===== CORROSIVE SKILLS
// ==============================================================

// skill name ideas: bile surge

var Skill_Molt = Skill{
	Id:          "molt",
	Name:        "Molt",
	Description: "Shed your exoskeleton to remove debuffs",
	Ticks: skillTicks().
		add(stanceTick(TickStanceExposed).repeat(4)...).
		add(), // TODO remove statuses
}.register()

var Skill_AcidSting = Skill{
	Id:          "acid_string",
	Name:        "Acid String",
	Description: "Spray a stream of corrosive acid at your foe",
	Ticks: skillTicks().
		add(tick().damageAmount(2).statusOpponent(StatusPoisoned, 4)).
		add(tick().repeat(3)...),
}.register()

// ==============================================================
// ===== GROWTH SKILLS
// ==============================================================

var Skill_MendSpores = Skill{
	Id:          "mend_spores",
	Name:        "Mend Spores",
	Description: "Release healing spores that restore health",
	Ticks: skillTicks().
		add(tick().statusSelf(StatusMending, 5).animate(newAnimation().
			sourceTransformations(wiggle()))).
		add(stanceTick(TickStanceExposed).repeat(4)...).
		add(tick().statusSelf(StatusMending, 5).animate(newAnimation().
			sourceTransformations(wiggle()))),
}.register()

var Skill_PhotoSurge = Skill{
	Id:          "photo_surge",
	Name:        "Photo Surge",
	Description: "Deal more damage if your mending",
	Ticks: skillTicks().
		add(stanceTick(TickStanceExposed).repeat(2)...).
		add(tick().damage(&SkillTickDamage{
			Amount: 5,
			ScaledBy: SkillTickDamageScalers{
				SourceStatus: map[StatusType]map[StatusLevel]float64{
					StatusMending: {
						StatusLevel1: 1.5,
						StatusLevel2: 2,
						StatusLevel3: 3,
					},
				},
			},
		}).animate(newAnimation().
			sourceTransformations(hop(2)))),
}.register()

// ========================================

// Dummy moves

var Skill_DoNothing5 = Skill{
	Id:          "do_nothing_5",
	Name:        "Do Nothing 5",
	Description: "not used",
	Ticks:       skillTicks().add(tick().repeat(5)...),
}.register()

var Skill_Dummy_Defend = Skill{
	Id:          "dummy_defend",
	Name:        "Dummy Defend",
	Description: "not used",
	Ticks: skillTicks().
		add(tick().repeat(2)...).
		add(stanceTick(TickStanceDefending).repeat(3)...),
}.register()

var Skill_Dummy_Attack = Skill{
	Id:          "dummy_attack",
	Name:        "Dummy Attack",
	Description: "not used",
	Ticks: skillTicks().
		add(tick().repeat(2)...).
		add(tick().damageAmountVaried(2, 0).animate(newAnimation().
			sourceTransformations(pounce()).
			targetTransformations(recoil()))),
}.register()
