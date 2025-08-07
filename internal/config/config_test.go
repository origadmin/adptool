package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestConfigFile creates a temporary config file for testing.
func setupTestConfigFile(t *testing.T, dir, filename string, content []byte) string {
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, content, 0644)
	assert.NoError(t, err)
	return path
}

func TestLoadConfig(t *testing.T) {
	t.Run("FullConfigNormalizationAndPriority", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := []byte(`
explicit:
  OldName: NewName
prefix: "Global"
suffix: "Adapter"
rename_rules:
  - type: regex
    pattern: "Service$"
    replace: "ServiceV2"
packages:
  "test.com/pkg1":
    alias: "p1"
    prefix: "Pkg1"
    explicit:
      OldFunc: NewFunc
    rename_rules:
      - type: suffix
        value: "Wrapper"
`)
		setupTestConfigFile(t, tmpDir, "adptool.yaml", configContent)

		wd, _ := os.Getwd()
		defer os.Chdir(wd)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Test top-level normalization and priority
		assert.Len(t, cfg.RenameRules, 4)
		assert.Equal(t, "explicit", cfg.RenameRules[0].Type)
		assert.Equal(t, "OldName", cfg.RenameRules[0].From)
		assert.Equal(t, "NewName", cfg.RenameRules[0].To)
		assert.Equal(t, "prefix", cfg.RenameRules[1].Type)
		assert.Equal(t, "Global", cfg.RenameRules[1].Value)
		assert.Equal(t, "suffix", cfg.RenameRules[2].Type)
		assert.Equal(t, "Adapter", cfg.RenameRules[2].Value)
		assert.Equal(t, "regex", cfg.RenameRules[3].Type)

		// Test package-level normalization and priority
		pkg1, ok := cfg.Packages["test.com/pkg1"]
		assert.True(t, ok)
		assert.Equal(t, "p1", pkg1.Alias)
		assert.Len(t, pkg1.RenameRules, 3)
		assert.Equal(t, "explicit", pkg1.RenameRules[0].Type)
		assert.Equal(t, "OldFunc", pkg1.RenameRules[0].From)
		assert.Equal(t, "NewFunc", pkg1.RenameRules[0].To)
		assert.Equal(t, "prefix", pkg1.RenameRules[1].Type)
		assert.Equal(t, "Pkg1", pkg1.RenameRules[1].Value)
		assert.Equal(t, "suffix", pkg1.RenameRules[2].Type)
		assert.Equal(t, "Wrapper", pkg1.RenameRules[2].Value)
	})

	t.Run("EmptyConfigWhenNoFileExists", func(t *testing.T) {
		tmpDir := t.TempDir()
		wd, _ := os.Getwd()
		defer os.Chdir(wd)
		os.Chdir(tmpDir)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.RenameRules)
		assert.Empty(t, cfg.Packages)
	})

	t.Run("FileLevelConfigOverridesDefault", func(t *testing.T) {
		tmpDir := t.TempDir()
		setupTestConfigFile(t, tmpDir, "adptool.yaml", []byte(`prefix: "Ignored"`))

		fileConfigContent := []byte(`prefix: "FileLevel"`)
		fileConfigPath := setupTestConfigFile(t, tmpDir, "file.yaml", fileConfigContent)

		cfg, err := LoadConfig(fileConfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "FileLevel", cfg.Prefix)
		assert.Len(t, cfg.RenameRules, 1)
		assert.Equal(t, "prefix", cfg.RenameRules[0].Type)
	})

	t.Run("ErrorOnInvalidYAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		fileConfigPath := setupTestConfigFile(t, tmpDir, "invalid.yaml", []byte(`prefix: [invalid`))

		_, err := LoadConfig(fileConfigPath)
		assert.Error(t, err)
	})

	t.Run("ErrorOnSpecifiedFileNotFound", func(t *testing.T) {
		_, err := LoadConfig("nonexistent.yaml")
		assert.Error(t, err)
	})
}
