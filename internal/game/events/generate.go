//go:build ignore

//go:generate go run generate.go

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"time"
)

type EventInfo struct {
	TypeName       string
	JSFunctionName string
	Fields         []FieldInfo
}

type FieldInfo struct {
	Name         string
	GoType       string
	TSType       string
	JSName       string
	AutoGenerate bool
	Optional     bool
	OneOf        string // group name for one_of validation
}

type EffectInfo struct {
	Name   string
	Fields []FieldInfo
}

func main() {
	// Parse events.go to extract event registrations
	events := parseEventRegistrations("events.go")

	// Parse effects.go to extract effect types
	effects := parseEffectTypes("effects.go")

	// Generate Go BasicHandlerBuilder
	builderOutput := generateBasicHandlerBuilder(events)
	builderPath := "handler_builder.generated.go"
	if err := os.WriteFile(builderPath, []byte(builderOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing builder: %v\n", err)
		os.Exit(1)
	}

	// Generate Effect struct
	effectStructOutput := generateEffectStruct(effects)
	effectStructPath := "effect_struct.generated.go"
	if err := os.WriteFile(effectStructPath, []byte(effectStructOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing effect struct: %v\n", err)
		os.Exit(1)
	}

	// Generate effect validation
	validationOutput := generateEffectValidation(effects)
	validationPath := "effect_validation.generated.go"
	if err := os.WriteFile(validationPath, []byte(validationOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing validation: %v\n", err)
		os.Exit(1)
	}

	// Generate effect builders
	buildersOutput := generateEffectBuilders(effects)
	buildersPath := "effect_builders.generated.go"
	if err := os.WriteFile(buildersPath, []byte(buildersOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing effect builders: %v\n", err)
		os.Exit(1)
	}

	// Generate chainable effect builder
	chainBuilderOutput := generateChainableEffectBuilder(effects)
	chainBuilderPath := "effect_chain_builder.generated.go"
	if err := os.WriteFile(chainBuilderPath, []byte(chainBuilderOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing chainable effect builder: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", builderPath)
	fmt.Printf("Generated %s\n", effectStructPath)
	fmt.Printf("Generated %s\n", validationPath)
	fmt.Printf("Generated %s\n", buildersPath)
	fmt.Printf("Generated %s\n", chainBuilderPath)
}

func parseEventRegistrations(filename string) []EventInfo {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	var events []EventInfo

	// Auto-discover all Event* structs
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							typeName := typeSpec.Name.Name
							// Only include structs starting with "Event"
							if strings.HasPrefix(typeName, "Event") {
								// Convert EventOnInteract -> OnInteract
								jsFuncName := typeName[len("Event"):]
								event := EventInfo{
									TypeName:       typeName,
									JSFunctionName: jsFuncName,
									Fields:         parseStructFields(structType),
								}
								events = append(events, event)
							}
						}
					}
				}
			}
		}
		return true
	})

	return events
}

