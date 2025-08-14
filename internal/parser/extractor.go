package parser

import (
	goast "go/ast"
	gotoken "go/token"
	"iter"
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

// Seq returns an iter.Seq that yields *Directive objects.
// This allows DirectiveExtractor to be used in a for...range like pattern.
func (de *DirectiveExtractor) Seq() iter.Seq[*Directive] {
	return func(yield func(*Directive) bool) {
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

			pd := parseDirective(rawDirective, line) // parseDirective returns Directive (value type)
			if !yield(&pd) {                         // Yield the directive and check if iteration should continue
				return
			}
		}
	}
}

// Next is no longer needed for the Seq pattern, but kept for backward compatibility if needed.
// It returns the next parsed Directive and a boolean indicating if there are more directives.
// func (de *DirectiveExtractor) Next() (*Directive, bool) {
// 	for de.index < len(de.comments) {
// 		comment := de.comments[de.index]
// 		de.index++

// 		line := de.fset.Position(comment.Pos()).Line

// 		if !strings.HasPrefix(comment.Text, directivePrefix) {
// 			continue
// 		}

// 		rawDirective := strings.TrimPrefix(comment.Text, directivePrefix)
// 		commentStart := strings.Index(rawDirective, "//")
// 		if commentStart != -1 {
// 			rawDirective = strings.TrimSpace(rawDirective[:commentStart])
// 		}

// 		pd := parseDirective(rawDirective) // parseDirective now returns Directive (value type)
// 		pd.Line = line                     // Assign line to Directive
// 		return &pd, true                   // Return pointer and true for more directives
// 	}
// 	return nil, false // No more directives
// }
