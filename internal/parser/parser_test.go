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

func TestParseDefaults(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_defaults.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedDefaults := &config.Defaults{
		Mode: &config.Mode{
			Strategy: "replace",
			Prefix:   "append",
			Suffix:   "append",
			Explicit: "merge",
			Regex:    "merge",
			Ignores:  "merge",
		},
	}

	if !reflect.DeepEqual(cfg.Defaults, expectedDefaults) {
		t.Errorf("Defaults mismatch.\nExpected: %+v\nActual:   %+v", expectedDefaults, cfg.Defaults)
	}
}

func TestParseProps(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_props.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedProps := []*config.PropsEntry{
		{
			Name:  "GlobalVar1",
			Value: "globalValue1",
		},
		{
			Name:  "GlobalVar2",
			Value: "globalValue2",
		},
	}

	if !reflect.DeepEqual(cfg.Props, expectedProps) {
		t.Errorf("Props mismatch.\nExpected: %+v\nActual:   %+v", expectedProps, cfg.Props)
	}
}

func TestParsePackages(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_packages.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedPackages := []*config.Package{
		{
			Import: "github.com/my/package/v1",
			Alias:  "mypkg",
			Path:   "./vendor/my/package/v1",
			Props: []*config.PropsEntry{
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
	}

	if len(cfg.Packages) != len(expectedPackages) {
		t.Errorf("Packages count mismatch. Expected: %d, Actual: %d", len(expectedPackages), len(cfg.Packages))
		return
	}

	for i := range expectedPackages {
		expected := expectedPackages[i]
		actual := cfg.Packages[i]

		if expected.Import != actual.Import {
			t.Errorf("Package %d Import mismatch. Expected: %s, Actual: %s", i, expected.Import, actual.Import)
		}
		if expected.Alias != actual.Alias {
			t.Errorf("Package %d Alias mismatch. Expected: %s, Actual: %s", i, expected.Alias, actual.Alias)
		}
		if expected.Path != actual.Path {
			t.Errorf("Package %d Path mismatch. Expected: %s, Actual: %s", i, expected.Path, actual.Path)
		}

		// Compare Props
		if len(actual.Props) != len(expected.Props) {
			t.Errorf("Package %d Props count mismatch. Expected: %d, Actual: %d", i, len(expected.Props), len(actual.Props))
		} else {
			for j := range expected.Props {
				if !reflect.DeepEqual(*actual.Props[j], *expected.Props[j]) {
					t.Errorf("Package %d Prop %d mismatch.\nExpected: %+v\nActual:   %+v", i, j, *expected.Props[j], *actual.Props[j])
				}
			}
		}

		// Compare Types
		if len(actual.Types) != len(expected.Types) {
			t.Errorf("Package %d Types count mismatch. Expected: %d, Actual: %d", i, len(expected.Types), len(actual.Types))
		} else {
			for j := range expected.Types {
				if !reflect.DeepEqual(*actual.Types[j], *expected.Types[j]) {
					t.Errorf("Package %d Type %d mismatch.\nExpected: %+v\nActual:   %+v", i, j, *expected.Types[j], *actual.Types[j])
				}
			}
		}

		// Compare Functions
		if len(actual.Functions) != len(expected.Functions) {
			t.Errorf("Package %d Functions count mismatch. Expected: %d, Actual: %d", i, len(expected.Functions), len(actual.Functions))
		} else {
			for j := range expected.Functions {
				if !reflect.DeepEqual(*actual.Functions[j], *expected.Functions[j]) {
					t.Errorf("Package %d Function %d mismatch.\nExpected: %+v\nActual:   %+v", i, j, *expected.Functions[j], *actual.Functions[j])
				}
			}
		}

		// Compare Variables
		if len(actual.Variables) != len(expected.Variables) {
			t.Errorf("Package %d Variables count mismatch. Expected: %d, Actual: %d", i, len(expected.Variables), len(actual.Variables))
		} else {
			for j := range expected.Variables {
				if !reflect.DeepEqual(*actual.Variables[j], *expected.Variables[j]) {
					t.Errorf("Package %d Variable %d mismatch.\nExpected: %+v\nActual:   %+v", i, j, *expected.Variables[j], *actual.Variables[j])
				}
			}
		}

		// Compare Constants
		if len(actual.Constants) != len(expected.Constants) {
			t.Errorf("Package %d Constants count mismatch. Expected: %d, Actual: %d", i, len(expected.Constants), len(actual.Constants))
		} else {
			for j := range expected.Constants {
				if !reflect.DeepEqual(*actual.Constants[j], *expected.Constants[j]) {
					t.Errorf("Package %d Constant %d mismatch.\nExpected: %+v\nActual:   %+v", i, j, *expected.Constants[j], *actual.Constants[j])
				}
			}
		}
	}
}

func TestParseTypes(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_types.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedTypes := []*config.TypeRule{
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
			Methods: []*config.MemberRule{
				{
					Name:    ".DoSomethingA",
					RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: ".DoSomethingA", To: "DoSomethingA_New"}}},
				},
			},
		},
		{
			Name:    "ext1.TypeB",
			RuleSet: config.RuleSet{},
			Kind:    "struct",
			Pattern: "copy",
			Fields: []*config.MemberRule{
				{
					Name:    ".FieldB",
					RuleSet: config.RuleSet{},
				},
			},
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
			Kind:    "struct",
			Pattern: "define",
		},
		{
			Name:    "ctx3.ContextType",
			RuleSet: config.RuleSet{},
			Kind:    "type",
			Methods: []*config.MemberRule{
				{
					Name:    ".DoSomethingCtx",
					RuleSet: config.RuleSet{},
				},
			},
		},
		{
			Name:    "nested4.NestedType",
			RuleSet: config.RuleSet{},
			Kind:    "struct",
			Pattern: "copy",
			Fields: []*config.MemberRule{
				{
					Name:    ".NestedField",
					RuleSet: config.RuleSet{},
				},
			},
		},
		{
			Name:    "ctx3.AfterNestedType",
			RuleSet: config.RuleSet{},
			Kind:    "type",
			Methods: []*config.MemberRule{
				{
					Name:    "DoSomethingAfterNested",
					RuleSet: config.RuleSet{},
				},
			},
		},
		{
			Name:    "github.com/another/pkg/v2.AnotherExternalType",
			RuleSet: config.RuleSet{},
			Kind:    "type",
			Methods: []*config.MemberRule{
				{
					Name:    "DoAnother",
					RuleSet: config.RuleSet{},
				},
			},
		},
	}

	if !reflect.DeepEqual(cfg.Types, expectedTypes) {
		t.Errorf("Types mismatch.\nExpected: %+v\nActual:   %+v", expectedTypes, cfg.Types)
	}
}

