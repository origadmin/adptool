package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	t.Run("DefaultConfigWhenNoFileExists", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			cfg, err := LoadConfig("")
			assert.NoError(t, err)
			assert.NotNil(t, cfg)

			// Verify that all fields have their default zero-values.
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

			// Verify category-specific defaults are also zero-valued.
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

			// Verify compiled rules are empty, ensuring no hidden defaults.
			assert.Empty(t, cfg.CompiledTypes.Rules)
			assert.Empty(t, cfg.CompiledTypes.Ignore)
			assert.Empty(t, cfg.CompiledFunctions.Rules)
			assert.Empty(t, cfg.CompiledFunctions.Ignore)
			assert.Empty(t, cfg.CompiledMethods.Rules)
			assert.Empty(t, cfg.CompiledMethods.Ignore)
			assert.Empty(t, cfg.Packages)
		})
	})

	t.Run("TopLevelGlobalConfig", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify top-level fields are correctly loaded.
			assert.Equal(t, "Global", cfg.Prefix)
			assert.Equal(t, "All", cfg.Suffix)
			assert.Len(t, cfg.Explicit, 1)
			assert.Equal(t, "GlobalOld", cfg.Explicit[0].From)
			assert.Equal(t, "GlobalNew", cfg.Explicit[0].To)
			assert.Len(t, cfg.Regex, 1)
			assert.Equal(t, "GlobalRegex$", cfg.Regex[0].Pattern)
			assert.Len(t, cfg.Ignore, 1)
			assert.Equal(t, "GlobalIgnored", cfg.Ignore[0])

			// Verify that inheritance flags are correctly unmarshalled.
			assert.True(t, cfg.Types.InheritExplicit, "cfg.Types.InheritExplicit should be true")

			// Verify compiled rules for types inherit from top-level settings.
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "explicit", From: "GlobalOld", To: "GlobalNew"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "prefix", Value: "Global"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "suffix", Value: "All"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "GlobalRegex$", Replace: "GlobalReplacement"})
			assert.Contains(t, cfg.CompiledTypes.Ignore, "GlobalIgnored")
		})
	})

	t.Run("CategorySpecificConfigOverridesGlobal", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify top-level prefix remains unchanged.
			assert.Equal(t, "Global", cfg.Prefix)

			// Verify types category uses its own specific settings.
			assert.Equal(t, "TypeSpecific", cfg.Types.Prefix)
			assert.Len(t, cfg.Types.Explicit, 1)
			assert.Equal(t, "TypeOld", cfg.Types.Explicit[0].From)
			assert.Equal(t, "TypeNew", cfg.Types.Explicit[0].To)

			// Verify compiled types rules reflect category-specific overrides.
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "explicit", From: "TypeOld", To: "TypeNew"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "prefix", Value: "TypeSpecific"})
		})
	})

	t.Run("PackageSpecificConfigOverridesCategory", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify global type prefix is correctly loaded.
			assert.Equal(t, "GlobalType", cfg.Types.Prefix)

			// Verify package-specific settings override category settings.
			assert.Len(t, cfg.Packages, 1)
			pkg1 := cfg.Packages[0]
			assert.Equal(t, "Pkg1Type", pkg1.Types.Prefix)
			assert.Len(t, pkg1.Types.Explicit, 1)
			assert.Equal(t, "Pkg1Old", pkg1.Types.Explicit[0].From)
			assert.Equal(t, "Pkg1New", pkg1.Types.Explicit[0].To)

			// Verify compiled rules for the package reflect package-specific overrides.
			assert.Contains(t, pkg1.CompiledTypes.Rules, &RenameRule{Type: "explicit", From: "Pkg1Old", To: "Pkg1New"})
			assert.Contains(t, pkg1.CompiledTypes.Rules, &RenameRule{Type: "prefix", Value: "Pkg1Type"})
		})
	})

	t.Run("InheritExplicitTrueWithOverride", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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
    - from: GlobalA # This should override the global setting for GlobalA
      to: TypeA_Override
`)
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

			cfg, err := LoadConfig("")
			assert.NoError(t, err)

			// Verify that compiled explicit rules are correctly merged and overridden.
			expectedExplicit := map[string]string{
				"GlobalA": "TypeA_Override", // Overridden by type-specific config
				"GlobalB": "GlobalB_New",    // Inherited from global
				"TypeC":   "TypeC_New",      // Defined in type-specific config
			}
			compiledExplicitMap := make(map[string]string)
			for _, rule := range cfg.CompiledTypes.Rules {
				if rule.Type == "explicit" {
					compiledExplicitMap[rule.From] = rule.To
				}
			}
			assert.True(t, reflect.DeepEqual(expectedExplicit, compiledExplicitMap))
		})
	})

	t.Run("InheritExplicitFalseDisablesMerge", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			configContent := []byte(`
explicit:
  - from: GlobalA
    to: GlobalA_New
types:
  inherit_explicit: false # Explicitly disable inheritance
  explicit:
    - from: TypeA
      to: TypeA_New # Only this rule should be present
`)
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

			cfg, err := LoadConfig("")
			assert.NoError(t, err)

			// Verify that only category-specific explicit rules are present.
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
	})

	t.Run("InheritExplicitFalseWithEmptyListClearsRules", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			configContent := []byte(`
explicit:
  - from: GlobalA
    to: GlobalA_New
types:
  inherit_explicit: false # Disable inheritance
  explicit: [] # Explicitly empty list
`)
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

			cfg, err := LoadConfig("")
			assert.NoError(t, err)

			// Verify that no explicit rules are present in the compiled types rules.
			for _, rule := range cfg.CompiledTypes.Rules {
				assert.NotEqual(t, "explicit", rule.Type)
			}
		})
	})

	t.Run("InheritRegexTrueAppendsRules", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify that regex rules from both global and category are present.
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "GlobalR1", Replace: "G1"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "GlobalR2", Replace: "G2"})
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "TypeR3", Replace: "T3"})
		})
	})

	t.Run("InheritRegexFalseDisablesAppend", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify that only category-specific regex rules are present.
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "TypeR2", Replace: "T2"})
			assert.NotContains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "regex", Pattern: "GlobalR1", Replace: "G1"})
		})
	})

	t.Run("InheritRegexFalseWithEmptyListClearsRules", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			configContent := []byte(`
regex:
  - pattern: "GlobalR1"
    replace: "G1"
types:
  inherit_regex: false
  regex: [] # Explicitly empty list
`)
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

			cfg, err := LoadConfig("")
			assert.NoError(t, err)

			// Verify that no regex rules are present in the compiled types rules.
			for _, rule := range cfg.CompiledTypes.Rules {
				assert.NotEqual(t, "regex", rule.Type)
			}
		})
	})

	t.Run("InheritIgnoreTrueAppendsRules", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify that ignore rules from both global and category are present.
			expectedIgnore := []string{"GlobalIgnore1", "GlobalIgnore2", "TypeIgnore3"}
			assert.ElementsMatch(t, expectedIgnore, cfg.CompiledTypes.Ignore)
		})
	})

	t.Run("InheritIgnoreFalseDisablesAppend", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify that only category-specific ignore rules are present.
			expectedIgnore := []string{"TypeIgnore2"}
			assert.True(t, reflect.DeepEqual(expectedIgnore, cfg.CompiledTypes.Ignore))
		})
	})

	t.Run("InheritIgnoreFalseWithEmptyListClearsRules", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			configContent := []byte(`
ignore:
  - "GlobalIgnore1"
types:
  inherit_ignore: false
  ignore: [] # Explicitly empty list
`)
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", configContent)

			cfg, err := LoadConfig("")
			assert.NoError(t, err)

			// Verify that the compiled ignore list for types is empty.
			assert.Empty(t, cfg.CompiledTypes.Ignore)
		})
	})

	t.Run("PrefixSuffixInheritanceLogic", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
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

			// Verify prefix/suffix logic for the Types category.
			assert.Contains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "prefix", Value: "GlobalP"})
			assert.NotContains(t, cfg.CompiledTypes.Rules, &RenameRule{Type: "suffix", Value: "GlobalS"})

			// Verify prefix/suffix logic for the Functions category.
			assert.Contains(t, cfg.CompiledFunctions.Rules, &RenameRule{Type: "prefix", Value: "FuncP"})
			assert.Contains(t, cfg.CompiledFunctions.Rules, &RenameRule{Type: "suffix", Value: "FuncS"})

			// Verify prefix/suffix logic for the Methods category.
			assert.NotContains(t, cfg.CompiledMethods.Rules, &RenameRule{Type: "prefix", Value: "GlobalP"})
			assert.Contains(t, cfg.CompiledMethods.Rules, &RenameRule{Type: "suffix", Value: "MethodS"})
		})
	})

	t.Run("FileLevelConfigOverridesProjectLevel", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			// Setup a project-level config file.
			setupTestConfigFile(t, tmpDir, ".adptool.yaml", []byte(`prefix: "ProjectLevel"`))

			// Setup a file-level config that should override the project-level one.
			fileConfigContent := []byte(`prefix: "FileLevel"`)
			fileConfigPath := setupTestConfigFile(t, tmpDir, "file_level_config.yaml", fileConfigContent)

			cfg, err := LoadConfig(fileConfigPath)
			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, "FileLevel", cfg.Prefix) // File-level config should take precedence.
		})
	})

	t.Run("ErrorOnInvalidYAML", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			// Create an invalid YAML file.
			fileConfigPath := setupTestConfigFile(t, tmpDir, "invalid.yaml", []byte(`prefix: [invalid`))

			_, err := LoadConfig(fileConfigPath)
			assert.Error(t, err) // Expect an error due to parsing failure.
		})
	})

	t.Run("ErrorOnSpecifiedFileNotFound", func(t *testing.T) {
		runInTempDir(t, func(tmpDir string) {
			// Attempt to load a non-existent config file.
			_, err := LoadConfig("nonexistent.yaml")
			assert.Error(t, err) // Expect an error as the file does not exist.
		})
	})
}
