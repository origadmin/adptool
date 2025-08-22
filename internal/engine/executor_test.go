package engine

import (
	"context"
	"testing"
)

func TestExecutor_New(t *testing.T) {
	generator := newTestGenerator(t)
	logger := newTestLogger(t)
	executor := NewExecutor(generator, nil, logger)
	if executor == nil {
		t.Error("Expected executor to be created, got nil")
	}
}

func TestExecutor_Execute(t *testing.T) {
	logger := newTestLogger(t)
	generator := newTestGenerator(t)
	executor := NewExecutor(generator, nil, logger)
	
	ctx := context.Background()
	plan := &ExecutionPlan{
		Packages: make([]*PackagePlan, 0),
	}

	err := executor.Execute(ctx, plan)
	if err != nil {
		t.Errorf("Expected Execute to succeed with empty plan, got error: %v", err)
	}
}

