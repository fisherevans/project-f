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

var DefaultUnlockableSkills = map[SkillId]UnlockableSkill{
	Skill_Strike.Id: {
		Cost: 5,
	},
	Skill_Brace.Id: {
		Prerequisites: []SkillId{Skill_Strike.Id},
		Cost:          5,
	},
}

var Primortal_Dummy = Primortal{
	Type:        "dummy",
	Name:        "Dummy",
	Description: "A test robot to hit for fun.",
	BaseSync:    0,
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"onehit": {
			AdditionalSync: 1,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_DoNothing5.Id: 1,
					},
				},
			},
		},
		"training.1": {
			AdditionalSync: 40,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					InitialOrderedSkills: []SkillId{
						Skill_DoNothing5.Id,
						Skill_DoNothing5.Id,
						Skill_DoNothing5.Id,
						Skill_Dummy_Defend.Id,
						Skill_DoNothing5.Id,
					},
					WeightedSkills: map[SkillId]int{
						Skill_DoNothing5.Id:   3,
						Skill_Dummy_Defend.Id: 1,
					},
				},
			},
		},
		"default": {
			AdditionalSync: 40,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_Brace.Id: 40,
						Skill_Guard.Id: 20,
						Skill_Jab.Id:   10,
					},
				},
			},
		},
	},
}.register()

// kinetic
var Primortal_Pumbl = Primortal{
	Type:         "pumbl",
	Name:         "Pumbl",
	XenoLogIndex: 3,
	Description:  "A compact, rock-armored creature that can curl into an impenetrable sphere, using its dense body for both offense and defense.",
	BaseSync:     30,
	UnlockableSkills: map[SkillId]UnlockableSkill{
		Skill_ShoulderRoll.Id: {
			Cost: 2,
		},
		Skill_CurlUp.Id: {
			Prerequisites: []SkillId{Skill_ShoulderRoll.Id},
			Cost:          5,
		},
		Skill_Strike.Id: {
			Prerequisites: []SkillId{Skill_ShoulderRoll.Id},
			Cost:          5,
		},
		Skill_Brace.Id: {
			Prerequisites: []SkillId{Skill_Strike.Id},
			Cost:          5,
		},
		Skill_Searline.Id: {
			Prerequisites: []SkillId{Skill_Strike.Id},
			Cost:          7,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         10,
			AdditionalSyncVariance: 5,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_ShoulderRoll.Id: 40,
						Skill_CurlUp.Id:       40,
						Skill_Guard.Id:        20,
					},
				},
			},
		},
	},
}.register()

// thermal
var Primortal_Scintail = Primortal{
	Type:         "scintail",
	Name:         "Scintail",
	XenoLogIndex: 4,
	Description:  "A nimble reptilian creature with heat-radiating dorsal fins that pulse with an inner fire, capable of unleashing searing thermal attacks.",
	BaseSync:     25,
	UnlockableSkills: map[SkillId]UnlockableSkill{
		Skill_Cinder.Id: {
			Cost: 3,
		},
		Skill_Searline.Id: {
			Prerequisites: []SkillId{Skill_Cinder.Id},
			Cost:          7,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         5,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_Cinder.Id:   20,
						Skill_Searline.Id: 40,
						Skill_Brace.Id:    10,
						Skill_Strike.Id:   20,
					},
				},
			},
		},
	},
}.register()

// voltaic
var Primortal_Volteel = Primortal{
	Type:         "volteel",
	Name:         "Volteel",
	XenoLogIndex: 8,
	Description:  "An amphibious eel-like creature that generates and stores bioelectricity, using its stubby limbs to maneuver on land and deliver shocking attacks.",
	BaseSync:     20,
	UnlockableSkills: map[SkillId]UnlockableSkill{
		Skill_ArcDart.Id: {
			Cost: 2,
		},
		Skill_ZapWrap.Id: {
			Prerequisites: []SkillId{Skill_ArcDart.Id},
			Cost:          4,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         10,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_ArcDart.Id: 20,
						Skill_ZapWrap.Id: 10,
						Skill_Jab.Id:     40,
					},
				},
			},
		},
	},
}.register()

// voltaic - arcmander

// corrosive
var Primortal_Toxmidge = Primortal{
	Type:         "toxmidge",
	Name:         "Toxmidge",
	XenoLogIndex: 9,
	Description:  "A menacing aerial predator that secretes highly corrosive enzymes, capable of dissolving even the toughest materials with its acidic spray.",
	BaseSync:     25,
	UnlockableSkills: map[SkillId]UnlockableSkill{
		Skill_AcidSting.Id: {
			Cost: 3,
		},
		Skill_Molt.Id: {
			Prerequisites: []SkillId{Skill_AcidSting.Id},
			Cost:          7,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         5,
			AdditionalSyncVariance: 5,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_AcidSting.Id: 20,
						Skill_Molt.Id:      10,
						Skill_Guard.Id:     20,
						Skill_Jab.Id:       20,
					},
				},
			},
		},
		"training.2": {
			AdditionalSync: 5,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					InitialOrderedSkills: []SkillId{
						Skill_Guard.Id,
						Skill_Guard.Id,
						Skill_AcidSting.Id,
					},
					WeightedSkills: map[SkillId]int{
						Skill_AcidSting.Id: 1,
						Skill_Guard.Id:     2,
					},
				},
			},
		},
	},
}.register()

// mutagenic
var Primortal_Myceli = Primortal{
	Type:         "myceli",
	Name:         "Myceli",
	XenoLogIndex: 19,
	Description:  "A mysterious fungal entity with a network of mycelial tendrils that can rapidly regenerate and spread spores with various effects.",
	BaseSync:     40,
	UnlockableSkills: map[SkillId]UnlockableSkill{
		Skill_PhotoSurge.Id: {
			Cost: 4,
		},
		Skill_MendSpores.Id: {
			Prerequisites: []SkillId{Skill_PhotoSurge.Id},
			Cost:          5,
		},
	},
	CombatArchetypes: map[string]PrimortalCombatArchetype{
		"default": {
			AdditionalSync:         0,
			AdditionalSyncVariance: 10,
			SkillPool: CombatSkillPool{
				Random: &CombatSkillPoolRandom{
					WeightedSkills: map[SkillId]int{
						Skill_PhotoSurge.Id: 40,
						Skill_MendSpores.Id: 10,
						Skill_Jab.Id:        20,
						Skill_Brace.Id:      20,
					},
				},
			},
		},
	},
}.register()
