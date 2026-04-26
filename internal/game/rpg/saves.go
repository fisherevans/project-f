package rpg

import (
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var gameSaveDirectory = "game_data/saves"

type GameSave struct {
	SaveId        string                               `yaml:"save_id"`
	CharacterName string                               `yaml:"character_name"`
	Animech       *Animech                             `yaml:"animech"`
	Loadouts      []*Loadout                           `yaml:"loadouts"`
	Inventory     *Inventory                           `yaml:"inventory"`
	Primortals    map[PrimortalType]*PrimortalProgress `yaml:"primortals"`

	ControlledUnlockedSkills map[SkillId]struct{} `yaml:"unlocked_skills"`

	Globals *defaultGlobals `yaml:"globals"`

	SystemSettings *SystemSettings `yaml:"system_settings"`
}

func (g *GameSave) FillDefaults() {
	if g.SystemSettings == nil {
		g.SystemSettings = &SystemSettings{}
	}
	g.SystemSettings.FillDefaults()
	if g.Globals == nil {
		g.Globals = &defaultGlobals{
			values: make(map[string]any),
		}
	}
	if g.ControlledUnlockedSkills == nil {
		g.ControlledUnlockedSkills = make(map[SkillId]struct{})
	}
	if _, ok := g.ControlledUnlockedSkills["jab"]; !ok {
		g.ControlledUnlockedSkills["jab"] = struct{}{}
	}
	if g.Animech == nil {
		g.Animech = &Animech{}
	}
	g.Animech.FillDefaults()
	if g.Primortals == nil {
		g.Primortals = make(map[PrimortalType]*PrimortalProgress)
	}
}

func (g *GameSave) IsSkillUnlocked(skill SkillId) bool {
	_, isUnlocked := g.ControlledUnlockedSkills[skill]
	return isUnlocked
}

func (g *GameSave) UnlockSkill(skill SkillId) {
	if g.ControlledUnlockedSkills == nil {
		g.ControlledUnlockedSkills = map[SkillId]struct{}{}
	}
	g.ControlledUnlockedSkills[skill] = struct{}{}
}

func (g *GameSave) UnlockedSkillsSorted() []SkillId {
	ids := make([]SkillId, 0, len(g.ControlledUnlockedSkills))
	for id := range g.ControlledUnlockedSkills {
		ids = append(ids, id)
	}
	slices.SortFunc(ids, func(a, b SkillId) int {
		return strings.Compare(a.Get().Name, b.Get().Name)
	})
	return ids
}

func (g *GameSave) RemoveUnlockedSkill(skill SkillId) {
	if g.ControlledUnlockedSkills == nil {
		g.ControlledUnlockedSkills = map[SkillId]struct{}{}
	}
	delete(g.ControlledUnlockedSkills, skill)
}

func (g *GameSave) GrantResearchPoints(p PrimortalType, points int) {
	if _, ok := Primortals[p]; !ok {
		log.Warn().Msgf("unknown primortal type, cannot grant RP: %v", p)
		return
	}
	if _, ok := g.Primortals[p]; !ok {
		g.Primortals[p] = &PrimortalProgress{}
	}
	g.Primortals[p].Visibility = PrimortalVisibilityDefeated
	g.Primortals[p].PendingResearchPoints += points
	g.Primortals[p].LastSeen = time.Now()

}

func (g *GameSave) GrantExperience(points int) {
	if g.Animech == nil {
		g.Animech = &Animech{}
		g.Animech.FillDefaults()
	}
	g.Animech.PendingExperience += points
}

