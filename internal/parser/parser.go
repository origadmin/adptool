package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	// "github.com/origadmin/adptool/internal/config" // Not needed for basic parsing
)

// ParsedFile represents a Go source file that has been parsed for adptool directives.
type ParsedFile struct {
	FilePath     string
	PackageName  string
	Declarations []*ParsedDeclaration
}

// ParsedDeclaration represents a Go declaration (type, func, var, const) with its associated adptool directive.
type ParsedDeclaration struct {
	Name      string     // Name of the declaration (e.g., "MyType", "MyFunc")
	Kind      string     // Type of declaration (e.g., "type", "func", "var", "const", "method", "field")
	Directive *Directive // The parsed adptool directive, if any
	// Add more fields as needed, e.g., receiver for methods, parent type for fields
}

// Directive represents a parsed //go:adapter: directive.
// This is a simplified version for initial parsing. Inline rule parsing will be added later.
type Directive struct {
	Command  string // e.g., "package", "type", "func"
	Argument string // e.g., "github.com/some/pkg", "MyType"
}

// ParseFile parses a Go source file and extracts adptool directives.
func ParseFile(filePath string) (*ParsedFile, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	parsedFile := &ParsedFile{
		FilePath:     filePath,
		PackageName:  node.Name.Name,
		Declarations: make([]*ParsedDeclaration, 0),
	}

	// Create a comment map for easier lookup of comments associated with declarations.
	// commentMap := ast.NewCommentMap(fset, node, node.Comments) // Not used directly in this simplified version

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl: // import, const, type, var declarations
			// Check for directives on the GenDecl itself
			if decl.Doc != nil {
				for _, comment := range decl.Doc.List {
					if strings.HasPrefix(comment.Text, "//go:adapter:") {
						dir, err := parseDirective(comment.Text)
						if err != nil {
							// Log error or return, for now just skip
							continue
						}
						// For GenDecl, the directive might apply to the whole block or individual specs
						// This needs more sophisticated handling based on command (e.g., package directive)
						// For now, let's just add a placeholder declaration.
						parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
							Name:      "", // Name is ambiguous for GenDecl block
							Kind:      "GenDecl",
							Directive: dir,
						})
					}
				}
			}

			// Check for directives on individual Specs within the GenDecl
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// Directives on TypeSpec itself
					if s.Doc != nil {
						for _, comment := range s.Doc.List {
							if strings.HasPrefix(comment.Text, "//go:adapter:") {
								dir, err := parseDirective(comment.Text)
								if err != nil {
									continue
								}
								parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
									Name:      s.Name.Name,
									Kind:      "type",
									Directive: dir,
								})
							}
						}
					}
				case *ast.ValueSpec: // var or const
					// Directives on ValueSpec itself
					// Note: ValueSpec can declare multiple names (e.g., var a, b int)
					// For simplicity, we'll associate the directive with the first name.
					if s.Doc != nil {
						for _, comment := range s.Doc.List {
							if strings.HasPrefix(comment.Text, "//go:adapter:") {
								dir, err := parseDirective(comment.Text)
								if err != nil {
									continue
								}
								kind := "var"
								if decl.Tok == token.CONST { // Corrected syntax
									kind = "const"
								}
								parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
									Name:      s.Names[0].Name, // Take the first name
									Kind:      kind,
									Directive: dir,
								})
							}
						}
					}
				}
			}
		case *ast.FuncDecl: // function or method declarations
			if decl.Doc != nil {
				for _, comment := range decl.Doc.List {
					if strings.HasPrefix(comment.Text, "//go:adapter:") {
						dir, err := parseDirective(comment.Text)
						if err != nil {
							continue
						}
						kind := "func"
						if decl.Recv != nil { // It's a method
							kind = "method"
						}
						parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
							Name:      decl.Name.Name,
							Kind:      kind,
							Directive: dir,
						})
					}
				}
			}
		}
		return true
	})

	return parsedFile, nil
}

// parseDirective parses a single //go:adapter: comment line into a Directive struct.
// This is a simplified parser and needs to be expanded to handle all directive types and inline rules.
func parseDirective(commentText string) (*Directive, error) {
	// Remove //go:adapter: prefix
	text := strings.TrimPrefix(commentText, "//go:adapter:")

	parts := strings.SplitN(text, " ", 2)
	command := parts[0]
	argument := ""
	if len(parts) > 1 {
		argument = parts[1]
	}

	// For now, we only parse the command and argument. Inline rule parsing will be added later.
	return &Directive{
		Command:  command,
		Argument: argument,
	}, nil
}
