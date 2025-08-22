package engine

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/origadmin/adptool/internal/config"
)

func TestPlanner_New(t *testing.T) {
	planner := NewPlanner(nil, nil, nil, nil)
	if planner == nil {
		t.Error("Expected planner to be created, got nil")
	}
}

func TestPlanner_Plan(t *testing.T) {
	logger := newTestLogger(t)
	compiler := newTestCompiler(t)
	generator := newTestGenerator(t)
	
	planner := NewPlanner(
		config.New(),
		logger,
		compiler,
		generator,
	)
	
	loadCtx := &LoadContext{
		Files:    make(map[string]*ast.File),
		FileSets: make(map[string]*token.FileSet),
		Config:   config.New(),
	}

	plan, err := planner.Plan(loadCtx)
	if err != nil {
		t.Errorf("Expected Plan to succeed, got error: %v", err)
	}

	if plan == nil {
		t.Error("Expected plan to be returned, got nil")
	}

	// Currently the plan is empty, but in a full implementation it would contain packages
	if plan.Packages == nil {
		t.Error("Expected Packages slice to be initialized")
	}
}

