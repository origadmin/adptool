package generator

import (
	"github.com/origadmin/adptool/internal/interfaces"
)

// Generator holds the state and configuration for code generation.
type Generator struct {
	collector    *Collector
	builder      *Builder
	noEditHeader bool // 控制是否添加"do not edit"头部注释
}

// NewGenerator creates a new Generator instance.
func NewGenerator(packageName string, outputFilePath string, replacer interfaces.Replacer) *Generator {
	return &Generator{
		collector: NewCollector(replacer),
		builder:   NewBuilder(packageName, outputFilePath, true), // 默认添加"do not edit"头部注释
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

// WithFormatCode 设置是否在生成代码后自动格式化
func (g *Generator) WithFormatCode(format bool) *Generator {
	g.builder.WithFormatCode(format)
	return g
}

// WithNoEditHeader 设置是否添加"do not edit"头部注释
func (g *Generator) WithNoEditHeader(noEditHeader bool) *Generator {
	g.noEditHeader = noEditHeader
	g.builder.noEditHeader = noEditHeader
	return g
}
