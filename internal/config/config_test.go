package config

import (
	"os"
	"path/filepath"
	"reflect"
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
	// Save current working directory and restore it after tests
	// Removed: wd, _ := os.Getwd()
	// Removed: defer os.Chdir(wd)

	t.Run("DefaultConfigWhenNoFileExists", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Verify top-level defaults
		assert.Equal(t, "", cfg.Prefix)
		assert.Equal(t, "", cfg.Suffix)
		assert.Empty(t, cfg.Explicit)
		assert.Empty(t, cfg.Regex)
		assert.Empty(t, cfg.Ignore)
		assert.False(t, cfg.InheritPrefix)
		assert.False(t, cfg.InheritSuffix)
		assert.False(t, cfg.InheritExplicit)
		assert.False(t, cfg.InheritRegex)
		assert.False(t, cfg.InheritIgnore)

		// Verify category-specific defaults
		assert.Equal(t, "", cfg.Types.Prefix)
		assert.Equal(t, "", cfg.Types.Suffix)
		assert.Empty(t, cfg.Types.Explicit)
		assert.Empty(t, cfg.Types.Regex)
		assert.Empty(t, cfg.Types.Ignore)
		assert.False(t, cfg.Types.InheritPrefix)
		assert.False(t, cfg.Types.InheritSuffix)
		assert.False(t, cfg.Types.InheritExplicit)
		assert.False(t, cfg.Types.InheritRegex)
		assert.False(t, cfg.Types.InheritIgnore)

		// Verify compiled rules are empty
		assert.Empty(t, cfg.CompiledTypes.Rules)
		assert.Empty(t, cfg.CompiledTypes.Ignore)
		assert.Empty(t, cfg.CompiledFunctions.Rules)
		assert.Empty(t, cfg.CompiledFunctions.Ignore)
		assert.Empty(t, cfg.CompiledMethods.Rules)
		assert.Empty(t, cfg.CompiledMethods.Ignore)
		assert.Empty(t, cfg.Packages)
	})

	t.Run("TopLevelGlobalConfig", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		t.Logf("Using temporary directory: %s, original working directory: %s", tmpDir, currentWd)
		SetSearchPaths(tmpDir)
		configContent := []byte(`
prefix: "Global"
suffix: "All"
explicit:
  - from: GlobalOld
    to: GlobalNew
regex:
  - pattern: "GlobalRegex$"
    replace: "GlobalReplacement"
ignore:
  - "GlobalIgnored"
types: # Explicitly enable inheritance for types
  inherit_explicit: true
  inherit_regex: true
  inherit_ignore: true
  inherit_prefix: true
  inherit_suffix: true
`)

		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		// Add this assertion to verify inherit_explicit flag
		assert.True(t, cfg.Types.InheritExplicit, "cfg.Types.InheritExplicit should be true after unmarshalling")

		// Verify top-level fields
		assert.Equal(t, "Global", cfg.Prefix)
		assert.Equal(t, "All", cfg.Suffix)
		assert.Len(t, cfg.Explicit, 1)
		assert.Equal(t, "GlobalOld", cfg.Explicit[0].From)
		assert.Equal(t, "GlobalNew", cfg.Explicit[0].To)
		assert.Len(t, cfg.Regex, 1)
		assert.Equal(t, "GlobalRegex$", cfg.Regex[0].Pattern)
		assert.Len(t, cfg.Ignore, 1)
		assert.Equal(t, "GlobalIgnored", cfg.Ignore[0])

		// Verify compiled rules for types (should inherit from top-level)
		assert.Equal(t, "GlobalNew", cfg.CompiledTypes.Rules[0].To)         // Explicit
		assert.Equal(t, "Global", cfg.CompiledTypes.Rules[1].Value)         // Prefix
		assert.Equal(t, "All", cfg.CompiledTypes.Rules[2].Value)            // Suffix
		assert.Equal(t, "GlobalRegex$", cfg.CompiledTypes.Rules[3].Pattern) // Regex
		assert.Equal(t, "GlobalIgnored", cfg.CompiledTypes.Ignore[0])       // Ignore
	})

	t.Run("CategorySpecificConfigOverridesGlobal", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
prefix: "Global"
types:
  inherit_explicit: true
  inherit_regex: true
  inherit_ignore: true
  inherit_prefix: true
  inherit_suffix: true
  prefix: "TypeSpecific"
  explicit:
    - from: TypeOld
      to: TypeNew
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify top-level prefix is still "Global"
		assert.Equal(t, "Global", cfg.Prefix)

		// Verify types category uses its own prefix and explicit
		assert.Equal(t, "TypeSpecific", cfg.Types.Prefix)
		assert.Len(t, cfg.Types.Explicit, 1)
		assert.Equal(t, "TypeOld", cfg.Types.Explicit[0].From)
		assert.Equal(t, "TypeNew", cfg.Types.Explicit[0].To)

		// Verify compiled types rules reflect category-specific overrides
		assert.Equal(t, "TypeNew", cfg.CompiledTypes.Rules[0].To)
		assert.Equal(t, "TypeSpecific", cfg.CompiledTypes.Rules[1].Value)
	})

	t.Run("PackageSpecificConfigOverridesCategory", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
types:
  prefix: "GlobalType"
packages:
  - import: "test.com/pkg1"
    types:
      prefix: "Pkg1Type"
      explicit:
        - from: Pkg1Old
          to: Pkg1New
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify global type prefix
		assert.Equal(t, "GlobalType", cfg.Types.Prefix)

		// Verify package-specific type prefix
		assert.Len(t, cfg.Packages, 1)
		pkg1 := cfg.Packages[0]
		assert.Equal(t, "Pkg1Type", pkg1.Types.Prefix)
		assert.Len(t, pkg1.Types.Explicit, 1)
		assert.Equal(t, "Pkg1Old", pkg1.Types.Explicit[0].From)
		assert.Equal(t, "Pkg1New", pkg1.Types.Explicit[0].To)

		// Verify compiled rules for package types reflect package-specific overrides
		assert.Equal(t, "Pkg1New", pkg1.CompiledTypes.Rules[0].To)
		assert.Equal(t, "Pkg1Type", pkg1.CompiledTypes.Rules[1].Value)
	})

	t.Run("InheritExplicitTrueMerge", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
explicit:
  - from: GlobalA
    to: GlobalA_New
  - from: GlobalB
    to: GlobalB_New
types:
  inherit_explicit: true
  explicit:
    - from: TypeC
      to: TypeC_New
    - from: GlobalA
      to: TypeA_Override # Should override GlobalA
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types explicit rules are merged
		expectedExplicit := map[string]string{
			"GlobalA": "TypeA_Override",
			"GlobalB": "GlobalB_New",
			"TypeC":   "TypeC_New",
		}
		compiledExplicitMap := make(map[string]string)
		for _, rule := range cfg.CompiledTypes.Rules {
			if rule.Type == "explicit" {
				compiledExplicitMap[rule.From] = rule.To
			}
		}
		assert.True(t, reflect.DeepEqual(expectedExplicit, compiledExplicitMap))
	})

	t.Run("InheritExplicitFalseOverride", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
explicit:
  - from: GlobalA
    to: GlobalA_New
types:
  inherit_explicit: false # Explicitly do not inherit
  explicit:
    - from: TypeA
      to: TypeA_New # Only this should be present
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types explicit rules are only from category
		expectedExplicit := map[string]string{
			"TypeA": "TypeA_New",
		}
		compiledExplicitMap := make(map[string]string)
		for _, rule := range cfg.CompiledTypes.Rules {
			if rule.Type == "explicit" {
				compiledExplicitMap[rule.From] = rule.To
			}
		}
		assert.True(t, reflect.DeepEqual(expectedExplicit, compiledExplicitMap))
	})

	t.Run("InheritExplicitFalseClear", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
explicit:
  - from: GlobalA
    to: GlobalA_New
types:
  inherit_explicit: false # Explicitly do not inherit
  explicit: [] # Explicitly empty
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types explicit rules are empty
		for _, rule := range cfg.CompiledTypes.Rules {
			assert.NotEqual(t, "explicit", rule.Type)
		}
	})

	t.Run("InheritRegexTrueAppend", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
regex:
  - pattern: "GlobalR1"
    replace: "G1"
  - pattern: "GlobalR2"
    replace: "G2"
types:
  inherit_regex: true
  regex:
    - pattern: "TypeR3"
      replace: "T3"
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types regex rules are appended
		expectedRegex := []RegexRule{
			{Pattern: "GlobalR1", Replace: "G1"},
			{Pattern: "GlobalR2", Replace: "G2"},
			{Pattern: "TypeR3", Replace: "T3"},
		}
		compiledRegex := []RegexRule{} // Initialize as empty slice
		for _, rule := range cfg.CompiledTypes.Rules {
			if rule.Type == "regex" {
				compiledRegex = append(compiledRegex, RegexRule{Pattern: rule.Pattern, Replace: rule.Replace})
			}
		}
		assert.True(t, reflect.DeepEqual(expectedRegex, compiledRegex))
	})

	t.Run("InheritRegexFalseOverride", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
regex:
  - pattern: "GlobalR1"
    replace: "G1"
types:
  inherit_regex: false
  regex:
    - pattern: "TypeR2"
      replace: "T2"
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types regex rules are only from category
		expectedRegex := []RegexRule{
			{Pattern: "TypeR2", Replace: "T2"},
		}
		compiledRegex := []RegexRule{} // Initialize as empty slice
		for _, rule := range cfg.CompiledTypes.Rules {
			if rule.Type == "regex" {
				compiledRegex = append(compiledRegex, RegexRule{Pattern: rule.Pattern, Replace: rule.Replace})
			}
		}
		assert.True(t, reflect.DeepEqual(expectedRegex, compiledRegex))
	})

	t.Run("InheritRegexFalseClear", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
regex:
  - pattern: "GlobalR1"
    replace: "G1"
types:
  inherit_regex: false
  regex: [] # Explicitly empty
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types regex rules are empty
		for _, rule := range cfg.CompiledTypes.Rules {
			assert.NotEqual(t, "regex", rule.Type)
		}
	})

	t.Run("InheritIgnoreTrueAppend", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
ignore:
  - "GlobalIgnore1"
  - "GlobalIgnore2"
types:
  inherit_ignore: true
  ignore:
    - "TypeIgnore3"
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types ignore rules are appended
		expectedIgnore := []string{"GlobalIgnore1", "GlobalIgnore2", "TypeIgnore3"}
		assert.True(t, reflect.DeepEqual(expectedIgnore, cfg.CompiledTypes.Ignore))
	})

	t.Run("InheritIgnoreFalseOverride", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
ignore:
  - "GlobalIgnore1"
types:
  inherit_ignore: false
  ignore:
    - "TypeIgnore2"
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types ignore rules are only from category
		expectedIgnore := []string{"TypeIgnore2"}
		assert.True(t, reflect.DeepEqual(expectedIgnore, cfg.CompiledTypes.Ignore))
	})

	t.Run("InheritIgnoreFalseClear", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
ignore:
  - "GlobalIgnore1"
types:
  inherit_ignore: false
  ignore: [] # Explicitly empty
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Verify compiled types ignore rules are empty
		assert.Empty(t, cfg.CompiledTypes.Ignore)
	})

	t.Run("PrefixSuffixInheritanceLogic", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		configContent := []byte(`
prefix: "GlobalP"
suffix: "GlobalS"
types:
  # Case 1: Category prefix is empty, inherit_prefix is true -> should inherit GlobalP
  inherit_prefix: true
  # Case 2: Category suffix is empty, inherit_suffix is false -> should be empty
  inherit_suffix: false
functions:
  # Case 3: Category prefix is non-empty, inherit_prefix is true -> should use category prefix
  prefix: "FuncP"
  inherit_prefix: true
  # Case 4: Category suffix is non-empty, inherit_suffix is false -> should use category suffix
  suffix: "FuncS"
  inherit_suffix: false
methods:
  # Case 5: Category prefix is empty, inherit_prefix is false -> should be empty
  inherit_prefix: false
  # Case 6: Category suffix is non-empty, inherit_suffix is true -> should use category suffix
  suffix: "MethodS"
  inherit_suffix: true
`)
		setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

		cfg, err := LoadConfig("")
		assert.NoError(t, err)

		// Test Types category
		assert.Equal(t, "GlobalP", cfg.CompiledTypes.Rules[0].Value) // Prefix: Inherited
		assert.Empty(t, cfg.CompiledTypes.Rules[1].Value)            // Suffix: Explicitly cleared by inherit_suffix:false and empty suffix

		// Test Functions category
		assert.Equal(t, "FuncP", cfg.CompiledFunctions.Rules[0].Value) // Prefix: Uses category prefix
		assert.Equal(t, "FuncS", cfg.CompiledFunctions.Rules[1].Value) // Suffix: Uses category suffix

		// Test Methods category
		assert.Empty(t, cfg.CompiledMethods.Rules[0].Value)            // Prefix: Explicitly cleared by inherit_prefix:false and empty prefix
		assert.Equal(t, "MethodS", cfg.CompiledMethods.Rules[1].Value) // Suffix: Uses category suffix
	})

	t.Run("FileLevelConfigOverridesProjectLevel", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		setupTestConfigFile(t, tmpDir, ".adptool.yaml", []byte(`prefix: "ProjectLevel"`))

		fileConfigContent := []byte(`prefix: "FileLevel"`)
		fileConfigPath := setupTestConfigFile(t, tmpDir, "file_level_config.yaml", fileConfigContent)

		cfg, err := LoadConfig(fileConfigPath)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "FileLevel", cfg.Prefix) // File-level should override
	})

	t.Run("ErrorOnInvalidYAML", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		fileConfigPath := setupTestConfigFile(t, tmpDir, "invalid.yaml", []byte(`prefix: [invalid`))

		_, err := LoadConfig(fileConfigPath)
		assert.Error(t, err)
	})

	t.Run("ErrorOnSpecifiedFileNotFound", func(t *testing.T) {
		currentWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		defer os.Chdir(currentWd)

		_, err := LoadConfig("nonexistent.yaml")
		assert.Error(t, err)
	})
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}
