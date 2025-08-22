package generator

import (
	"github.com/origadmin/adptool/internal/interfaces"
)

// Generator holds the state and configuration for code generation.
type Generator struct {
	collector    *Collector
	builder      *Builder
	noEditHeader bool // Controls whether to add "do not edit" header comment
}

// NewGenerator creates a new Generator instance.
func NewGenerator(packageName string, outputFilePath string, replacer interfaces.Replacer) *Generator {
	return &Generator{
		collector: NewCollector(replacer),
		builder:   NewBuilder(packageName, outputFilePath, true), // Add "do not edit" header comment by default
	}
}

// Generate generates the output code.
func (g *Generator) Generate(packages []*PackageInfo) error {
	if err := g.collector.Collect(packages); err != nil {
		return err
	}

	g.builder.Build(g.collector.importSpecs, g.collector.allPackageDecls, g.collector.definedTypes)

	return g.builder.Write()
}

// WithFormatCode sets whether to automatically format after generating code
func (g *Generator) WithFormatCode(format bool) *Generator {
	g.builder.WithFormatCode(format)
	return g
}

// WithNoEditHeader sets whether to add "do not edit" header comment
func (g *Generator) WithNoEditHeader(noEditHeader bool) *Generator {
	g.noEditHeader = noEditHeader
	g.builder.noEditHeader = noEditHeader
	return g
}