//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type EventInfo struct {
	TypeName       string
	JSFunctionName string
	Fields         []FieldInfo
}

type FieldInfo struct {
	Name   string
	GoType string
	TSType string
	JSName string
}

type EffectInfo struct {
	Name   string
	Fields []FieldInfo
}

func main() {
	// Parse handler.go to extract event registrations
	events := parseEventRegistrations("handler.go")

	// Parse effect.go to extract effect types
	effects := parseEffectTypes("effect.go")

	// Generate TypeScript definitions
	output := generateTypeScript(events, effects)

	// Write to assets/scripts/global.d.ts
	outputPath := filepath.Join("..", "..", "..", "assets", "scripts", "global.d.ts")
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	// Generate JSDoc template
	templateOutput := generateJSDocTemplate(events)
	templatePath := filepath.Join("..", "..", "..", "assets", "scripts", "template.js")
	if err := os.WriteFile(templatePath, []byte(templateOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated %s\n", outputPath)
	fmt.Printf("Generated %s\n", templatePath)
}

func parseEventRegistrations(filename string) []EventInfo {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filename, err)
		os.Exit(1)
	}

	var events []EventInfo
	registrations := make(map[string]string) // TypeName -> JSFunctionName

	// First pass: collect registrations from init()
	ast.Inspect(node, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if selExpr, ok := callExpr.Fun.(*ast.IndexExpr); ok {
				if ident, ok := selExpr.X.(*ast.Ident); ok && ident.Name == "registerGojaEventHandler" {
					if typeIdent, ok := selExpr.Index.(*ast.Ident); ok {
						if len(callExpr.Args) > 0 {
							if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
								jsFuncName := strings.Trim(lit.Value, `"`)
								registrations[typeIdent.Name] = jsFuncName
							}
						}
					}
				}
			}
		}
		return true
	})

	// Second pass: collect struct definitions
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							typeName := typeSpec.Name.Name
							if jsFuncName, exists := registrations[typeName]; exists {
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
							// Only process Effect* types (but not the main Effect struct)
							if strings.HasPrefix(typeName, "Effect") && typeName != "Effect" {
								effects = append(effects, info)
							} else if typeName == "DynamicAnimationReference" || typeName == "LightConfig" {
								// Include helper types
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

		fields = append(fields, FieldInfo{
			Name:   fieldName,
			GoType: goType,
			TSType: tsType,
			JSName: jsName,
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

func generateTypeScript(events []EventInfo, effects []EffectInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED - DO NOT EDIT\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// Source: internal/game/events/handler.go, internal/game/events/effect.go\n\n")

	sb.WriteString("// ============================================================================\n")
	sb.WriteString("// LOGGING\n")
	sb.WriteString("// ============================================================================\n\n")
	sb.WriteString("interface Logger {\n")
	sb.WriteString("  Debug(message: string): void;\n")
	sb.WriteString("  Info(message: string): void;\n")
	sb.WriteString("  Warn(message: string): void;\n")
	sb.WriteString("  Error(message: string): void;\n")
	sb.WriteString("}\n\n")
	sb.WriteString("declare const log: Logger;\n\n")

	sb.WriteString("// ============================================================================\n")
	sb.WriteString("// CONTEXT\n")
	sb.WriteString("// ============================================================================\n\n")
	sb.WriteString("interface EntityContext {\n")
	sb.WriteString("  Id(): string;\n")
	sb.WriteString("}\n\n")
	sb.WriteString("interface WorldStateReader {\n")
	sb.WriteString("  // Add world state methods as needed\n")
	sb.WriteString("}\n\n")

	// Generate event types
	sb.WriteString("// ============================================================================\n")
	sb.WriteString("// EVENT TYPES\n")
	sb.WriteString("// ============================================================================\n\n")

	for _, event := range events {
		sb.WriteString(fmt.Sprintf("interface %s {\n", event.TypeName))
		for _, field := range event.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", field.JSName, field.TSType))
		}
		sb.WriteString("}\n\n")
	}

	// Generate effect types
	sb.WriteString("// ============================================================================\n")
	sb.WriteString("// EFFECT TYPES\n")
	sb.WriteString("// ============================================================================\n\n")

	var effectTypes []EffectInfo
	for _, effect := range effects {
		sb.WriteString(fmt.Sprintf("interface %s {\n", effect.Name))
		for _, field := range effect.Fields {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", field.JSName, field.TSType))
		}
		sb.WriteString("}\n\n")
		
		// Track actual Effect* types (not helper types)
		if strings.HasPrefix(effect.Name, "Effect") {
			effectTypes = append(effectTypes, effect)
		}
	}

	// Generate Effect union type (only for Effect* types)
	sb.WriteString("interface Effect {\n")
	for _, effect := range effectTypes {
		fieldName := strings.TrimPrefix(effect.Name, "Effect")
		fieldName = toJSName(fieldName)
		sb.WriteString(fmt.Sprintf("  %s?: %s;\n", fieldName, effect.Name))
	}
	sb.WriteString("}\n\n")

	// Generate HandlerOutput
	sb.WriteString("interface HandlerOutput {\n")
	sb.WriteString("  state?: any;\n")
	sb.WriteString("  effects?: Effect[];\n")
	sb.WriteString("}\n\n")

	// Generate EntityHandler interface
	sb.WriteString("// ============================================================================\n")
	sb.WriteString("// ENTITY HANDLER\n")
	sb.WriteString("// ============================================================================\n\n")
	sb.WriteString("interface EntityHandler {\n")
	sb.WriteString("  Init?: (self: EntityContext, world: WorldStateReader, state: any) => HandlerOutput | null;\n")
	for _, event := range events {
		sb.WriteString(fmt.Sprintf("  %s?: (self: EntityContext, world: WorldStateReader, state: any, event: %s) => HandlerOutput | null;\n",
			event.JSFunctionName, event.TypeName))
	}
	sb.WriteString("}\n\n")

	sb.WriteString("declare const handler: EntityHandler;\n")

	return sb.String()
}

func generateJSDocTemplate(events []EventInfo) string {
	var sb strings.Builder

	sb.WriteString("// AUTO-GENERATED TEMPLATE - Copy and customize for your entity scripts\n")
	sb.WriteString(fmt.Sprintf("// Generated at %s by go generate\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("// @ts-check\n\n")

	sb.WriteString("/** @type {import('./global').EntityHandler} */\n")
	sb.WriteString("const handler = {\n")
	sb.WriteString("    Init(self, world, state) {\n")
	sb.WriteString("        return {\n")
	sb.WriteString("            state: state || {},\n")
	sb.WriteString("        };\n")
	sb.WriteString("    },\n\n")

	for i, event := range events {
		sb.WriteString(fmt.Sprintf("    %s(self, world, state, event) {\n", event.JSFunctionName))
		sb.WriteString("        return {\n")
		sb.WriteString("            state: state,\n")
		sb.WriteString("            effects: []\n")
		sb.WriteString("        };\n")
		sb.WriteString("    }")
		if i < len(events)-1 {
			sb.WriteString(",\n\n")
		} else {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("};\n")

	return sb.String()
}
