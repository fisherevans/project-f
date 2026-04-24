package adventure

import "github.com/rs/zerolog/log"

type ScriptAction func(s *State, source EntityReader, params map[string]any)
type ScriptConditionFactory func(params map[string]any) ConditionCheck

var scriptActions = map[string]ScriptAction{}
var scriptConditions = map[string]ScriptConditionFactory{}

func RegisterScriptAction(name string, fn ScriptAction) {
	if _, exists := scriptActions[name]; exists {
		log.Fatal().Str("name", name).Msg("duplicate script action")
	}
	scriptActions[name] = fn
}

func RegisterScriptCondition(name string, fn ScriptConditionFactory) {
	if _, exists := scriptConditions[name]; exists {
		log.Fatal().Str("name", name).Msg("duplicate script condition")
	}
	scriptConditions[name] = fn
}

func getScriptAction(name string) (ScriptAction, bool) {
	fn, ok := scriptActions[name]
	return fn, ok
}

func getScriptConditionFactory(name string) (ScriptConditionFactory, bool) {
	fn, ok := scriptConditions[name]
	return fn, ok
}
