package loader

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/origadmin/adptool/internal/config"
)

// getAdptoolModuleRoot dynamically determines the root directory of the adptool module.
// This is necessary because test execution context can vary.
func getAdptoolModuleRoot() string {
	_, b, _, _ := runtime.Caller(0)
	// The directory of this file (loader_test.go)
	basepath := filepath.Dir(b)

	// Navigate up until we find the adptool module root (where go.mod is located)
	// In this project structure, adptool's go.mod is at tools/adptool/go.mod
	// loader_test.go is at tools/adptool/internal/loader/loader_test.go
	// So, we need to go up two directories from loader_test.go's directory.
	return filepath.Join(basepath, "..", "..")
}

func TestLoadConfigFile(t *testing.T) {
	fullExpectedConfig := &config.Config{
		Defaults: &config.Defaults{
			Mode: &config.Mode{
				Strategy: "replace",
				Prefix:   "append",
				Explicit: "merge",
			},
		},
		Ignores: []string{},
		Props: []*config.PropsEntry{
			{Name: "GlobalVar1", Value: "globalValue1"},
			{Name: "GlobalVar2", Value: "globalValue2"},
		},
		Types: []*config.TypeRule{
			{
				Name:     "*",
				Disabled: false,
				Kind:     "struct",
				Pattern:  "alias",
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{{From: "GlobalTypeOld", To: "GlobalTypeNew"}},
					Ignores:  nil,
				},
				Methods: []*config.MemberRule{
					{Name: "*", RuleSet: config.RuleSet{Prefix: "GlobalMethod"}},
				},
				Fields: []*config.MemberRule{
					{Name: "*", RuleSet: config.RuleSet{Suffix: "GlobalField"}},
				},
			},
			{
				Name:     "MyStruct",
				Disabled: false,
				Kind:     "struct",
				Pattern:  "wrap",
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{{From: "MyStructOld", To: "MyStructNew"}},
				},
				Methods: []*config.MemberRule{
					{Name: "DoSomething", Disabled: false, RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingOld", To: "DoSomethingNew"}}}},
					{Name: "Calculate", RuleSet: config.RuleSet{Prefix: "Calc"}},
				},
				Fields: []*config.MemberRule{
					{Name: "Data", RuleSet: config.RuleSet{Suffix: "Value"}},
				},
			},
			{
				Name:     "MyInterface",
				Disabled: false,
				Kind:     "interface",
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{{From: "MyInterfaceOld", To: "MyInterfaceNew"}},
				},
			},
		},
		Functions: []*config.FuncRule{
			{
				Name:     "*",
				Disabled: false,
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{{From: "GlobalFuncOld", To: "GlobalFuncNew"}},
				},
			},
			{
				Name:     "SpecificFunc",
				Disabled: false,
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{{From: "SpecificFuncOld", To: "SpecificFuncNew"}},
				},
			},
		},
		Variables: []*config.VarRule{
			{
				Name:     "*",
				Disabled: false,

				RuleSet: config.RuleSet{Prefix: "GlobalVar"},
			},
			{
				Name:     "SpecificVar",
				Disabled: false,
				RuleSet:  config.RuleSet{Suffix: "Specific"},
			},
		},
		Constants: []*config.ConstRule{
			{
				Name:     "*",
				Disabled: false,
				RuleSet:  config.RuleSet{Ignores: nil},
			},
			{
				Name:     "SpecificConst",
				Disabled: true,
				RuleSet:  config.RuleSet{},
			},
		},
		Packages: []*config.Package{
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "mypkg",
				Props: []*config.PropsEntry{
					{
						Name:  "PackageVar1",
						Value: "packageValue1",
					},
				},
				Types: []*config.TypeRule{
					{Name: "*", Pattern: "copy", RuleSet: config.RuleSet{}},
					{Name: "PackageStruct", Pattern: "define", RuleSet: config.RuleSet{}},
				},
				Functions: []*config.FuncRule{
					{Name: "*", RuleSet: config.RuleSet{Prefix: "PackageFunc"}},
				},
			},
		},
	}

	tests := []struct {
		name           string
		configFilePath string
		expectError    bool
	}{
		{
			name:           "Full Configuration Load - YAML",
			configFilePath: "testdata/config/full_config.yaml",
		},
		{
			name:           "Full Configuration Load - JSON",
			configFilePath: "testdata/config/full_config.json",
		},
		{
			name:           "Full Configuration Load - TOML",
			configFilePath: "testdata/config/full_config.toml",
		},
		{
			name:           "Empty Config",
			configFilePath: "testdata/config/empty_config.yaml",
		},
		{
			name:           "Invalid YAML",
			configFilePath: "testdata/config/invalid_config.yaml",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the testdata directory exists
			testDataDir := filepath.Join(t.TempDir(), "testdata")
			err := os.MkdirAll(testDataDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create testdata directory: %v", err)
			}

			// Copy the config file from testdata to the temp testdata directory
			srcPath := filepath.Join(getAdptoolModuleRoot(), tt.configFilePath)
			dstPath := filepath.Join(testDataDir, filepath.Base(tt.configFilePath))

			srcFile, err := os.Open(srcPath)
			if err != nil {
				t.Fatalf("Failed to open source config file %s: %v", srcPath, err)
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				t.Fatalf("Failed to create destination config file %s: %v", dstPath, err)
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				t.Fatalf("Failed to copy config file: %v", err)
			}

			// Load the config
			cfg, err := LoadConfigFile(dstPath)

			// For empty config, the expected config is config.New()
			// For invalid config, we expect an error
			// For full configs, we expect fullExpectedConfig
			var expected *config.Config
			switch tt.name {
			case "Empty Config":
				expected = config.New()
			case "Invalid YAML":
				// No expected config for invalid cases, just error check
			default:
				expected = fullExpectedConfig
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Compare the loaded config with the expected config
			if !cmp.Equal(cfg, expected) {
				t.Errorf("Loaded config mismatch.\nExpected: %+v\nActual:   %+v", expected, cfg)
				// Optionally, print differences for easier debugging
				diff := cmp.Diff(expected, cfg)
				t.Errorf("Diff: %s", diff)
			}
		})
	}
}

