package rpg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
