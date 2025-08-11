package loader

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/origadmin/adptool/internal/config"
)

func TestLoadConfigFile(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectedConfig *config.Config
		expectError    bool
	}{
		{
			name: "Full Configuration Load",
			yamlContent: `
defaults:
  mode:
    strategy: "replace"
    prefix: "append"
    explicit: "merge"
vars:
  GlobalVar1: "globalValue1"
  GlobalVar2: "globalValue2"

types:
  - name: "*"
    disabled: false
    kind: "struct"
    pattern: "alias"
    explicit:
      - { from: "GlobalTypeOld", to: "GlobalTypeNew" }
    ignore: ["GlobalIgnoredType"]
    methods:
      - name: "*"
        prefix: "GlobalMethod"
    fields:
      - name: "*"
        suffix: "GlobalField"

  - name: "MyStruct"
    disabled: false
    kind: "struct"
    pattern: "wrap"
    explicit:
      - { from: "MyStructOld", to: "MyStructNew" }
    methods:
      - name: "DoSomething"
        disabled: false
        explicit:
          - { from: "DoSomethingOld", to: "DoSomethingNew" }
      - name: "Calculate"
        prefix: "Calc"
    fields:
      - name: "Data"
        suffix: "Value"

  - name: "MyInterface"
    disabled: false
    kind: "interface"
    explicit:
      - { from: "MyInterfaceOld", to: "MyInterfaceNew" }

functions:
  - name: "*"
    disabled: false
    explicit:
      - { from: "GlobalFuncOld", to: "GlobalFuncNew" }
  - name: "SpecificFunc"
    disabled: false
    explicit:
      - { from: "SpecificFuncOld", to: "SpecificFuncNew" }

variables:
  - name: "*"
    disabled: false
    prefix: "GlobalVar"
  - name: "SpecificVar"
    disabled: false
    suffix: "Specific"

constants:
  - name: "*"
    disabled: false
    ignore: ["GlobalIgnoredConst"]
  - name: "SpecificConst"
    disabled: true

packages:
  - import: "github.com/my/package"
    alias: "mypkg"
    vars:
      PackageVar1: "packageValue1"
    types:
      - name: "*"
        pattern: "copy"
      - name: "PackageStruct"
        pattern: "define"
    functions:
      - name: "*"
        prefix: "PackageFunc"
`,
			expectedConfig: &config.Config{
				Defaults: &config.Defaults{
					Mode: &config.Mode{
						Strategy: "replace",
						Prefix:   "append",
						Explicit: "merge",
					},
				},
				Vars: map[string]string{
					"GlobalVar1": "globalValue1",
					"GlobalVar2": "globalValue2",
				},
				Types: []*config.TypeRule{
					{
						Name:     "*",
						Disabled: false,
						Kind:     "struct",
						Pattern:  "alias",
						RuleSet: &config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "GlobalTypeOld", To: "GlobalTypeNew"}},
							Ignore:   []string{"GlobalIgnoredType"},
						},
						Methods: []*config.MemberRule{
							{Name: "*", RuleSet: &config.RuleSet{Prefix: "GlobalMethod"}},
						},
						Fields: []*config.MemberRule{
							{Name: "*", RuleSet: &config.RuleSet{Suffix: "GlobalField"}},
						},
					},
					{
						Name:     "MyStruct",
						Disabled: false,
						Kind:     "struct",
						Pattern:  "wrap",
						RuleSet: &config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "MyStructOld", To: "MyStructNew"}},
						},
						Methods: []*config.MemberRule{
							{Name: "DoSomething", Disabled: false, RuleSet: &config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingOld", To: "DoSomethingNew"}}}},
							{Name: "Calculate", RuleSet: &config.RuleSet{Prefix: "Calc"}},
						},
						Fields: []*config.MemberRule{
							{Name: "Data", RuleSet: &config.RuleSet{Suffix: "Value"}},
						},
					},
					{
						Name:     "MyInterface",
						Disabled: false,
						Kind:     "interface",
						RuleSet: &config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "MyInterfaceOld", To: "MyInterfaceNew"}},
						},
					},
				},
				Functions: []*config.FuncRule{
					{
						Name:     "*",
						Disabled: false,
						RuleSet: &config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "GlobalFuncOld", To: "GlobalFuncNew"}},
						},
					},
					{
						Name:     "SpecificFunc",
						Disabled: false,
						RuleSet: &config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "SpecificFuncOld", To: "SpecificFuncNew"}},
						},
					},
				},
				Variables: []*config.VarRule{
					{
						Name:     "*",
						Disabled: false,
						RuleSet:  &config.RuleSet{Prefix: "GlobalVar"},
					},
					{
						Name:     "SpecificVar",
						Disabled: false,
						RuleSet:  &config.RuleSet{Suffix: "Specific"},
					},
				},
				Constants: []*config.ConstRule{
					{
						Name:     "*",
						Disabled: false,
						RuleSet:  &config.RuleSet{Ignore: []string{"GlobalIgnoredConst"}},
					},
					{
						Name:     "SpecificConst",
						Disabled: true,
						RuleSet:  &config.RuleSet{},
					},
				},
				Packages: []*config.Package{
					{
						Import: "github.com/my/package",
						Alias:  "mypkg",
						Vars:   map[string]string{"PackageVar1": "packageValue1"},
						Types: []*config.TypeRule{
							{Name: "*", Pattern: "copy", RuleSet: &config.RuleSet{}},
							{Name: "PackageStruct", Pattern: "define", RuleSet: &config.RuleSet{}},
						},
						Functions: []*config.FuncRule{
							{Name: "*", RuleSet: &config.RuleSet{Prefix: "PackageFunc"}},
						},
					},
				},
			},
		},
		{
			name:        "Empty Config",
			yamlContent: "",
			expectedConfig: &config.Config{
				Defaults:  &config.Defaults{Mode: &config.Mode{}},
				Vars:      make(map[string]string),
				Types:     make([]*config.TypeRule, 0),
				Functions: make([]*config.FuncRule, 0),
				Variables: make([]*config.VarRule, 0),
				Constants: make([]*config.ConstRule, 0),
				Packages:  make([]*config.Package, 0),
			},
		},
		{
			name:        "Invalid YAML",
			yamlContent: "invalid: - ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file for the test case
			tmpFile, err := ioutil.TempFile(t.TempDir(), "test_config_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name()) // Clean up

			if _, err := tmpFile.WriteString(tt.yamlContent); err != nil {
				t.Fatalf("Failed to write to temp file: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("Failed to close temp file: %v", err)
			}

			// Load the config
			cfg, err := LoadConfigFile(tmpFile.Name())

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
			if !reflect.DeepEqual(cfg, tt.expectedConfig) {
				t.Errorf("Loaded config mismatch.\nExpected: %+v\nActual:   %+v", tt.expectedConfig, cfg)
				// Optionally, print differences for easier debugging
				// diff := cmp.Diff(tt.expectedConfig, cfg)
				// t.Errorf("Diff: %s", diff)
			}
		})
	}
}
