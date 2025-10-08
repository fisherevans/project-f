package rpg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v3"
)

var gameSaveDirectory = "game_data/saves"

type GameSave struct {
	SaveId        string                               `yaml:"save_id"`
	CharacterName string                               `yaml:"character_name"`
	Animech       *Animech                             `yaml:"animech"`
	Loadouts      []*Loadout                           `yaml:"loadouts"`
	Inventory     *Inventory                           `yaml:"inventory"`
	Primortals    map[PrimortalType]*PrimortalProgress `yaml:"primortals"`

	unlockedSkillsList []SkillId            `yaml:"unlocked_skills"`
	UnlockedSkills     map[SkillId]struct{} `yaml:"-"`
}

func (g *GameSave) initializeAfterLoad() {
	g.UnlockedSkills = map[SkillId]struct{}{}
	for _, s := range g.unlockedSkillsList {
		g.UnlockedSkills[s] = struct{}{}
	}
}

func (g *GameSave) normalizeBeforeSave() {
	g.unlockedSkillsList = []SkillId{}
	for s := range g.UnlockedSkills {
		g.unlockedSkillsList = append(g.unlockedSkillsList, s)
	}
}

func (g *GameSave) IsSkillUnlocked(skill SkillId) bool {
	_, isUnlocked := g.UnlockedSkills[skill]
	return isUnlocked
}

func (g *GameSave) Save() error {
	if g.SaveId == "" {
		return fmt.Errorf("GameSave has empty saveId")
	}

	filename := fmt.Sprintf("%s.yaml", g.SaveId)
	path := filepath.Join(gameSaveDirectory, filename)

	g.normalizeBeforeSave()
	data, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("failed to marshal GameSave: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write GameSave: %w", err)
	}

	return nil
}

func LoadGameSaves() (map[string]*GameSave, error) {
	saves := make(map[string]*GameSave)

	entries, err := os.ReadDir(gameSaveDirectory)
	if err != nil {
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
		gs.initializeAfterLoad()

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
