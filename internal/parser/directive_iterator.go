package parser

import (
	goast "go/ast"
	gotoken "go/token"
	"iter"
	"strings"
)

type DirectiveIterator iter.Seq[*Directive]

// directiveIterator iterates over comments and extracts adptool directives.
type directiveIterator struct {
	comments []*goast.Comment
	fset     *gotoken.FileSet
	index    int
}

// NewDirectiveIterator creates a new directiveIterator.
func NewDirectiveIterator(file *goast.File, fset *gotoken.FileSet) DirectiveIterator {
	var comments []*goast.Comment
	for _, cg := range file.Comments {
		comments = append(comments, cg.List...)
	}
	di := &directiveIterator{
		comments: comments,
		fset:     fset,
		index:    0,
	}
	return di.Seq()
}

// Seq returns an iter.Seq that yields *Directive objects.
// This allows directiveIterator to be used in a for...range like pattern.
func (de *directiveIterator) Seq() DirectiveIterator {
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
