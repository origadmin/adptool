package engine

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
	adpparser "github.com/origadmin/adptool/internal/parser"
)

// LoadContext holds the context for the loading phase.
type LoadContext struct {
	Files       map[string]*ast.File
	FileSets    map[string]*token.FileSet
	Config      *config.Config
	CompiledCfg *config.Config
}

// Planner is responsible for creating an execution plan.
type Planner struct {
	config   *config.Config
	logger   Logger
	compiler Compiler
	generator Generator
}

// Compiler compiles package configurations.
type Compiler interface {
	Compile(pkgConfig *config.Config) (*interfaces.CompiledConfig, error)
}

// Generator generates adapter code.
type Generator interface {
	Generate(plan *PackagePlan) error
}

// Logger interface for logging.
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// NewPlanner creates a new Planner.
func NewPlanner(cfg *config.Config, logger Logger, compiler Compiler, generator Generator) *Planner {
	return &Planner{
		config:    cfg,
		logger:    logger,
		compiler:  compiler,
		generator: generator,
	}
}

// Plan creates an execution plan based on the load context.
func (p *Planner) Plan(loadCtx *LoadContext) (*ExecutionPlan, error) {
	p.logger.Info("Creating execution plan")

	plan := &ExecutionPlan{
		Packages: make([]*PackagePlan, 0),
	}

	// Create package plans from loaded files
	for filePath, file := range loadCtx.Files {
		// Parse file directives to get package configuration
		pkgConfig, err := adpparser.ParseFileDirectives(loadCtx.Config, file, loadCtx.FileSets[filePath])
		if err != nil {
			return nil, fmt.Errorf("failed to parse directives in %s: %w", filePath, err)
		}

		// Compile the package configuration
		compiledCfg, err := p.compiler.Compile(pkgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to compile config for %s: %w", filePath, err)
		}

		// Create a package plan
		pkgPlan := &PackagePlan{
			Name:        file.Name.Name,
			ImportPath:  pkgConfig.PackageName, // Use PackageName from config
			SourceFiles: []string{filePath},
			Config: compiledCfg,
		}

		plan.Packages = append(plan.Packages, pkgPlan)
		p.logger.Info("Added package to plan", "package", pkgPlan.Name)
	}

	p.logger.Info("Created execution plan", "packages", len(plan.Packages))
	return plan, nil
}