func parseEffectTypes(filename string) []EffectInfo {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	var effects []EffectInfo
	var helperTypes []EffectInfo

	// Auto-discover all Effect* structs and helper types
	ast.Inspect(node, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok {
			if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						typeName := typeSpec.Name.Name
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							info := EffectInfo{
								Name:   typeName,
								Fields: parseStructFields(structType),
							}
							// Include all Effect* types
							if strings.HasPrefix(typeName, "Effect") {
								effects = append(effects, info)
							} else if typeName == "DynamicAnimationReference" || typeName == "LightConfig" || typeName == "Location" || typeName == "PlanStep" {
								// Include helper types for TypeScript generation
								helperTypes = append(helperTypes, info)
							}
						}
					}
				}
			}
		}
		return true
	})

	// Prepend helper types so they're defined first
	return append(helperTypes, effects...)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseStructFields(structType *ast.StructType) []FieldInfo {
	var fields []FieldInfo
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue // Skip embedded fields
		}
		fieldName := field.Names[0].Name
		goType := exprToString(field.Type)
		tsType := goTypeToTS(goType)
		jsName := toJSName(fieldName)

		// Parse struct tags
		autoGenerate := false
		optional := false
		oneOf := ""
		if field.Tag != nil {
			tagValue := field.Tag.Value
			// Remove backticks
			tagValue = strings.Trim(tagValue, "`")

			if strings.Contains(tagValue, `auto_generate:"true"`) {
				autoGenerate = true
			}
			if strings.Contains(tagValue, `optional:"true"`) {
				optional = true
			}
			// Extract one_of value
			if idx := strings.Index(tagValue, `one_of:"`); idx != -1 {
				start := idx + len(`one_of:"`)
				end := strings.Index(tagValue[start:], `"`)
				if end != -1 {
					oneOf = tagValue[start : start+end]
				}
			}
		}

		fields = append(fields, FieldInfo{
			Name:         fieldName,
			GoType:       goType,
			TSType:       tsType,
			JSName:       jsName,
			AutoGenerate: autoGenerate,
			Optional:     optional,
			OneOf:        oneOf,
		})
	}
	return fields
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

func goTypeToTS(goType string) string {
	// Handle pointers
	isOptional := false
	if strings.HasPrefix(goType, "*") {
		isOptional = true
		goType = strings.TrimPrefix(goType, "*")
	}

	var tsType string
	switch goType {
	case "string":
		tsType = "string"
	case "int", "int64", "int32", "float64", "float32":
		tsType = "number"
	case "bool":
		tsType = "boolean"
	case "DynamicAnimationReference", "LightConfig":
		// Use the TypeScript interface name directly
		tsType = goType
	default:
		if strings.HasPrefix(goType, "[]") {
			elemType := goTypeToTS(strings.TrimPrefix(goType, "[]"))
			tsType = elemType + "[]"
		} else if strings.HasPrefix(goType, "map[") {
			// map[string]T -> Record<string, T>
			parts := strings.SplitN(goType, "]", 2)
			if len(parts) == 2 {
				keyType := goTypeToTS(strings.TrimPrefix(parts[0], "map["))
				valType := goTypeToTS(parts[1])
				tsType = fmt.Sprintf("Record<%s, %s>", keyType, valType)
			} else {
				tsType = "Record<string, any>"
			}
		} else {
			tsType = "any"
		}
	}

	if isOptional {
		tsType += " | undefined"
	}

	return tsType
}

func toJSName(goName string) string {
	if len(goName) == 0 {
		return goName
	}
	// Convert PascalCase to camelCase
	return strings.ToLower(goName[:1]) + goName[1:]
}

func camelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func generateBasicHandlerBuilder(events []EventInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/handler.go\n\n")

	sb.WriteString("package events\n\n")

	// Generate BasicHandlerBuilder struct
	sb.WriteString("// BasicHandlerBuilder provides a simple way to build event handlers with type-safe state\n")
	sb.WriteString("type BasicHandlerBuilder[T any] struct {\n")
	sb.WriteString("\tDefaultState func() T\n\n")
	sb.WriteString("\tInit func(ctx EntityContext, world WorldStateReader, state T) *HandlerOutput\n\n")

	for _, event := range events {
		sb.WriteString(fmt.Sprintf("\t%s func(ctx EntityContext, world WorldStateReader, state T, event *%s) *HandlerOutput\n",
			event.JSFunctionName, event.TypeName))
	}
	sb.WriteString("}\n\n")

	// Generate NewBasicHandler constructor
	sb.WriteString("func NewBasicHandler[T any](defaultState T) *BasicHandlerBuilder[T] {\n")
	sb.WriteString("\treturn &BasicHandlerBuilder[T]{\n")
	sb.WriteString("\t\tDefaultState: func() T {\n")
	sb.WriteString("\t\t\treturn defaultState\n")
	sb.WriteString("\t\t},\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Generate fluent builder methods
	sb.WriteString("func (b *BasicHandlerBuilder[T]) WithInit(init func(ctx EntityContext, world WorldStateReader, state T) *HandlerOutput) *BasicHandlerBuilder[T] {\n")
	sb.WriteString("\tb.Init = init\n")
	sb.WriteString("\treturn b\n")
	sb.WriteString("}\n\n")

	for _, event := range events {
		methodName := "With" + event.JSFunctionName
		sb.WriteString(fmt.Sprintf("func (b *BasicHandlerBuilder[T]) %s(%s func(ctx EntityContext, world WorldStateReader, state T, event *%s) *HandlerOutput) *BasicHandlerBuilder[T] {\n",
			methodName, camelCase(event.JSFunctionName), event.TypeName))
		sb.WriteString(fmt.Sprintf("\tb.%s = %s\n", event.JSFunctionName, camelCase(event.JSFunctionName)))
		sb.WriteString("\treturn b\n")
		sb.WriteString("}\n\n")
	}

	// Generate CreateHandler method
	sb.WriteString("// CreateHandler creates an EventHandler from the builder\n")
	sb.WriteString("func (b BasicHandlerBuilder[T]) CreateHandler() EventHandler {\n")
	sb.WriteString("\treturn &basicHandler[T]{\n")
	sb.WriteString("\t\tbuilder: b,\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	// Generate basicHandler struct
	sb.WriteString("type basicHandler[T any] struct {\n")
	sb.WriteString("\tbuilder BasicHandlerBuilder[T]\n")
	sb.WriteString("}\n\n")

	// Generate convertState helper
	sb.WriteString("func (h *basicHandler[T]) convertState(original any) T {\n")
	sb.WriteString("\tif original == nil {\n")
	sb.WriteString("\t\tif h.builder.DefaultState != nil {\n")
	sb.WriteString("\t\t\treturn h.builder.DefaultState()\n")
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t\tvar zero T\n")
	sb.WriteString("\t\treturn zero\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\tv, ok := original.(T)\n")
	sb.WriteString("\tif !ok {\n")
	sb.WriteString("\t\tvar zero T\n")
	sb.WriteString("\t\treturn zero\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn v\n")
	sb.WriteString("}\n\n")

	// Generate Init method
	sb.WriteString("func (h *basicHandler[T]) Init(ctx EntityContext, world WorldStateReader, state any) *HandlerOutput {\n")
	sb.WriteString("\tif h.builder.Init == nil {\n")
	sb.WriteString("\t\treturn nil\n")
	sb.WriteString("\t}\n")
	sb.WriteString("\treturn h.builder.Init(ctx, world, h.convertState(state))\n")
	sb.WriteString("}\n\n")

	// Generate HandleEvent method
	sb.WriteString("func (h *basicHandler[T]) HandleEvent(ctx EntityContext, world WorldStateReader, state any, event any) *HandlerOutput {\n")
	sb.WriteString("\tconvertedState := h.convertState(state)\n")
	sb.WriteString("\tswitch e := event.(type) {\n")

	for _, event := range events {
		sb.WriteString(fmt.Sprintf("\tcase *%s:\n", event.TypeName))
		sb.WriteString(fmt.Sprintf("\t\tif h.builder.%s == nil {\n", event.JSFunctionName))
		sb.WriteString("\t\t\treturn nil\n")
		sb.WriteString("\t\t}\n")
		sb.WriteString(fmt.Sprintf("\t\treturn h.builder.%s(ctx, world, convertedState, e)\n", event.JSFunctionName))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}

func generateEffectStruct(effects []EffectInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/effects.go\n\n")

	sb.WriteString("package events\n\n")

	sb.WriteString("// Effect represents a single effect that can be applied\n")
	sb.WriteString("// Only one effect type should be set per Effect instance\n")
	sb.WriteString("type Effect struct {\n")

	for _, effect := range effects {
		// Skip helper types - only include Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		fieldName := effect.Name[len("Effect"):] // Remove "Effect" prefix
		sb.WriteString(fmt.Sprintf("\t%s *%s\n", fieldName, effect.Name))
	}

	sb.WriteString("}\n")

	return sb.String()
}

func generateEffectValidation(effects []EffectInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/effect.go\n\n")

	sb.WriteString("package events\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"fmt\"\n")
	sb.WriteString("\t\"sync/atomic\"\n")
	sb.WriteString(")\n\n")

	// Generate counter variables for each effect type that has auto_generate fields
	for _, effect := range effects {
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}
		hasAutoGen := false
		for _, field := range effect.Fields {
			if field.AutoGenerate {
				hasAutoGen = true
				break
			}
		}
		if hasAutoGen {
			effectName := effect.Name[len("Effect"):]
			counterName := fmt.Sprintf("%sCounter", camelCase(effectName))
			sb.WriteString(fmt.Sprintf("var %s atomic.Uint64\n", counterName))
		}
	}
	sb.WriteString("\n")

	// First pass: determine which types need validation
	typesWithValidation := make(map[string]bool)
	// The Effect union type always needs validation
	typesWithValidation["Effect"] = true
	for _, effect := range effects {
		if strings.HasPrefix(effect.Name, "Effect") {
			typesWithValidation[effect.Name] = true
		} else {
			// Check if helper type needs validation
			needsValidation := false
			for _, field := range effect.Fields {
				if field.OneOf != "" {
					needsValidation = true
					break
				}
			}
			if needsValidation {
				typesWithValidation[effect.Name] = true
			}
		}
	}

	// Generate Validate method for helper types first (they may be used by Effect types)
	for _, effect := range effects {
		// Only process helper types (non-Effect types) that need validation
		if strings.HasPrefix(effect.Name, "Effect") {
			continue
		}
		if !typesWithValidation[effect.Name] {
			continue
		}

		generateValidateMethod(&sb, effect, typesWithValidation)
	}

	// Generate Validate method for each Effect* type
	for _, effect := range effects {
		// Only process Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		generateValidateMethod(&sb, effect, typesWithValidation)
	}

	// Generate Validate method for Effect union
	sb.WriteString("// Validate checks that the effect has exactly one effect type set and calls its validator\n")
	sb.WriteString("func (e Effect) Validate() error {\n")
	sb.WriteString("\treporter := newIssueReporter()\n")
	sb.WriteString("\teffectCount := 0\n\n")

	for _, effect := range effects {
		// Skip helper types - only process Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		fieldName := effect.Name[len("Effect"):] // Remove "Effect" prefix to get field name
		sb.WriteString(fmt.Sprintf("\tif e.%s != nil {\n", fieldName))
		sb.WriteString("\t\teffectCount++\n")
		sb.WriteString(fmt.Sprintf("\t\tif err := e.%s.Validate(); err != nil {\n", fieldName))
		sb.WriteString(fmt.Sprintf("\t\t\treporter.sub(%q).addf(\"\", \"%%v\", err)\n", camelCase(fieldName)))
		sb.WriteString("\t\t}\n")
		sb.WriteString("\t}\n")
	}

	sb.WriteString("\n\treporter.requirePositive(\"effectCount\", float64(effectCount))\n")
	sb.WriteString("\treturn reporter.report()\n")
	sb.WriteString("}\n")

	return sb.String()
}

func generateValidateMethod(sb *strings.Builder, effect EffectInfo, typesWithValidation map[string]bool) {
	receiverName := "e"
	if !strings.HasPrefix(effect.Name, "Effect") {
		receiverName = "p" // Use 'p' for helper types like PlanStep
	}

	// Helper to check if a type needs validation
	needsValidation := func(goType string) bool {
		// Remove pointer prefix
		goType = strings.TrimPrefix(goType, "*")
		// Remove slice prefix
		goType = strings.TrimPrefix(goType, "[]")
		// Check if it's in our map
		return typesWithValidation[goType]
	}

	sb.WriteString(fmt.Sprintf("func (%s *%s) Validate() error {\n", receiverName, effect.Name))
	sb.WriteString("\treporter := newIssueReporter()\n\n")

	// Auto-generate fields (only for Effect* types)
	if strings.HasPrefix(effect.Name, "Effect") {
		for _, field := range effect.Fields {
			if field.AutoGenerate && !strings.HasPrefix(field.GoType, "*") {
				effectName := effect.Name[len("Effect"):]
				counterName := fmt.Sprintf("%sCounter", camelCase(effectName))
				prefix := toKebabCase(effectName)

				sb.WriteString(fmt.Sprintf("\tif %s.%s == \"\" {\n", receiverName, field.Name))
				sb.WriteString(fmt.Sprintf("\t\tid := %s.Add(1)\n", counterName))
				sb.WriteString(fmt.Sprintf("\t\t%s.%s = fmt.Sprintf(\"%s-%%05d\", id)\n", receiverName, field.Name, prefix))
				sb.WriteString("\t}\n\n")
			}
		}
	}

	// Group one_of fields
	oneOfGroups := make(map[string][]FieldInfo)
	for _, field := range effect.Fields {
		if field.OneOf != "" {
			oneOfGroups[field.OneOf] = append(oneOfGroups[field.OneOf], field)
		}
	}

	// Validate one_of groups
	for groupName, groupFields := range oneOfGroups {
		sb.WriteString(fmt.Sprintf("\t// Validate one_of group: %s\n", groupName))
		sb.WriteString(fmt.Sprintf("\t%sCount := 0\n", groupName))
		for _, field := range groupFields {
			// For slices in one_of, check if they're non-empty
			if strings.HasPrefix(field.GoType, "[]") {
				sb.WriteString(fmt.Sprintf("\tif len(%s.%s) > 0 {\n", receiverName, field.Name))
			} else {
				sb.WriteString(fmt.Sprintf("\tif %s.%s != nil {\n", receiverName, field.Name))
			}
			sb.WriteString(fmt.Sprintf("\t\t%sCount++\n", groupName))
			// Validate nested struct if it needs validation
			fieldType := strings.TrimPrefix(field.GoType, "*")
			if strings.HasPrefix(fieldType, "[]") {
				// Handle slices
				elemType := strings.TrimPrefix(fieldType, "[]")
				if needsValidation(elemType) {
					sb.WriteString(fmt.Sprintf("\t\tfor i, item := range %s.%s {\n", receiverName, field.Name))
					sb.WriteString("\t\t\tif err := item.Validate(); err != nil {\n")
					sb.WriteString(fmt.Sprintf("\t\t\t\treporter.sub(%q).sub(fmt.Sprintf(\"[%%d]\", i)).addf(\"\", \"%%v\", err)\n", camelCase(field.Name)))
					sb.WriteString("\t\t\t}\n")
					sb.WriteString("\t\t}\n")
				}
			} else if needsValidation(fieldType) {
				sb.WriteString(fmt.Sprintf("\t\tif err := %s.%s.Validate(); err != nil {\n", receiverName, field.Name))
				sb.WriteString(fmt.Sprintf("\t\t\treporter.sub(%q).addf(\"\", \"%%v\", err)\n", camelCase(field.Name)))
				sb.WriteString("\t\t}\n")
			}
			sb.WriteString("\t}\n")
		}
		sb.WriteString(fmt.Sprintf("\tif %sCount != 1 {\n", groupName))
		sb.WriteString(fmt.Sprintf("\t\treporter.addf(%q, \"exactly one of [%s] must be set\")\n",
			groupName, getFieldNames(groupFields)))
		sb.WriteString("\t}\n\n")
	}

	// Validate required fields (non-pointer, non-optional, not in one_of)
	for _, field := range effect.Fields {
		if strings.HasPrefix(field.GoType, "*") || field.Optional || field.OneOf != "" {
			continue
		}

		fieldType := field.GoType
		if fieldType == "string" {
			sb.WriteString(fmt.Sprintf("\treporter.requireString(%q, %s.%s)\n", camelCase(field.Name), receiverName, field.Name))
		} else if isNumericType(fieldType) {
			// For numeric types, we might want different validation
			// For now, just check they're set (non-zero)
		} else if strings.HasPrefix(fieldType, "[]") {
			// Handle slices
			elemType := strings.TrimPrefix(fieldType, "[]")
			if needsValidation(elemType) {
				sb.WriteString(fmt.Sprintf("\tfor i, item := range %s.%s {\n", receiverName, field.Name))
				sb.WriteString("\t\tif err := item.Validate(); err != nil {\n")
				sb.WriteString(fmt.Sprintf("\t\t\treporter.sub(%q).sub(fmt.Sprintf(\"[%%d]\", i)).addf(\"\", \"%%v\", err)\n", camelCase(field.Name)))
				sb.WriteString("\t\t}\n")
				sb.WriteString("\t}\n")
			}
		} else if needsValidation(fieldType) {
			// For struct types that need validation
			sb.WriteString(fmt.Sprintf("\tif err := %s.%s.Validate(); err != nil {\n", receiverName, field.Name))
			sb.WriteString(fmt.Sprintf("\t\treporter.sub(%q).addf(\"\", \"%%v\", err)\n", camelCase(field.Name)))
			sb.WriteString("\t}\n")
		}
	}

	// Validate optional pointer fields that are set
	for _, field := range effect.Fields {
		if !strings.HasPrefix(field.GoType, "*") || field.OneOf != "" {
			continue
		}

		fieldType := strings.TrimPrefix(field.GoType, "*")
		if needsValidation(fieldType) {
			sb.WriteString(fmt.Sprintf("\tif %s.%s != nil {\n", receiverName, field.Name))
			sb.WriteString(fmt.Sprintf("\t\tif err := %s.%s.Validate(); err != nil {\n", receiverName, field.Name))
			sb.WriteString(fmt.Sprintf("\t\t\treporter.sub(%q).addf(\"\", \"%%v\", err)\n", camelCase(field.Name)))
			sb.WriteString("\t\t}\n")
			sb.WriteString("\t}\n")
		}
	}

	sb.WriteString("\n\treturn reporter.report()\n")
	sb.WriteString("}\n\n")
}

func isBasicType(goType string) bool {
	basicTypes := []string{"string", "int", "int32", "int64", "float32", "float64", "bool", "byte", "rune", "any", "RunnableFunction"}
	for _, bt := range basicTypes {
		if goType == bt {
			return true
		}
	}
	// Also consider types from other packages as basic (they won't have Validate methods we generate)
	if strings.Contains(goType, ".") {
		return true
	}
	// Arrays and maps are basic
	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[") {
		return true
	}
	// Types that get Validate() methods generated are NOT basic
	// This includes Effect* types and helper types like PlanStep
	return false
}

func isNumericType(goType string) bool {
	numericTypes := []string{"int", "int32", "int64", "float32", "float64", "byte"}
	for _, nt := range numericTypes {
		if goType == nt {
			return true
		}
	}
	return false
}

func getFieldNames(fields []FieldInfo) string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = camelCase(f.Name)
	}
	return strings.Join(names, ", ")
}

