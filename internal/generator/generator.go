package generator

import (
	"github.com/origadmin/adptool/internal/interfaces"
)

// Generator holds the state and configuration for code generation.
type Generator struct {
	collector *Collector
	builder   *Builder
}

// NewGenerator creates a new Generator instance.
func NewGenerator(packageName string, outputFilePath string, replacer interfaces.Replacer, copyrightHolder string) *Generator {
	return &Generator{
		collector: NewCollector(replacer),
		builder:   NewBuilder(packageName, outputFilePath, copyrightHolder),
	}
}

// RenderHeader renders the header for the generated file.
func (g *Generator) RenderHeader(sourceFile string) error {
	return g.builder.RenderHeader(sourceFile)
}

// Generate generates the output code.
func (g *Generator) Generate(packages []*PackageInfo) error {
	if err := g.collector.Collect(packages); err != nil {
		return err
	}

	// Pass the pathToAlias map from the collector to the builder.
	g.builder.Build(g.collector.importSpecs, g.collector.allPackageDecls, g.collector.definedTypes, g.collector.pathToAlias)

	return g.builder.Write()
}

// WithFormatCode sets whether to automatically format after generating code
func (g *Generator) WithFormatCode(format bool) *Generator {
	g.builder.WithFormatCode(format)
	return g
}
