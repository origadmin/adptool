package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// runInTempDir sets up a temporary directory for a test, changes the working
// directory to it, and handles cleanup. It's a helper to avoid boilerplate
// code in multiple test cases.
func runInTempDir(t *testing.T, testFunc func(tmpDir string)) {
	// Save the original working directory.
	originalWd, err := os.Getwd()
	assert.NoError(t, err, "Failed to get current working directory")

	// Create a new temporary directory.
	tmpDir := t.TempDir()

	// Change the working directory to the temporary one.
	err = os.Chdir(tmpDir)
	assert.NoError(t, err, "Failed to change to temporary directory")

	// Defer changing back to the original directory.
	defer func() {
		err := os.Chdir(originalWd)
		assert.NoError(t, err, "Failed to restore original working directory")
	}()

	// Run the actual test function.
	testFunc(tmpDir)
}

// setupTestConfigFile creates a temporary config file for testing.
func setupTestConfigFile(t *testing.T, dir, filename string, content []byte) string {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, content, 0644)
	assert.NoError(t, err)
	return path
}
