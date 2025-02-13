package rpg

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var gameSaveDirectory = "game_data/saves"

type GameSave struct {
	SaveId             string                              `yaml:"save_id"`
	CharacterName      string                              `yaml:"character_name"`
	Animech            *Animech                            `yaml:"animech"`
	CapturedPrimortals map[PrimortalType]CapturedPrimortal `yaml:"captured_primortals"`
	Inventory          *Inventory                          `yaml:"inventory"`
}

func (g *GameSave) Save() error {
	if g.SaveId == "" {
		return fmt.Errorf("GameSave has empty saveId")
	}

	filename := fmt.Sprintf("%s.yaml", g.SaveId)
	path := filepath.Join(gameSaveDirectory, filename)

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

		// Ignore files whose name doesn't match the saveId
		if gs.SaveId != base {
			log.Warn().Msgf("Mismatched saveId in %s (expected %s, got %s)", path, base, gs.SaveId)
			continue
		}

		saves[gs.SaveId] = &gs
	}

	return saves, nil
}
