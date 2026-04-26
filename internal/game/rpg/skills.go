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

func (s Skill) String() string {
    y, _ := yaml.Marshal(s)
    return string(y)
}
