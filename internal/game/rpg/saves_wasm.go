//go:build js && wasm

package rpg

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"fisherevans.com/project/f/internal/util/delta"
)

const localStorageSavePrefix = "save:"

func localStorage() js.Value {
	return js.Global().Get("localStorage")
}

func (g *GameSave) Save() (retErr error) {
	if g.SaveId == "" {
		return fmt.Errorf("GameSave has empty saveId")
	}

	key := localStorageSavePrefix + g.SaveId
	ls := localStorage()

	var oldSave *GameSave
	if existing := ls.Call("getItem", key); !existing.IsNull() && !existing.IsUndefined() {
		var prev GameSave
		if err := yaml.Unmarshal([]byte(existing.String()), &prev); err == nil {
			oldSave = &prev
		}
	}

	data, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("failed to marshal GameSave: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("localStorage.setItem threw: %v", r)
		}
	}()
	ls.Call("setItem", key, string(data))

	d := delta.HumanDiff(oldSave, g)
	if strings.TrimSpace(d) == "" {
		log.Info().Msgf("Saved game to %s (no changes)", key)
	} else {
		log.Info().Msgf("Saved game to %s. Changes:\n%s", key, d)
	}

	return nil
}

func LoadGameSaves() (map[string]*GameSave, error) {
	saves := make(map[string]*GameSave)

	ls := localStorage()
	if ls.IsNull() || ls.IsUndefined() {
		return saves, nil
	}

	length := ls.Get("length").Int()
	for i := 0; i < length; i++ {
		keyVal := ls.Call("key", i)
		if keyVal.IsNull() || keyVal.IsUndefined() {
			continue
		}
		key := keyVal.String()
		if !strings.HasPrefix(key, localStorageSavePrefix) {
			continue
		}
		base := strings.TrimPrefix(key, localStorageSavePrefix)

		item := ls.Call("getItem", key)
		if item.IsNull() || item.IsUndefined() {
			continue
		}

		var gs GameSave
		if err := yaml.Unmarshal([]byte(item.String()), &gs); err != nil {
			log.Warn().Msgf("Failed to unmarshal localStorage key %s: %v", key, err)
			continue
		}

		gs.FillDefaults()

		if gs.SaveId != base {
			log.Warn().Msgf("Mismatched saveId in localStorage %s (expected %s, got %s)", key, base, gs.SaveId)
			continue
		}

		saves[gs.SaveId] = &gs
	}

	return saves, nil
}
