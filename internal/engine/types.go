package engine

import (
	"github.com/origadmin/adptool/internal/interfaces"
)

// ExecutionPlan represents the plan for executing the adapter generation.
type ExecutionPlan struct {
	Packages []*PackagePlan
}

// PackagePlan represents the plan for a single package.
type PackagePlan struct {
	Name        string
	ImportPath  string
	SourceFiles []string
	TargetFiles []string
	Config      *interfaces.CompiledPackage
}