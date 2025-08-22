package generator

import (
	"go/ast"
	"log/slog"
)

// isBuiltinType checks if a type name is a built-in Go type that should not be qualified.
func isBuiltinType(name string) bool {
	builtinTypes := map[string]bool{
		"any":        true,
		"bool":       true,
		"byte":       true,
		"complex64":  true,
		"complex128": true,
		"error":      true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,
	}

	return builtinTypes[name]
}

// qualifyType recursively qualifies types with the given package alias.
// It ensures that references to types from the source package use the correct alias.
func qualifyType(expr ast.Expr, pkgAlias string, definedTypes map[string]bool, typeParams map[string]bool) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		if typeParams != nil && typeParams[t.Name] {
			return t // It's a generic type parameter, don't qualify.
		}
		if definedTypes != nil && definedTypes[t.Name] {
			slog.Debug("Using local type", "func", "qualifyType", "type", t.Name)
			return t
		}

		if isBuiltinType(t.Name) {
			slog.Debug("Using built-in type", "func", "qualifyType", "type", t.Name)
			return t
		}

		// Skip empty package aliases (like for local package)
		//if pkgAlias == "" {
		//	return t
		//}

		slog.Debug("Qualifying identifier with package", "func", "qualifyType", "identifier", t.Name, "package", pkgAlias)
		return &ast.SelectorExpr{
			X:   ast.NewIdent(pkgAlias),
			Sel: t,
		}
	case *ast.StarExpr:
		slog.Debug("Processing pointer type", "func", "qualifyType")
		return &ast.StarExpr{
			X: qualifyType(t.X, pkgAlias, definedTypes, typeParams),
		}
	case *ast.ArrayType:
		slog.Debug("Processing array type", "func", "qualifyType")
		return &ast.ArrayType{
			Len: t.Len, // Array length is an expression, should not be qualified in this context
			Elt: qualifyType(t.Elt, pkgAlias, definedTypes, typeParams),
		}
	case *ast.MapType:
		slog.Debug("Processing map type", "func", "qualifyType")
		return &ast.MapType{
			Key:   qualifyType(t.Key, pkgAlias, definedTypes, typeParams),
			Value: qualifyType(t.Value, pkgAlias, definedTypes, typeParams),
		}
	case *ast.ChanType:
		slog.Debug("Processing channel type", "func", "qualifyType")
		return &ast.ChanType{
			Dir:   t.Dir,
			Value: qualifyType(t.Value, pkgAlias, definedTypes, typeParams),
		}
	case *ast.FuncType:
		slog.Debug("Processing function type", "func", "qualifyType")
		newTypeParams := make(map[string]bool)
		if typeParams != nil {
			for k, v := range typeParams {
				newTypeParams[k] = v
			}
		}
		if t.TypeParams != nil {
			for _, field := range t.TypeParams.List {
				for _, name := range field.Names {
					newTypeParams[name.Name] = true
				}
			}
		}

		if t.TypeParams != nil {
			for _, field := range t.TypeParams.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes, newTypeParams)
			}
		}
		if t.Params != nil {
			for _, field := range t.Params.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes, newTypeParams)
			}
		}
		if t.Results != nil {
			for _, field := range t.Results.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes, newTypeParams)
			}
		}
		return t
	case *ast.IndexExpr:
		slog.Debug("Processing index expression", "func", "qualifyType")
		t.X = qualifyType(t.X, pkgAlias, definedTypes, typeParams)
		t.Index = qualifyType(t.Index, pkgAlias, definedTypes, typeParams)
		return t
	case *ast.IndexListExpr:
		slog.Debug("Processing index list expression", "func", "qualifyType")
		t.X = qualifyType(t.X, pkgAlias, definedTypes, typeParams)
		for i, index := range t.Indices {
			t.Indices[i] = qualifyType(index, pkgAlias, definedTypes, typeParams)
		}
		return t
	case *ast.Ellipsis:
		slog.Debug("Processing ellipsis type", "func", "qualifyType")
		t.Elt = qualifyType(t.Elt, pkgAlias, definedTypes, typeParams)
		return t
	case *ast.InterfaceType, *ast.StructType, *ast.SelectorExpr:
		return t // These types (and selectors) are already context-complete.
	default:
		slog.Debug("Unknown type, returning as is", "func", "qualifyType", "type", t)
		return t
	}
}

// getIdentName gets the name from an identifier expression.
func getIdentName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func hasUnexportedTypes(f *ast.FuncType) bool {
	if f.Params != nil {
		for _, p := range f.Params.List {
			if containsUnexported(p.Type) {
				return true
			}
		}
	}
	if f.Results != nil {
		for _, r := range f.Results.List {
			if containsUnexported(r.Type) {
				return true
			}
		}
	}
	return false
}

func containsUnexported(expr ast.Expr) bool {
	var unexported bool
	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if !isBuiltinType(id.Name) && !ast.IsExported(id.Name) {
				unexported = true
				return false
			}
		}
		return true
	})
	return unexported
}