func generateEffectBuilders(effects []EffectInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/effects.go\n\n")

	sb.WriteString("package events\n\n")
	sb.WriteString("import (\n")
	sb.WriteString("\t\"fisherevans.com/project/f/internal/game/input\"\n")
	sb.WriteString("\t\"fisherevans.com/project/f/internal/game/rpg\"\n")
	sb.WriteString("\t\"fisherevans.com/project/f/internal/game/states/adventure/types\"\n")
	sb.WriteString(")\n\n")

	for _, effect := range effects {
		// Skip helper types - only process Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		effectName := effect.Name[len("Effect"):] // Remove "Effect" prefix

		// Separate required (non-pointer) and optional (pointer) fields
		var requiredFields []FieldInfo
		var optionalFields []FieldInfo

		for _, field := range effect.Fields {
			if strings.HasPrefix(field.GoType, "*") {
				optionalFields = append(optionalFields, field)
			} else {
				requiredFields = append(requiredFields, field)
			}
		}

		// Generate constructor function
		sb.WriteString(fmt.Sprintf("func New%sEffect(", effectName))
		for i, field := range requiredFields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s %s", camelCase(field.Name), field.GoType))
		}
		sb.WriteString(fmt.Sprintf(") *%s {\n", effect.Name))
		sb.WriteString(fmt.Sprintf("\treturn &%s{\n", effect.Name))
		for _, field := range requiredFields {
			sb.WriteString(fmt.Sprintf("\t\t%s: %s,\n", field.Name, camelCase(field.Name)))
		}
		sb.WriteString("\t}\n")
		sb.WriteString("}\n\n")

		// Generate With* methods for optional fields
		for _, field := range optionalFields {
			methodName := "With" + field.Name
			paramType := strings.TrimPrefix(field.GoType, "*")
			paramName := camelCase(field.Name)

			sb.WriteString(fmt.Sprintf("func (e *%s) %s(%s %s) *%s {\n",
				effect.Name, methodName, paramName, paramType, effect.Name))
			sb.WriteString(fmt.Sprintf("\te.%s = &%s\n", field.Name, paramName))
			sb.WriteString("\treturn e\n")
			sb.WriteString("}\n\n")
		}
	}

	return sb.String()
}