func TestLoadAllFieldsConfigFile(t *testing.T) {
	// The config file to test, which contains all possible fields.
	configFilePath := filepath.Join(getAdptoolModuleRoot(), "testdata", "parser", "config.yaml")

	// The expected configuration that should be loaded from the YAML file.
	expectedCfg := &config.Config{
		PackageName: "my_package",
		Ignores:     []string{"file1.go", "dir1/file2.go"},
		Defaults: &config.Defaults{
			Mode: &config.Mode{
				Strategy: "replace",
				Prefix:   "append",
				Suffix:   "prepend",
				Explicit: "merge",
				Regex:    "merge",
				Ignores:  "merge",
			},
		},
		Props: []*config.PropsEntry{
			{Name: "GlobalVar1", Value: "globalValue1"},
			{Name: "GlobalVar2", Value: "globalValue2"},
		},
		Packages: []*config.Package{
			{
				Import: "github.com/my/package/v1",
				Alias:  "mypkg",
				Path:   "./vendor/my/package/v1",
				Props: []*config.PropsEntry{
					{Name: "PkgVar", Value: "PkgValue"},
				},
				Types:     nil,
				Functions: nil,
				Variables: nil,
				Constants: nil,
			},
		},
		Types: []*config.TypeRule{
			{
				Name:     "MyStruct",
				Kind:     "struct",
				Pattern:  "wrap",
				Disabled: true,
				Methods: []*config.MemberRule{
					{
						Name: "DoSomething",
						RuleSet: config.RuleSet{
							Prefix: "Pre",
							Suffix: "Post",
						},
					},
				},
				Fields: []*config.MemberRule{
					{
						Name: "MyField",
						RuleSet: config.RuleSet{
							Transforms: &config.Transform{
								Before: "(.*)",
								After:  "New$1",
							},
						},
					},
				},
			},
		},
		Functions: []*config.FuncRule{
			{
				Name:     "MyFunc",
				Disabled: false,
				RuleSet: config.RuleSet{
					Regex: []*config.RegexRule{
						{Pattern: "Old(.*)", Replace: "New$1"},
					},
				},
			},
		},
		Variables: []*config.VarRule{
			{
				Name: "MyVar",
				RuleSet: config.RuleSet{
					Explicit: []*config.ExplicitRule{
						{From: "MyVar", To: "NewVar"},
					},
				},
			},
		},
		Constants: []*config.ConstRule{
			{
				Name: "MyConst",
				RuleSet: config.RuleSet{
					Ignores: []string{"IgnoredConst"},
				},
			},
		},
	}

	// Load the configuration from the specified file.
	loadedCfg, err := LoadConfigFile(configFilePath)
	if err != nil {
		t.Fatalf("Failed to load config file '%s': %v", configFilePath, err)
	}

	// Viper might unmarshal empty collections as nil, while our expected struct might have empty slices.
	// To ensure a fair comparison, we'll use cmp.Diff which can handle this gracefully.
	if diff := cmp.Diff(expectedCfg, loadedCfg); diff != "" {
		t.Errorf("Loaded config mismatch (-want +got):\n%s", diff)
	}
}
