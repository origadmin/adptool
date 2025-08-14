package parser

import (
	goast "go/ast"
	gotoken "go/token"
	"iter"
	"strings"
)

// DirectiveIterator iterates over comments and extracts adptool directives.
type DirectiveIterator struct {
	comments []*goast.Comment
	fset     *gotoken.FileSet
	index    int
}

// NewDirectiveIterator creates a new DirectiveIterator.
func NewDirectiveIterator(file *goast.File, fset *gotoken.FileSet) *DirectiveIterator {
	var comments []*goast.Comment
	for _, cg := range file.Comments {
		comments = append(comments, cg.List...)
	}
	return &DirectiveIterator{
		comments: comments,
		fset:     fset,
		index:    0,
	}
}

// Seq returns an iter.Seq that yields *Directive objects.
// This allows DirectiveIterator to be used in a for...range like pattern.
func (de *DirectiveIterator) Seq() iter.Seq[*Directive] {
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

			pd := extractDirective(rawDirective, line) // parseDirective returns Directive (value type)
			if !yield(&pd) {                           // Yield the directive and check if iteration should continue
				return
			}
		}
	}
}