func generateChainableEffectBuilder(effects []EffectInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/effects.go\n\n")

	sb.WriteString("package events\n\n")

	// Generate NewEffect constructor
	sb.WriteString("// NewEffect creates a new Effect that can be chained with With methods\n")
	sb.WriteString("func NewEffect() *Effect {\n")
	sb.WriteString("\treturn &Effect{}\n")
	sb.WriteString("}\n\n")

	// Generate With method for each effect type
	for _, effect := range effects {
		// Skip helper types - only process Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		effectName := effect.Name[len("Effect"):] // Remove "Effect" prefix

		sb.WriteString(fmt.Sprintf("// With%s sets the %s field and returns the Effect for chaining\n", effectName, effectName))
		sb.WriteString(fmt.Sprintf("func (e *Effect) With%s(v *%s) *Effect {\n", effectName, effect.Name))
		sb.WriteString(fmt.Sprintf("\te.%s = v\n", effectName))
		sb.WriteString("\treturn e\n")
		sb.WriteString("}\n\n")
	}

	// Generate generic With method that accepts any effect pointer
	sb.WriteString("// With sets an effect field by inspecting the type and returns the Effect for chaining\n")
	sb.WriteString("func (e *Effect) With(v any) *Effect {\n")
	sb.WriteString("\tswitch val := v.(type) {\n")

	for _, effect := range effects {
		// Skip helper types - only process Effect* types
		if !strings.HasPrefix(effect.Name, "Effect") {
			continue
		}

		effectName := effect.Name[len("Effect"):] // Remove "Effect" prefix
		sb.WriteString(fmt.Sprintf("\tcase *%s:\n", effect.Name))
		sb.WriteString(fmt.Sprintf("\t\treturn e.With%s(val)\n", effectName))
	}

	sb.WriteString("\tdefault:\n")
	sb.WriteString("\t\t// Unknown type, ignore\n")
	sb.WriteString("\t\treturn e\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}
