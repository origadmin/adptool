package parser_test

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/loader"
	"github.com/origadmin/adptool/internal/parser"
)

// getModuleRoot returns the absolute path to the adptool module root.
func getModuleRoot() string {
	_, b, _, _ := runtime.Caller(0)
	// b is the path to this file: .../tools/adptool/internal/parser/parser_test.go
	// Go up two levels to get to tools/adptool
	return filepath.Join(filepath.Dir(b), "..", "..")
}

func TestParseFileDirectives(t *testing.T) {
	tests := []struct {
		name           string
		filePath       string
		expectedConfig *config.Config
		expectError    bool
	}{
		{
			name:     "Full Directives Parse",
			filePath: filepath.Join(getModuleRoot(), "testdata", "parser_test_file.go"),
			expectedConfig: &config.Config{
				Defaults: &config.Defaults{
					Mode: &config.Mode{
						Strategy: "replace",
						Prefix:   "append",
						Suffix:   "append",
						Explicit: "merge",
						Regex:    "merge",
						Ignore:   "merge",
					},
				},
				Vars: []*config.VarEntry{
					{
						Name:  "GlobalVar1",
						Value: "globalValue1",
					},
					{
						Name:  "GlobalVar2",
						Value: "globalValue2",
					},
				},
				Packages: []*config.Package{
					{
						Import: "github.com/my/package/v1",
						Alias:  "mypkg",
						Path:   "./vendor/my/package/v1",
						Vars: []*config.VarEntry{
							{
								Name:  "PackageVar1",
								Value: "packageValue1",
							},
						},
						Types: []*config.TypeRule{
							{
								Name:    "MyStructInPackage",
								Kind:    "struct",
								Pattern: "wrap",
								Methods: []*config.MemberRule{
									{
										Name:    "DoSomethingInPackage",
										RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingInPackage", To: "DoSomethingNewInPackage"}}},
									},
								},
							},
						},
						Functions: []*config.FuncRule{
							{
								Name:    "MyFuncInPackage",
								RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "MyFuncInPackage", To: "MyNewFuncInPackage"}}},
							},
						},
					},
				},
				Types: []*config.TypeRule{
					{
						Name:    "*",
						RuleSet: config.RuleSet{},
						Kind:    "struct",
						Pattern: "wrap",
					},
					{
						Name:    "ext1.TypeA",
						RuleSet: config.RuleSet{},
						Kind:    "type",
						Methods: []*config.MemberRule{{Name: ".DoSomethingA", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: ".DoSomethingA", To: "DoSomethingA_New"}}}}},
					},
					{
						Name:    "ext1.TypeB",
						RuleSet: config.RuleSet{},
						Kind:    "struct",
						Pattern: "copy",
						Fields:  []*config.MemberRule{{Name: ".FieldB", RuleSet: config.RuleSet{}}},
					},
					{
						Name:    "ext1.TypeC",
						RuleSet: config.RuleSet{},
						Kind:    "struct",
						Pattern: "alias",
					},
					{
						Name:    "ext1.TypeD",
						RuleSet: config.RuleSet{},
						Kind:    "struct", // Changed from "type" to "struct"
						Pattern: "define",
					},
					{
						Name:    "ctx3.ContextType",
						RuleSet: config.RuleSet{},
						Kind:    "type",
						Methods: []*config.MemberRule{{Name: ".DoSomethingCtx", RuleSet: config.RuleSet{}}},
					},
					{
						Name:    "nested4.NestedType",
						RuleSet: config.RuleSet{},
						Kind:    "type",
						Pattern: "copy",
						Fields:  []*config.MemberRule{{Name: ".NestedField", RuleSet: config.RuleSet{}}},
					},
					{
						Name:    "ctx3.AfterNestedType",
						RuleSet: config.RuleSet{},
						Kind:    "type",
						Methods: []*config.MemberRule{{Name: ".DoSomethingAfterNested", RuleSet: config.RuleSet{}}},
					},
					{
						Name:    "github.com/another/pkg/v2.AnotherExternalType",
						RuleSet: config.RuleSet{},
						Kind:    "type",
						Methods: []*config.MemberRule{{Name: ".DoAnother", RuleSet: config.RuleSet{}}},
					},
				},
				Functions: []*config.FuncRule{
					{
						Name: "ext1.MyExternalFunction",
						RuleSet: config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalFunction", To: "MyNewFunction"}},
						},
					},
				},
				Variables: []*config.VarRule{
					{
						Name: "ext1.MyExternalVariable",
						RuleSet: config.RuleSet{
							Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalVariable", To: "MyNewVariable"}},
						},
					},
				},
				Constants: []*config.ConstRule{
					{
						Name:    "ext1.MyExternalConstant",
						RuleSet: config.RuleSet{Ignore: []string{"ext1.MyExternalConstant"}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, fset, err := loader.LoadGoFile(tt.filePath)
			if err != nil {
				t.Fatalf("Failed to load Go file %s: %v", tt.filePath, err)
			}

			cfg, err := parser.ParseFileDirectives(file, fset)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to parse directives: %v", err)
			}

			// --- Granular Comparison ---
			// Compare Types
			if len(cfg.Types) != len(tt.expectedConfig.Types) {
				t.Errorf("Types count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Types), len(cfg.Types))
			} else {
				for i := range cfg.Types {
					if !reflect.DeepEqual(*cfg.Types[i], *tt.expectedConfig.Types[i]) {
						t.Errorf("Type rule at index %d mismatch.\nExpected: %+v\nActual:   %+v", i,
							*tt.expectedConfig.Types[i], *cfg.Types[i])
					}
				}
			}

			// Compare Functions
			if len(cfg.Functions) != len(tt.expectedConfig.Functions) {
				t.Errorf("Functions count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Functions), len(cfg.Functions))
			} else {
				for i := range cfg.Functions {
					if !reflect.DeepEqual(*cfg.Functions[i], *tt.expectedConfig.Functions[i]) {
						t.Errorf("Function rule at index %d mismatch.\nExpected: %+v\nActual:   %+v", i, *tt.expectedConfig.Functions[i], *cfg.Functions[i])
					}
				}
			}

			// Compare Variables
			if len(cfg.Variables) != len(tt.expectedConfig.Variables) {
				t.Errorf("Variables count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Variables), len(cfg.Variables))
			} else {
				for i := range cfg.Variables {
					if !reflect.DeepEqual(*cfg.Variables[i], *tt.expectedConfig.Variables[i]) {
						t.Errorf("Variable rule at index %d mismatch.\nExpected: %+v\nActual:   %+v", i, *tt.expectedConfig.Variables[i], *cfg.Variables[i])
					}
				}
			}

			// Compare Constants
			if len(cfg.Constants) != len(tt.expectedConfig.Constants) {
				t.Errorf("Constants count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Constants), len(cfg.Constants))
			} else {
				for i := range cfg.Constants {
					if !reflect.DeepEqual(*cfg.Constants[i], *tt.expectedConfig.Constants[i]) {
						t.Errorf("Constant rule at index %d mismatch.\nExpected: %+v\nActual:   %+v", i, *tt.expectedConfig.Constants[i], *cfg.Constants[i])
					}
				}
			}

			// Compare Defaults (if not nil)
			if cfg.Defaults != nil || tt.expectedConfig.Defaults != nil {
				if !reflect.DeepEqual(cfg.Defaults, tt.expectedConfig.Defaults) {
					t.Errorf("Defaults mismatch.\nExpected: %+v\nActual:   %+v", tt.expectedConfig.Defaults, cfg.Defaults)
				}
			}

			// Compare Vars (if not nil)
			if cfg.Vars != nil || tt.expectedConfig.Vars != nil {
				if len(cfg.Vars) != len(tt.expectedConfig.Vars) {
					t.Errorf("Vars count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Vars), len(cfg.Vars))
				} else {
					for i := range cfg.Vars {
						if !reflect.DeepEqual(*cfg.Vars[i], *tt.expectedConfig.Vars[i]) {
							t.Errorf("Var entry at index %d mismatch.\nExpected: %+v\nActual:   %+v", i, *tt.expectedConfig.Vars[i], *cfg.Vars[i])
						}
					}
				}
			}

			// Compare Packages (if not nil)
			if cfg.Packages != nil || tt.expectedConfig.Packages != nil {
				if len(cfg.Packages) != len(tt.expectedConfig.Packages) {
					t.Errorf("Packages count mismatch. Expected: %d, Actual: %d", len(tt.expectedConfig.Packages), len(cfg.Packages))
				} else {
					for i := range cfg.Packages {
						if !reflect.DeepEqual(*cfg.Packages[i], *tt.expectedConfig.Packages[i]) {
							t.Errorf("Package at index %d mismatch.\nExpected: %+v\nActual:   %+v", i, *tt.expectedConfig.Packages[i], *cfg.Packages[i])
						}
					}
				}
			}
		})
	}
}
