package delta

import (
	"fmt"
	"reflect"
	"strings"
)

// HumanDiff returns a human-readable delta between old and new.
// It walks structs, pointers, maps, and slices/arrays using reflection.
// Output format examples:
//
//	Field: old -> new
//	Struct.Field: changed
//	MapField[Key]: added/removed/modified
//	SliceField: length changed (N -> M)
//	Root-level creation: "New value created"
func HumanDiff(old, new any) string {
	var b strings.Builder
	diffValue(reflect.ValueOf(old), reflect.ValueOf(new), "", &b)
	return strings.Trim(b.String(), "\n")
}

func diffValue(ov, nv reflect.Value, path string, b *strings.Builder) {
	// Handle invalid values (nil interfaces or missing)
	if !ov.IsValid() && !nv.IsValid() {
		return
	}
	if !ov.IsValid() && nv.IsValid() {
		if path == "" {
			b.WriteString("  New value created\n")
		} else {
			b.WriteString(fmt.Sprintf("  %s: nil -> <set>\n", path))
		}
		return
	}
	if ov.IsValid() && !nv.IsValid() {
		if path == "" {
			b.WriteString("  Value deleted\n")
		} else {
			b.WriteString(fmt.Sprintf("  %s: <set> -> nil\n", path))
		}
		return
	}

	// Handle pointers
	if ov.Kind() == reflect.Ptr || nv.Kind() == reflect.Ptr {
		// Normalize pointer states
		if ov.Kind() == reflect.Ptr && ov.IsNil() && nv.Kind() == reflect.Ptr && nv.IsNil() {
			return
		}
		if ov.Kind() == reflect.Ptr && ov.IsNil() {
			b.WriteString(fmt.Sprintf("  %s: nil -> <set>\n", nonEmpty(path)))
			return
		}
		if nv.Kind() == reflect.Ptr && nv.IsNil() {
			b.WriteString(fmt.Sprintf("  %s: <set> -> nil\n", nonEmpty(path)))
			return
		}
		// Deref if possible
		if ov.Kind() == reflect.Ptr && ov.IsNil() == false {
			ov = ov.Elem()
		}
		if nv.Kind() == reflect.Ptr && nv.IsNil() == false {
			nv = nv.Elem()
		}
		// Continue with underlying kinds
	}

	// Different types entirely
	if ov.Type() != nv.Type() {
		b.WriteString(fmt.Sprintf("  %s: type changed (%s -> %s)\n", nonEmpty(path), ov.Type(), nv.Type()))
		return
	}

	switch ov.Kind() {
	case reflect.Struct:
		t := ov.Type()
		for i := 0; i < ov.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			subPath := joinPath(path, f.Name)
			diffValue(ov.Field(i), nv.Field(i), subPath, b)
		}
	case reflect.Map:
		compareMap(ov, nv, path, b)
	case reflect.Slice, reflect.Array:
		compareSlice(ov, nv, path, b)
	default:
		// Primitives and everything else
		if !reflect.DeepEqual(ov.Interface(), nv.Interface()) {
			b.WriteString(fmt.Sprintf("  %s: %v -> %v\n", nonEmpty(path), ov.Interface(), nv.Interface()))
		}
	}
}

func compareMap(ov, nv reflect.Value, path string, b *strings.Builder) {
	// Added or modified
	for _, k := range nv.MapKeys() {
		nvVal := nv.MapIndex(k)
		ovVal := ov.MapIndex(k)
		keyPath := fmt.Sprintf("%s[%v]", nonEmpty(path), k.Interface())

		if !ovVal.IsValid() {
			b.WriteString(fmt.Sprintf("  %s: added\n", keyPath))
			continue
		}
		// If values are primitives and differ, show old -> new; else recurse for detail
		if isPrimitive(ovVal.Kind()) && isPrimitive(nvVal.Kind()) {
			if !reflect.DeepEqual(ovVal.Interface(), nvVal.Interface()) {
				b.WriteString(fmt.Sprintf("  %s: %v -> %v\n", keyPath, ovVal.Interface(), nvVal.Interface()))
			}
		} else {
			diffValue(ovVal, nvVal, keyPath, b)
		}
	}
	// Removed
	for _, k := range ov.MapKeys() {
		if !nv.MapIndex(k).IsValid() {
			keyPath := fmt.Sprintf("%s[%v]", nonEmpty(path), k.Interface())
			b.WriteString(fmt.Sprintf("  %s: removed\n", keyPath))
		}
	}
}

func compareSlice(ov, nv reflect.Value, path string, b *strings.Builder) {
	if ov.Len() != nv.Len() {
		b.WriteString(fmt.Sprintf("  %s: length changed (%d -> %d)\n", nonEmpty(path), ov.Len(), nv.Len()))
	}
	// Compare elements up to min length for detailed changes
	min := ov.Len()
	if nv.Len() < min {
		min = nv.Len()
	}
	for i := 0; i < min; i++ {
		subPath := fmt.Sprintf("%s[%d]", nonEmpty(path), i)
		diffValue(ov.Index(i), nv.Index(i), subPath, b)
	}
}

func isPrimitive(k reflect.Kind) bool {
	switch k {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func joinPath(base, add string) string {
	if base == "" {
		return add
	}
	return base + "." + add
}

func nonEmpty(path string) string {
	if path == "" {
		return "<root>"
	}
	return path
}
