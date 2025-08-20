package interfaces

import (
	"go/ast"
)

// Replacer defines the interface for applying code transformations based on compiled rules.
// It takes an AST node and returns a potentially modified node.
type Replacer interface {
	Apply(ctx Context, node ast.Node) ast.Node
}


