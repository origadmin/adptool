package engine

import (
	"testing"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/interfaces"
)

// testLogger implements the Logger interface for testing
type testLogger struct{}

func (m *testLogger) Info(msg string, args ...interface{}) {}

func (m *testLogger) Warn(msg string, args ...interface{}) {}

func (m *testLogger) Error(msg string, args ...interface{}) {}

// testGenerator implements the Generator interface for testing
type testGenerator struct{}

func (m *testGenerator) Generate(plan *PackagePlan) error {
	return nil
}

// testCompiler implements the Compiler interface for testing
type testCompiler struct{}

func (m *testCompiler) Compile(pkgConfig *config.Config) (*interfaces.CompiledConfig, error) {
	return &interfaces.CompiledConfig{
		PackageName: pkgConfig.PackageName,
	}, nil
}

// Helper functions for testing
func newTestLogger(t *testing.T) *testLogger {
	return &testLogger{}
}

func newTestGenerator(t *testing.T) *testGenerator {
	return &testGenerator{}
}

func newTestCompiler(t *testing.T) *testCompiler {
	return &testCompiler{}
}