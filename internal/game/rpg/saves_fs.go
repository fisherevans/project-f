//go:build !js

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

func (g *GameSave) Save() error {
	if g.SaveId == "" {
		return fmt.Errorf("GameSave has empty saveId")
	}

	filename := fmt.Sprintf("%s.yaml", g.SaveId)
	path := filepath.Join(gameSaveDirectory, filename)

	if err := os.MkdirAll(gameSaveDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

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
		if os.IsNotExist(err) {
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

		gs.FillDefaults()

		if gs.SaveId != base {
			log.Warn().Msgf("Mismatched saveId in %s (expected %s, got %s)", path, base, gs.SaveId)
			continue
		}

		saves[gs.SaveId] = &gs
	}

	return saves, nil
}
