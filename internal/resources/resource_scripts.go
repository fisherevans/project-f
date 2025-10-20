package resources

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

var (
	scriptResources = map[string]*compiledScript{}
)

type compiledScript struct {
	name   string
	source string
	prog   *goja.Program
}

// GetScriptNames returns all loaded script names
func GetScriptNames() []string {
	names := make([]string, 0, len(scriptResources))
	for name := range scriptResources {
		names = append(names, name)
	}
	return names
}

func loadScript(path string, name string, data []byte) error {
	return RegisterCompiledScript(name, string(data))
}

// RegisterCompiledScript is called during init() for each .js file
func RegisterCompiledScript(name string, source string) error {
	registered, exists := scriptResources[name]
	if exists {
		if registered.source == source {
			log.Warn().Msgf("Compiled script '%s' was registered reduntantly", name)
			return nil
		}
		return fmt.Errorf("compiled script '%s' already registered with different source", name)
	}
	prog, err := goja.Compile(name, source, true) // strict=true
	if err != nil {
		return fmt.Errorf("failed to compile script %s: %w", name, err)
	}
	scriptResources[name] = &compiledScript{
		name:   name,
		source: source,
		prog:   prog,
	}
	return nil
}

func GetScript(name string) *goja.Program {
	script, exists := scriptResources[name]
	if !exists {
		return nil
	}
	return script.prog
}
