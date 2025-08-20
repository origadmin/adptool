package generator

import (
	"go/ast"
	"log"
)

// isBuiltinType checks if a type name is a built-in Go type that should not be qualified.
func isBuiltinType(name string) bool {
	builtinTypes := map[string]bool{
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
func qualifyType(expr ast.Expr, pkgAlias string, definedTypes map[string]bool) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		if definedTypes != nil && definedTypes[t.Name] {
			log.Printf("qualifyType: Using local type %s", t.Name)
			return t
		}

		if isBuiltinType(t.Name) {
			log.Printf("qualifyType: Using built-in type %s", t.Name)
			return t
		}

		log.Printf("qualifyType: Qualifying identifier %s with package %s", t.Name, pkgAlias)
		return &ast.SelectorExpr{
			X:   ast.NewIdent(pkgAlias),
			Sel: t,
		}
	case *ast.StarExpr:
		log.Printf("qualifyType: Processing pointer type")
		return &ast.StarExpr{
			X: qualifyType(t.X, pkgAlias, definedTypes),
		}
	case *ast.ArrayType:
		log.Printf("qualifyType: Processing array type")
		return &ast.ArrayType{
			Len: t.Len, // Array length is an expression, should not be qualified in this context
			Elt: qualifyType(t.Elt, pkgAlias, definedTypes),
		}
	case *ast.MapType:
		log.Printf("qualifyType: Processing map type")
		return &ast.MapType{
			Key:   qualifyType(t.Key, pkgAlias, definedTypes),
			Value: qualifyType(t.Value, pkgAlias, definedTypes),
		}
	case *ast.ChanType:
		log.Printf("qualifyType: Processing channel type")
		return &ast.ChanType{
			Dir:   t.Dir,
			Value: qualifyType(t.Value, pkgAlias, definedTypes),
		}
	case *ast.FuncType:
		log.Printf("qualifyType: Processing function type")
		if t.Params != nil {
			for _, field := range t.Params.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		if t.Results != nil {
			for _, field := range t.Results.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		return t
	case *ast.InterfaceType, *ast.StructType, *ast.SelectorExpr:
		return t // These types (and selectors) are already context-complete.
	default:
		log.Printf("qualifyType: Unknown type %T, returning as is", t)
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
