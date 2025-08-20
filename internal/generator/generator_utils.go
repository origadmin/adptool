package generator

import (
	"go/ast"
	"log"

	"github.com/origadmin/adptool/internal/interfaces"
)

// replacerVisitor implements the ast.Visitor interface to apply replacements.
type replacerVisitor struct {
	replacer interfaces.Replacer
}

// Visit implements the ast.Visitor interface.
func (v replacerVisitor) Visit(n ast.Node) ast.Visitor {
	if n != nil {
		v.replacer.Apply(n)
	}
	return v // Continue traversal
}

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
		// Check if this identifier refers to a type we've defined
		if definedTypes != nil && definedTypes[t.Name] {
			// Use the local type name directly
			log.Printf("qualifyType: Using local type %s", t.Name)
			return t
		}

		// Check if this is a built-in type that should not be qualified
		if isBuiltinType(t.Name) {
			log.Printf("qualifyType: Using built-in type %s", t.Name)
			return t
		}

		// For identifiers that are not our own defined types, use selector expression
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
			Len: qualifyType(t.Len, pkgAlias, definedTypes),
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
		// Process function parameters
		var newParams []*ast.Field
		if t.Params != nil {
			for _, field := range t.Params.List {
				newField := &ast.Field{
					Names: field.Names,
					Type:  qualifyType(field.Type, pkgAlias, definedTypes),
				}
				newParams = append(newParams, newField)
			}
		}

		// Process function results
		var newResults []*ast.Field
		if t.Results != nil {
			for _, field := range t.Results.List {
				newField := &ast.Field{
					Names: field.Names,
					Type:  qualifyType(field.Type, pkgAlias, definedTypes),
				}
				newResults = append(newResults, newField)
			}
		}

		return &ast.FuncType{
			Params:  &ast.FieldList{List: newParams},
			Results: &ast.FieldList{List: newResults},
		}
	case *ast.InterfaceType:
		// For interface types, we generally don't need to qualify methods
		// as they are part of the interface definition
		log.Printf("qualifyType: Processing interface type")
		return t
	case *ast.StructType:
		// For struct types, we generally don't need to qualify fields
		// as they are part of the struct definition
		log.Printf("qualifyType: Processing struct type")
		return t
	case *ast.SelectorExpr:
		// Handle selector expressions (e.g., pkg.Type)
		// If the selector is a defined type, use it directly
		log.Printf("qualifyType: Processing selector expression %s.%s",
			getIdentName(t.X), getIdentName(t.Sel))
		// Otherwise, return as is
		log.Printf("qualifyType: Keeping selector expression as is")
		return t
	default:
		// For all other types, return as is
		log.Printf("qualifyType: Unknown type %T, returning as is", t)
		return t
	}
}

// getIdentName 获取标识符名称的辅助函数
func getIdentName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
