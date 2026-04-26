package adventure

import (
	"fmt"
	"io/fs"
	"strings"

	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/internal/util"
	"github.com/rs/zerolog/log"
)

var (
	scriptHandlerDefs      = map[string]*HandlerDef{}
	scriptSequences        = map[string]*SequenceDef{}
	scriptConsts           = map[string]any{}
	scriptCustomActions    = map[string]*CustomActionDef{}
	scriptPropertyTemplates = map[string]map[string]any{}
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
		for name, val := range sf.Consts {
			if _, exists := scriptConsts[name]; exists {
				log.Fatal().Str("name", name).Str("path", path).Msg("duplicate const name")
			}
			scriptConsts[name] = val
		}
		for name, action := range sf.CustomActions {
			if _, exists := scriptCustomActions[name]; exists {
				log.Fatal().Str("name", name).Str("path", path).Msg("duplicate custom action name")
			}
			scriptCustomActions[name] = action
			log.Debug().Str("name", name).Str("path", path).Msg("loaded custom action")
		}
		for name, tmpl := range sf.PropertyTemplates {
			if _, exists := scriptPropertyTemplates[name]; exists {
				log.Fatal().Str("name", name).Str("path", path).Msg("duplicate property template name")
			}
			scriptPropertyTemplates[name] = tmpl
		}
		return nil
	})
	if err != nil {
		log.Warn().Err(err).Msg("failed to walk scripts directory")
	}
	// Merge YAML property templates into Go-side registry
	for name, tmpl := range scriptPropertyTemplates {
		if _, exists := entityPropertyTemplates[name]; exists {
			log.Fatal().Str("name", name).Msg("YAML property template conflicts with Go-registered template")
		}
		entityPropertyTemplates[name] = tmpl
	}

	log.Info().
		Int("handlers", len(scriptHandlerDefs)).
		Int("sequences", len(scriptSequences)).
		Int("consts", len(scriptConsts)).
		Int("custom_actions", len(scriptCustomActions)).
		Int("property_templates", len(scriptPropertyTemplates)).
		Msg("loaded script files")

	validateScriptFiles()
}

func registerScriptHandlerFactory() {
	for name, def := range scriptHandlerDefs {
		handlerName := name
		handlerDef := def
		if _, exists := eventHandlerRegistry[handlerName]; exists {
			log.Debug().Str("name", handlerName).Msg("script handler skipped - Go handler already registered")
			continue
		}
		registerEventHandler(handlerName, func(props *util.Properties) EventHandler {
			var rawProps map[string]any
			if props != nil {
				rawProps = props.All()
			}
			return newScriptHandlerFactory(handlerDef, scriptSequences, templateContextFromProps(props), rawProps)
		})
	}
}

func getDataList(name string) ([]string, bool) {
	val, ok := scriptConsts[name]
	if !ok {
		return nil, false
	}
	if list, ok := val.([]any); ok {
		strs := make([]string, len(list))
		for i, v := range list {
			strs[i] = fmt.Sprintf("%v", v)
		}
		return strs, true
	}
	if list, ok := val.([]string); ok {
		return list, true
	}
	return nil, false
}
