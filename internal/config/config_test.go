package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Test Case 1: Load default adptool.yaml
	t.Run("LoadDefaultConfig", func(t *testing.T) {
		// Create a dummy adptool.yaml
		defaultConfigContent := []byte(`
global_prefix: "TEST_GLOBAL"
packages:
  "test.com/pkg1":
    alias: "tp1"
    global_prefix: "PKG1_GLOBAL"
`)
		defaultConfigPath := filepath.Join(tmpDir, "adptool.yaml")
		err := os.WriteFile(defaultConfigPath, defaultConfigContent, 0644)
		assert.NoError(t, err)

		// Change working directory to tmpDir for this test
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "TEST_GLOBAL", cfg.GlobalPrefix)
		assert.Contains(t, cfg.Packages, "test.com/pkg1")
		assert.Equal(t, "PKG1_GLOBAL", cfg.Packages["test.com/pkg1"].GlobalPrefix)
	})

	// Test Case 2: adptool.yaml does not exist
	t.Run("DefaultConfigNotFound", func(t *testing.T) {
		// Ensure adptool.yaml does not exist in a new temp dir
		newTmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(newTmpDir)

		cfg, err := LoadConfig("")
		assert.NoError(t, err) // Should not return error if file not found
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.GlobalPrefix)
		assert.Empty(t, cfg.Packages)
	})

	// Test Case 3: Load file-level config, which replaces default
	t.Run("LoadFileLevelConfig", func(t *testing.T) {
		// Create a dummy adptool.yaml (should be ignored)
		defaultConfigContent := []byte(`global_prefix: "SHOULD_BE_IGNORED"`)
		defaultConfigPath := filepath.Join(tmpDir, "adptool.yaml")
		os.WriteFile(defaultConfigPath, defaultConfigContent, 0644)

		// Create a file-level config
		fileConfigContent := []byte(`global_prefix: "FILE_LEVEL_GLOBAL"`)
		fileConfigPath := filepath.Join(tmpDir, "file_config.yaml")
		err := os.WriteFile(fileConfigPath, fileConfigContent, 0644)
		assert.NoError(t, err)

		// Change working directory to tmpDir for this test
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig(fileConfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "FILE_LEVEL_GLOBAL", cfg.GlobalPrefix)
	})

	// Test Case 4: File-level config does not exist
	t.Run("FileLevelConfigNotFound", func(t *testing.T) {
		// Ensure adptool.yaml exists for this test
		defaultConfigContent := []byte(`global_prefix: "DEFAULT_EXISTS"`)
		defaultConfigPath := filepath.Join(tmpDir, "adptool.yaml")
		os.WriteFile(defaultConfigPath, defaultConfigContent, 0644)

		// Try to load a non-existent file-level config
		fileConfigPath := filepath.Join(tmpDir, "non_existent_file.yaml")

		// Change working directory to tmpDir for this test
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig(fileConfigPath)
		assert.Error(t, err) // Should return error if file-level config not found
		assert.Nil(t, cfg)
	})

	// Test Case 5: Invalid config content
	t.Run("InvalidConfigContent", func(t *testing.T) {
		invalidConfigContent := []byte(`global_prefix: [invalid`) // Invalid YAML
		invalidConfigPath := filepath.Join(tmpDir, "invalid_config.yaml")
		err := os.WriteFile(invalidConfigPath, invalidConfigContent, 0644)
		assert.NoError(t, err)

		// Change working directory to tmpDir for this test
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig(invalidConfigPath)
		assert.Error(t, err) // Should return error for invalid content
		assert.Nil(t, cfg)
	})
}
