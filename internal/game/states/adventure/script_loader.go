package adventure

import (
	"io/fs"
	"strings"

	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

var (
	scriptHandlerDefs = map[string]*HandlerDef{}
	scriptSequences   = map[string]*SequenceDef{}
)

func init() {
	loadScriptFiles()
	registerScriptHandlerFactory()
}

func loadScriptFiles() {
	err := fs.WalkDir(assets.FS, "scripts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		data, err := assets.FS.ReadFile(path)
		if err != nil {
			log.Error().Str("path", path).Err(err).Msg("failed to read script file")
			return nil
		}
		sf, err := ParseScriptFile(data)
		if err != nil {
			log.Error().Str("path", path).Err(err).Msg("failed to parse script file")
			return nil
		}
		for name, seq := range sf.Sequences {
			if _, exists := scriptSequences[name]; exists {
				log.Fatal().Str("name", name).Str("path", path).Msg("duplicate sequence name")
			}
			scriptSequences[name] = seq
			log.Debug().Str("name", name).Str("path", path).Msg("loaded script sequence")
		}
		for name, handler := range sf.Handlers {
			if _, exists := scriptHandlerDefs[name]; exists {
				log.Fatal().Str("name", name).Str("path", path).Msg("duplicate script handler name")
			}
			scriptHandlerDefs[name] = handler
			log.Debug().Str("name", name).Str("path", path).Msg("loaded script handler def")
		}
		return nil
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to walk scripts directory")
	}
	log.Info().
		Int("handlers", len(scriptHandlerDefs)).
		Int("sequences", len(scriptSequences)).
		Msg("loaded script files")
}

func registerScriptHandlerFactory() {
	for name, def := range scriptHandlerDefs {
		handlerName := name
		handlerDef := def
		if _, exists := eventHandlerRegistry[handlerName]; exists {
			log.Debug().Str("name", handlerName).Msg("script handler skipped - Go handler already registered")
			continue
		}
		registerEventHandler(handlerName, func(_ *util.Properties) EventHandler {
			return newScriptHandlerFactory(handlerDef, scriptSequences)
		})
	}
}
