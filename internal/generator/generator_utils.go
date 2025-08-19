package generator

import (
	"go/ast"

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

// qualifyType recursively qualifies ast.Ident nodes within an ast.Expr
// with the given package alias if they are not already qualified.
// definedTypes is a map of type names that are defined in the current package.
func qualifyType(expr ast.Expr, pkgAlias string, definedTypes map[string]bool) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		// Only qualify if it's not already qualified (i.e., not a SelectorExpr),
		// it's an exported identifier, and it's not defined in the current package.
		if t.IsExported() && !definedTypes[t.Name] {
			return &ast.SelectorExpr{
				X:   ast.NewIdent(pkgAlias),
				Sel: t,
			}
		}
		return t
	case *ast.StarExpr:
		t.X = qualifyType(t.X, pkgAlias, definedTypes)
		return t
	case *ast.ArrayType:
		t.Elt = qualifyType(t.Elt, pkgAlias, definedTypes)
		return t
	case *ast.MapType:
		t.Key = qualifyType(t.Key, pkgAlias, definedTypes)
		t.Value = qualifyType(t.Value, pkgAlias, definedTypes)
		return t
	case *ast.ChanType:
		t.Value = qualifyType(t.Value, pkgAlias, definedTypes)
		return t
	case *ast.FuncType:
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
	case *ast.InterfaceType:
		// Interface methods' types need to be qualified
		if t.Methods != nil {
			for _, field := range t.Methods.List {
				if funcType, ok := field.Type.(*ast.FuncType); ok {
					field.Type = qualifyType(funcType, pkgAlias, definedTypes)
				}
			}
		}
		return t
	case *ast.StructType:
		// Struct fields' types need to be qualified
		if t.Fields != nil {
			for _, field := range t.Fields.List {
				field.Type = qualifyType(field.Type, pkgAlias, definedTypes)
			}
		}
		return t
	case *ast.SelectorExpr:
		// If it's already a SelectorExpr, check if it's a reference to a type from the source package
		// that we've defined locally
		if ident, ok := t.X.(*ast.Ident); ok && ident.Name == pkgAlias {
			typeName := t.Sel.Name
			if definedTypes[typeName] {
				// Replace with local alias
				return ast.NewIdent(typeName)
			}
		}
		return t
	default:
		return expr
	}
}