func TestParseFunctions(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_functions.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedFunctions := []*config.FuncRule{
		{
			Name: "ext1.MyExternalFunction",
			RuleSet: config.RuleSet{
				Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalFunction", To: "MyNewFunction"}},
			},
		},
	}

	if !reflect.DeepEqual(cfg.Functions, expectedFunctions) {
		t.Errorf("Functions mismatch.\nExpected: %+v\nActual:   %+v", expectedFunctions, cfg.Functions)
	}
}

func TestParseVariables(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_variables.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedVariables := []*config.VarRule{
		{
			Name: "ext1.MyExternalVariable",
			RuleSet: config.RuleSet{
				Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalVariable", To: "MyNewVariable"}},
			},
		},
	}

	if !reflect.DeepEqual(cfg.Variables, expectedVariables) {
		t.Errorf("Variables mismatch.\nExpected: %+v\nActual:   %+v", expectedVariables, cfg.Variables)
	}
}

func TestParseConstants(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_constants.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedConstants := []*config.ConstRule{
		{
			Name:    "ext1.MyExternalConstant",
			RuleSet: config.RuleSet{Ignores: []string{"ext1.MyExternalConstant"}},
		},
	}

	if !reflect.DeepEqual(cfg.Constants, expectedConstants) {
		t.Errorf("Constants mismatch.\nExpected: %+v\nActual:   %+v", expectedConstants, cfg.Constants)
	}
}
