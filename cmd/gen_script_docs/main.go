package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"fisherevans.com/project/f/internal/schema"
)

func main() {
	s, err := schema.LoadScriptSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load schema: %v\n", err)
		os.Exit(1)
	}

	var b strings.Builder

	b.WriteString("# Script System Reference\n\n")
	b.WriteString("Auto-generated from `internal/schema/script_schema.json`. Do not edit manually.\n\n")

	writeStepKinds(&b, s)
	writeActions(&b, s)
	writeConditions(&b, s)
	writeBuiltinConditions(&b, s)
	writeEventHooks(&b, s)
	writeTemplateVars(&b, s)

	outPath := "docs/script_reference.md"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}
	if err := os.WriteFile(outPath, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", outPath, err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s\n", outPath)
}

func writeStepKinds(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Step Kinds\n\n")
	b.WriteString("Each step in a handler's rule is a single-key YAML map: `{kind: params}`.\n\n")

	categories := map[string][]schema.StepKindDef{}
	for _, def := range s.StepKinds {
		categories[def.Category] = append(categories[def.Category], def)
	}
	catNames := sortedKeys(categories)
	for _, cat := range catNames {
		defs := categories[cat]
		sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
		b.WriteString(fmt.Sprintf("### %s\n\n", cat))
		for _, def := range defs {
			b.WriteString(fmt.Sprintf("#### `%s`\n\n", def.Name))
			b.WriteString(def.Description + "\n\n")
			b.WriteString(fmt.Sprintf("- **Param style:** %s\n", def.ParamStyle))
			if def.AcceptsSubSteps {
				b.WriteString("- **Accepts sub-steps:** yes\n")
			}
			writeParams(b, def.Params)
			b.WriteString("\n")
		}
	}
}

func writeActions(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Named Actions\n\n")
	b.WriteString("Invoked via `action: name` or `action: {name: ..., param: value}` in YAML steps.\n\n")

	names := sortedMapKeys(s.Actions)
	for _, name := range names {
		def := s.Actions[name]
		b.WriteString(fmt.Sprintf("### `%s`\n\n", name))
		b.WriteString(def.Description + "\n\n")
		writeParams(b, def.Params)
		b.WriteString("\n")
	}
}

func writeConditions(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Named Conditions\n\n")
	b.WriteString("Used in `wait_for` steps or `when` clauses via `check` delegation.\n\n")

	names := sortedMapKeys(s.Conditions)
	for _, name := range names {
		def := s.Conditions[name]
		b.WriteString(fmt.Sprintf("### `%s`\n\n", name))
		b.WriteString(def.Description + "\n\n")
		writeParams(b, def.Params)
		b.WriteString("\n")
	}
}

func writeBuiltinConditions(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Built-in Conditions\n\n")
	b.WriteString("Used in `when` clauses as single-key maps: `{condition_type: params}`.\n\n")

	names := sortedMapKeys(s.BuiltinConditions)
	for _, name := range names {
		def := s.BuiltinConditions[name]
		b.WriteString(fmt.Sprintf("### `%s`\n\n", name))
		b.WriteString(def.Description + "\n")
		if def.IsComposite {
			b.WriteString("\n*Composite condition - takes sub-conditions as params.*\n")
		}
		b.WriteString("\n")
		writeParams(b, def.Params)
		b.WriteString("\n")
	}
}

func writeEventHooks(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Event Hooks\n\n")
	b.WriteString("Each handler can define rules under these YAML keys. Rules fire when the corresponding event occurs.\n\n")

	names := sortedMapKeys(s.EventHooks)
	for _, name := range names {
		def := s.EventHooks[name]
		b.WriteString(fmt.Sprintf("### `%s`\n\n", def.YAMLKey))
		b.WriteString(def.Description + "\n\n")
		b.WriteString(fmt.Sprintf("- **Event type:** %s\n", def.EventType))
		if len(def.FilterFields) > 0 {
			b.WriteString("- **Filter fields:**\n")
			for _, f := range def.FilterFields {
				b.WriteString(fmt.Sprintf("  - `%s` (%s): %s\n", f.Name, f.Type, f.Description))
			}
		}
		b.WriteString("\n")
	}
}

func writeTemplateVars(b *strings.Builder, s *schema.ScriptSchema) {
	b.WriteString("## Template Variables\n\n")
	b.WriteString("Available in all string values via `{{variable}}` syntax.\n\n")

	for _, v := range s.TemplateVars {
		b.WriteString(fmt.Sprintf("- `%s` - %s\n", v.Pattern, v.Description))
	}
	b.WriteString("\n")
}

func writeParams(b *strings.Builder, params []schema.ParamDef) {
	if len(params) == 0 {
		return
	}
	b.WriteString("**Parameters:**\n\n")
	for _, p := range params {
		req := ""
		if p.Required {
			req = " (required)"
		}
		def := ""
		if p.Default != nil {
			def = fmt.Sprintf(", default: %v", p.Default)
		}
		enum := ""
		if len(p.Enum) > 0 {
			enum = fmt.Sprintf(", values: [%s]", strings.Join(p.Enum, ", "))
		}
		b.WriteString(fmt.Sprintf("- `%s` (%s%s%s%s): %s\n", p.Name, p.Type, req, def, enum, p.Description))
	}
}

func sortedMapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys[V any](m map[string][]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
