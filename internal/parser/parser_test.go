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
							Name: "DoSomethingInPackage",
							RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{
								{From: "DoSomethingInPackage", To: "DoSomethingNewInPackage"}},
							},
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

	if !reflect.DeepEqual(cfg.Packages, expectedPackages) {
		t.Errorf("Packages mismatch.\nExpected: %+v\nActual:   %+v", expectedPackages, cfg.Packages)
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
					Name:    ".DoSomethingAfterNested",
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
					Name:    ".DoAnother",
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
