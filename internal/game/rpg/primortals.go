package rpg

import "fmt"

var Primortals = map[PrimortalType]Primortal{}
var XenoLogEntries = map[int]PrimortalType{}
var MaxXenoLogEntryIndex = 0
var reservedUnlockableSkill = map[SkillId]bool{}

const SkillCostUnavailable = 0

type PrimortalType string

func (pt PrimortalType) Primortal() Primortal {
    if p, exists := Primortals[pt]; exists {
        return p
    }
    panic("unknown primortal type: " + pt)
}

type Primortal struct {
    XenoLogIndex     int
    Type             PrimortalType
    Name             string
    Description      string
    BaseSync         int
    UnlockableSkills map[SkillId]UnlockableSkill
    CombatArchetypes map[string]PrimortalCombatArchetype
}

type PrimortalCombatArchetype struct {
    AdditionalSync         int             `yaml:"additional_sync"`
    AdditionalSyncVariance int             `yaml:"additional_sync_variance"`
    SkillPool              CombatSkillPool `yaml:"skill_pool"`
}

type CombatSkillPool struct {
    Random *CombatSkillPoolRandom `yaml:"random"`
}

type CombatSkillPoolRandom struct {
    InitialOrderedSkills []SkillId       `yaml:"initial_ordered_skills"`
    WeightedSkills       map[SkillId]int `yaml:"weighted_skills"`
}

type UnlockableSkill struct {
    Prerequisites []SkillId
    Cost          int
}

func (p Primortal) register() Primortal {
    if _, exists := Primortals[p.Type]; exists {
        panic("duplicate primortal type: " + p.Type)
    }
    for skillId := range p.UnlockableSkills {
        if taken, _ := reservedUnlockableSkill[skillId]; taken {
            panic("duplicate unlockable skill: " + skillId)
        }
    }
    if p.XenoLogIndex < 0 || p.XenoLogIndex > 1000 {
        panic(fmt.Sprintf("invalid XenoLogIndex: %d", p.XenoLogIndex))
    }
    if p.XenoLogIndex != 0 {
        if _, exists := XenoLogEntries[p.XenoLogIndex]; exists {
            panic(fmt.Sprintf("duplicate XenoLogIndex: %d", p.XenoLogIndex))
        }
        XenoLogEntries[p.XenoLogIndex] = p.Type
        if p.XenoLogIndex > MaxXenoLogEntryIndex {
            MaxXenoLogEntryIndex = p.XenoLogIndex
        }
    }
    Primortals[p.Type] = p
    return p
}
