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
func NewGenerator(packageName string, outputFilePath string, replacer interfaces.Replacer) *Generator {
	return &Generator{
		collector: NewCollector(replacer),
		builder:   NewBuilder(packageName, outputFilePath),
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

// ApplyRules applies a set of rename rules to a given name and returns the result.
// This is a wrapper around rules.ApplyRules for backward compatibility.
func ApplyRules(name string, rulesList []interfaces.RenameRule) (string, error) {
	// Placeholder - actual implementation would depend on the rules package
	return name, nil
}
