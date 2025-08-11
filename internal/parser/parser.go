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
	Name       string       // Name of the declaration (e.g., "MyType", "MyFunc")
	Kind       string       // Type of declaration (e.g., "type", "func", "var", "const", "method", "field")
	Directives []*Directive // The parsed adptool directives, if any
	// Add more fields as needed, e.g., receiver for methods, parent type for fields
}

// Directive represents a parsed //go:adapter: directive.
// This is a simplified version for initial parsing. Inline rule parsing will be added later.
type Directive struct {
	Command  string // e.g., "package", "type", "func"
	Argument string // e.g., "github.com/some/pkg", "MyType"

	// Fields for inline rules (will be populated in a later stage of parsing)
	RulePath   string // e.g., "prefix", "methods:explicit"
	TargetName string // For inline rules, the name of the target (e.g., "MyType")
	Value      string // For inline rules, the raw value string
	IsJSON     bool   // True if the inline rule value is JSON
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

	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl: // import, const, type, var declarations
			// Check for directives on the GenDecl itself
			if decl.Doc != nil {
				var directives []*Directive
				for _, comment := range decl.Doc.List {
					if strings.HasPrefix(comment.Text, "//go:adapter:") {
						dir, err := parseDirective(comment.Text)
						if err != nil {
							// Log error or return, for now just skip
							continue
						}
						directives = append(directives, dir)
					}
				}
				if len(directives) > 0 {
					parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
						Name:       "", // Name is ambiguous for GenDecl block, will be refined
						Kind:       "GenDecl",
						Directives: directives,
					})
				}
			}

			// Check for directives on individual Specs within the GenDecl
			for _, spec := range decl.Specs {
				var directives []*Directive
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Doc != nil {
						for _, comment := range s.Doc.List {
							if strings.HasPrefix(comment.Text, "//go:adapter:") {
								dir, err := parseDirective(comment.Text)
								if err != nil {
									continue
								}
								directives = append(directives, dir)
							}
						}
					}
					if len(directives) > 0 {
						parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
							Name:       s.Name.Name,
							Kind:       "type",
							Directives: directives,
						})
					}
				case *ast.ValueSpec: // var or const
					if s.Doc != nil {
						for _, comment := range s.Doc.List {
							if strings.HasPrefix(comment.Text, "//go:adapter:") {
								dir, err := parseDirective(comment.Text)
								if err != nil {
									continue
								}
								directives = append(directives, dir)
							}
						}
					}
					if len(directives) > 0 {
						kind := "var"
						if decl.Tok == token.CONST {
							kind = "const"
						}
						parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
							Name:       s.Names[0].Name, // Take the first name
							Kind:       kind,
							Directives: directives,
						})
					}
				}
			}
		case *ast.FuncDecl: // function or method declarations
			var directives []*Directive
			if decl.Doc != nil {
				for _, comment := range decl.Doc.List {
					if strings.HasPrefix(comment.Text, "//go:adapter:") {
						dir, err := parseDirective(comment.Text)
						if err != nil {
							continue
						}
						directives = append(directives, dir)
					}
				}
			}
			if len(directives) > 0 {
				kind := "func"
				if decl.Recv != nil { // It's a method
					kind = "method"
				}
				parsedFile.Declarations = append(parsedFile.Declarations, &ParsedDeclaration{
					Name:       decl.Name.Name,
					Kind:       kind,
					Directives: directives,
				})
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

	// Handle inline rules: type:prefix MyType My_
	// Or type:explicit:json MyType '[{"from": "Foo", "to": "Bar"}]'
	inlineRuleParts := strings.SplitN(command, ":", 2)
	if len(inlineRuleParts) > 1 {
		// This is an inline rule directive
		cmd := inlineRuleParts[0]             // e.g., "type"
		rulePathAndJSON := inlineRuleParts[1] // e.g., "prefix" or "explicit:json"

		rulePathParts := strings.SplitN(rulePathAndJSON, ":json", 2)
		rulePath := rulePathParts[0]
		isJSON := len(rulePathParts) > 1

		// The argument now contains targetName and value
		argParts := strings.SplitN(argument, " ", 2)
		targetName := argParts[0]
		value := ""
		if len(argParts) > 1 {
			value = argParts[1]
		}

		return &Directive{
			Command:    cmd,
			Argument:   argument, // Keep original argument for full context
			RulePath:   rulePath,
			TargetName: targetName,
			Value:      value,
			IsJSON:     isJSON,
		}, nil
	}

	// Standard directive (e.g., package, type, func)
	return &Directive{
		Command:  command,
		Argument: argument,
	}, nil
}
