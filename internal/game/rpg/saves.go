package rpg

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"

	"fisherevans.com/project/f/internal/util/delta"
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
	if _, ok := g.ControlledUnlockedSkills[Skill_Jab.Id]; !ok {
		g.ControlledUnlockedSkills[Skill_Jab.Id] = struct{}{}
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

func (g *GameSave) Save() error {
	if g.SaveId == "" {
		return fmt.Errorf("GameSave has empty saveId")
	}

	filename := fmt.Sprintf("%s.yaml", g.SaveId)
	path := filepath.Join(gameSaveDirectory, filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(gameSaveDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	// Load the existing save (if any) to compute a delta
	var oldSave *GameSave
	if existingData, err := os.ReadFile(path); err == nil {
		var prev GameSave
		if err := yaml.Unmarshal(existingData, &prev); err == nil {
			oldSave = &prev
		}
	}

	data, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("failed to marshal GameSave: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write GameSave: %w", err)
	}

	// Compute and log a human-readable delta
	d := delta.HumanDiff(oldSave, g)
	if strings.TrimSpace(d) == "" {
		log.Info().Msgf("Saved game to %s (no changes)", path)
	} else {
		log.Info().Msgf("Saved game to %s. Changes:\n%s", path, d)
	}

	return nil
}

func (g *GameSave) GrantResearchPoints(p PrimortalType, points int) {
	if _, ok := g.Primortals[p]; !ok {
		g.Primortals[p] = &PrimortalProgress{}
	}
	g.Primortals[p].Visibility = PrimortalVisibilityDefeated
	g.Primortals[p].ResearchPoints += points
	g.Primortals[p].LastSeen = time.Now()

}

func (g *GameSave) GrantExperience(points int) {
	if g.Animech == nil {
		g.Animech = &Animech{}
		g.Animech.FillDefaults()
	}
	g.Animech.AnimechExperience += points
}

func LoadGameSaves() (map[string]*GameSave, error) {
	saves := make(map[string]*GameSave)

	entries, err := os.ReadDir(gameSaveDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist yet, return empty map
			return saves, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}

		base := strings.TrimSuffix(e.Name(), ".yaml")
		path := filepath.Join(gameSaveDirectory, e.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			log.Warn().Msgf("Failed to read file %s: %v", path, err)
			continue
		}

		var gs GameSave
		if err := yaml.Unmarshal(data, &gs); err != nil {
			log.Warn().Msgf("Failed to unmarshal %s: %v", path, err)
			continue
		}

		// initialize config with defaults if needed
		gs.FillDefaults()

		// todo validate loaded saves (i.e. skills in loadouts are unlocked and valid ids)

		// Ignore files whose name doesn't match the saveId
		if gs.SaveId != base {
			log.Warn().Msgf("Mismatched saveId in %s (expected %s, got %s)", path, base, gs.SaveId)
			continue
		}

		saves[gs.SaveId] = &gs
	}

	return saves, nil
}
