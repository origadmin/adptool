// Package parser implements the functions, types, and interfaces for the module.
package parser

import (
	goast "go/ast"
	gotoken "go/token"
	"strings"
)

// DirectiveExtractor iterates over comments and extracts adptool directives.
type DirectiveExtractor struct {
	comments []*goast.Comment
	fset     *gotoken.FileSet
	index    int
}

// NewDirectiveExtractor creates a new DirectiveExtractor.
func NewDirectiveExtractor(file *goast.File, fset *gotoken.FileSet) *DirectiveExtractor {
	var comments []*goast.Comment
	for _, cg := range file.Comments {
		comments = append(comments, cg.List...)
	}
	return &DirectiveExtractor{
		comments: comments,
		fset:     fset,
		index:    0,
	}
}

// Next returns the next parsed Directive, its line number, and an error if any.
// It returns nil, 0, nil when there are no more directives.
func (de *DirectiveExtractor) Next() *Directive {
	for de.index < len(de.comments) {
		comment := de.comments[de.index]
		de.index++

		line := de.fset.Position(comment.Pos()).Line

		if !strings.HasPrefix(comment.Text, directivePrefix) {
			continue
		}

		rawDirective := strings.TrimPrefix(comment.Text, directivePrefix)
		commentStart := strings.Index(rawDirective, "//")
		if commentStart != -1 {
			rawDirective = strings.TrimSpace(rawDirective[:commentStart])
		}
		return parseDirective(rawDirective, line)
	}
	return nil // No more directives
}
