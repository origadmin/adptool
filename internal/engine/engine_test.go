package engine

import (
	"context"
	"testing"

	"github.com/origadmin/adptool/internal/config"
)

func TestEngine_New(t *testing.T) {
	engine := New()
	if engine == nil {
		t.Error("Expected engine to be created, got nil")
	}
}

func TestEngine_Execute(t *testing.T) {
	engine := New()
	ctx := context.Background()
	cfg := &Config{}

	_, err := engine.Execute(ctx, cfg)
	if err != nil {
		t.Errorf("Expected Execute to succeed, got error: %v", err)
	}
}

func TestEngine_ExecuteFile(t *testing.T) {
	engine := New()
	cfg := config.New()

	// Try to execute with a non-existent file - this should return an error
	err := engine.ExecuteFile("non-existent.go", cfg)
	if err == nil {
		t.Error("Expected ExecuteFile to return error for non-existent file")
	}
}