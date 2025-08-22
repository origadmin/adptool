package engine

import (
	"context"
	"fmt"
)

// Executor executes the execution plan.
type Executor struct {
	generator Generator
	logger    Logger
}

// NewExecutor creates a new Executor.
func NewExecutor(generator Generator, compiler Compiler, logger Logger) *Executor {
	return &Executor{
		generator: generator,
		logger:    logger,
	}
}

// Execute executes the execution plan.
func (e *Executor) Execute(ctx context.Context, plan *ExecutionPlan) error {
	e.logger.Info("Executing plan", "packages", len(plan.Packages))

	for _, pkgPlan := range plan.Packages {
		e.logger.Info("Generating adapter for package", "package", pkgPlan.Name)

		if err := e.generator.Generate(pkgPlan); err != nil {
			return fmt.Errorf("failed to generate adapter for package %s: %w", pkgPlan.Name, err)
		}

		e.logger.Info("Generated adapter for package", "package", pkgPlan.Name)
	}

	e.logger.Info("Executed plan successfully")
	return nil
}