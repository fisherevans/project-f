package adventure

import (
	"fisherevans.com/project/f/assets"
	"fisherevans.com/project/f/internal/util"
	"fmt"
	"github.com/rs/zerolog/log"
	"io/fs"
	"strings"
)

var (
	scriptHandlerDefs       = map[string]*HandlerDef{}
	scriptSequences         = map[string]*SequenceDef{}
	scriptConsts            = map[string]any{}
	scriptCustomActions     = map[string]*CustomActionDef{}
	scriptPropertyTemplates = map[string]map[string]any{}

	// origin tracking so a hot reload can remove only script-registered entries
	// from the shared registries without disturbing Go-registered ones.
	scriptRegisteredHandlers  = map[string]bool{}
	scriptRegisteredTemplates = map[string]bool{}
)

func init() {
	if err := loadScriptFiles(assets.FS); err != nil {
		log.Fatal().Err(err).Msg("failed to load script files")
	}
	registerScriptHandlerFactory()
}

func loadScriptFiles(fsys fs.FS) error {
	walkErr := fs.WalkDir(fsys, "scripts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			log.Error().Str("path", path).Err(err).Msg("failed to read script file")
			return nil
		}
		sf, err := ParseScriptFile(data)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for name, seq := range sf.Sequences {
			if _, exists := scriptSequences[name]; exists {
				return fmt.Errorf("%s: duplicate sequence name %q", path, name)
			}
			scriptSequences[name] = seq
		}
		for name, handler := range sf.Handlers {
			if _, exists := scriptHandlerDefs[name]; exists {
				return fmt.Errorf("%s: duplicate script handler name %q", path, name)
			}
			scriptHandlerDefs[name] = handler
		}
		for name, val := range sf.Consts {
			if _, exists := scriptConsts[name]; exists {
				return fmt.Errorf("%s: duplicate const name %q", path, name)
			}
			scriptConsts[name] = val
		}
		for name, action := range sf.CustomActions {
			if _, exists := scriptCustomActions[name]; exists {
				return fmt.Errorf("%s: duplicate custom action name %q", path, name)
			}
			scriptCustomActions[name] = action
		}
		for name, tmpl := range sf.PropertyTemplates {
			if _, exists := scriptPropertyTemplates[name]; exists {
				return fmt.Errorf("%s: duplicate property template name %q", path, name)
			}
			scriptPropertyTemplates[name] = tmpl
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}

	// Merge YAML property templates into the shared Go-side registry, recording
	// which ones came from scripts so a reload can roll them back.
	for name, tmpl := range scriptPropertyTemplates {
		if _, exists := entityPropertyTemplates[name]; exists && !scriptRegisteredTemplates[name] {
			return fmt.Errorf("YAML property template %q conflicts with Go-registered template", name)
		}
		entityPropertyTemplates[name] = tmpl
		scriptRegisteredTemplates[name] = true
	}

	log.Info().
		Int("handlers", len(scriptHandlerDefs)).
		Int("sequences", len(scriptSequences)).
		Int("consts", len(scriptConsts)).
		Int("custom_actions", len(scriptCustomActions)).
		Int("property_templates", len(scriptPropertyTemplates)).
		Msg("loaded script files")

	if errs := validateScriptFiles(); len(errs) > 0 {
		return fmt.Errorf("script validation failed (%d): %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

func registerScriptHandlerFactory() {
	for name, def := range scriptHandlerDefs {
		handlerName := name
		handlerDef := def
		if _, exists := eventHandlerRegistry[handlerName]; exists && !scriptRegisteredHandlers[handlerName] {
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
		scriptRegisteredHandlers[handlerName] = true
	}
}

// clearScriptRegistries removes all script-origin entries from the shared
// registries and resets the script def maps to empty.
func clearScriptRegistries() {
	for name := range scriptRegisteredHandlers {
		delete(eventHandlerRegistry, name)
	}
	for name := range scriptRegisteredTemplates {
		delete(entityPropertyTemplates, name)
	}
	scriptHandlerDefs = map[string]*HandlerDef{}
	scriptSequences = map[string]*SequenceDef{}
	scriptConsts = map[string]any{}
	scriptCustomActions = map[string]*CustomActionDef{}
	scriptPropertyTemplates = map[string]map[string]any{}
	scriptRegisteredHandlers = map[string]bool{}
	scriptRegisteredTemplates = map[string]bool{}
}

// ReloadScriptDefs re-reads all script YAML from fsys into the script registries.
// It does NOT rebind handlers on entities already loaded in a live map - callers
// should reload the current map afterward (see State.ReloadMap) so entities pick
// up fresh handler instances.
//
// If the new content fails to parse or validate, the registries are restored
// from the embedded (compile-time, known-good) scripts so the running game is
// never left in a half-loaded state, and the original error is returned.
func ReloadScriptDefs(fsys fs.FS) error {
	clearScriptRegistries()
	if err := loadScriptFiles(fsys); err != nil {
		clearScriptRegistries()
		if restoreErr := loadScriptFiles(assets.FS); restoreErr != nil {
			return fmt.Errorf("reload failed: %w; AND restoring embedded scripts failed: %v", err, restoreErr)
		}
		registerScriptHandlerFactory()
		return fmt.Errorf("reload failed, restored embedded scripts: %w", err)
	}
	registerScriptHandlerFactory()
	return nil
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
