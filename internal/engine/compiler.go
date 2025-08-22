package engine

import (
	"fmt"
	"path"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
)

// RealCompiler is a real implementation of the Compiler interface
type RealCompiler struct{}

// NewRealCompiler creates a new RealCompiler
func NewRealCompiler() *RealCompiler {
	return &RealCompiler{}
}

// Compile compiles package configurations
func (r *RealCompiler) Compile(pkgConfig *config.Config) (*interfaces.CompiledConfig, error) {
	if pkgConfig == nil {
		return nil, fmt.Errorf("package config cannot be nil")
	}

	// Compile the configuration using the real compiler
	compiledCfg, err := compiler.Compile(pkgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to compile configuration: %w", err)
	}

	// Ensure we have a valid package name
	if compiledCfg.PackageName == "" {
		compiledCfg.PackageName = path.Base(pkgConfig.PackageName)
	}

	return compiledCfg, nil
}