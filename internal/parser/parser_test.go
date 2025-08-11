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
	// Path to the test data file
	testFilePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_file.go")

	// Load the Go file once for all tests
	file, fset, err := loader.LoadGoFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", testFilePath, err)
	}

	// --- Test Cases for each Config section ---

	// Test Defaults parsing
	t.Run("Defaults Parsing", func(t *testing.T) {
		expected := &config.Config{
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
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Defaults, expected.Defaults) {
			t.Errorf("Defaults mismatch.\nExpected: %+v\nActual:   %+v", expected.Defaults, actual.Defaults)
		}
	})

	// Test Vars parsing
	t.Run("Vars Parsing", func(t *testing.T) {
		expected := &config.Config{
			Vars: []*config.VarEntry{
				{Name: "GlobalVar1", Value: "globalValue1"},
				{Name: "GlobalVar2", Value: "globalValue2"},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Vars, expected.Vars) {
			t.Errorf("Vars mismatch.\nExpected: %+v\nActual:   %+v", expected.Vars, actual.Vars)
		}
	})

	// Test Packages parsing
	t.Run("Packages Parsing", func(t *testing.T) {
		expected := &config.Config{
			Packages: []*config.Package{
				{
					Import: "github.com/my/package/v1",
					Alias:  "mypkg",
					Path:   "./vendor/my/package/v1",
					Vars: []*config.VarEntry{
						{Name: "PackageVar1", Value: "packageValue1"},
					},
					Types: []*config.TypeRule{
						{Name: "MyStructInPackage", Kind: "type", Pattern: "wrap", Methods: []*config.MemberRule{{Name: "DoSomethingInPackage", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingInPackage", To: "DoSomethingNewInPackage"}}}}}, RuleSet: config.RuleSet{}},
					},
					Functions: []*config.FuncRule{
						{Name: "MyFuncInPackage", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "MyFuncInPackage", To: "MyNewFuncInPackage"}}}},
					},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Packages, expected.Packages) {
			t.Errorf("Packages mismatch.\nExpected: %+v\nActual:   %+v", expected.Packages, actual.Packages)
		}
	})

	// Test Types parsing
	t.Run("Types Parsing", func(t *testing.T) {
		expected := &config.Config{
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
					Methods: []*config.MemberRule{
						{Name: ".DoSomethingA", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: ".DoSomethingA", To: "DoSomethingA_New"}}}},
					},
				},
				{
					Name:    "ext1.TypeB",
					RuleSet: config.RuleSet{},
					Kind:    "struct",
					Pattern: "copy",
					Fields: []*config.MemberRule{
						{Name: ".FieldB", RuleSet: config.RuleSet{}},
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
					Kind:    "type",
					Pattern: "define",
				},
				{
					Name:    "ctx3.ContextType",
					RuleSet: config.RuleSet{},
					Kind:    "type",
					Methods: []*config.MemberRule{
						{Name: ".DoSomethingCtx", RuleSet: config.RuleSet{}},
					},
				},
				{
					Name:    "nested4.NestedType",
					RuleSet: config.RuleSet{},
					Kind:    "type",
					Pattern: "copy",
					Fields: []*config.MemberRule{
						{Name: ".NestedField", RuleSet: config.RuleSet{}},
					},
				},
				{
					Name:    "ctx3.AfterNestedType",
					RuleSet: config.RuleSet{},
					Kind:    "type",
					Methods: []*config.MemberRule{
						{Name: ".DoSomethingAfterNested", RuleSet: config.RuleSet{}},
					},
				},
				{
					Name:    "github.com/another/pkg/v2.AnotherExternalType",
					RuleSet: config.RuleSet{},
					Kind:    "type",
					Methods: []*config.MemberRule{
						{Name: ".DoAnother", RuleSet: config.RuleSet{}},
					},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Types, expected.Types) {
			t.Errorf("Types mismatch.\nExpected: %+v\nActual:   %+v", expected.Types, actual.Types)
		}
	})

	// Test Functions parsing
	t.Run("Functions Parsing", func(t *testing.T) {
		expected := &config.Config{
			Functions: []*config.FuncRule{
				{
					Name: "ext1.MyExternalFunction",
					RuleSet: config.RuleSet{
						Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalFunction", To: "MyNewFunction"}},
					},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Functions, expected.Functions) {
			t.Errorf("Functions mismatch.\nExpected: %+v\nActual:   %+v", expected.Functions, actual.Functions)
		}
	})

	// Test Variables parsing
	t.Run("Variables Parsing", func(t *testing.T) {
		expected := &config.Config{
			Variables: []*config.VarRule{
				{
					Name: "ext1.MyExternalVariable",
					RuleSet: config.RuleSet{
						Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalVariable", To: "MyNewVariable"}},
					},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Variables, expected.Variables) {
			t.Errorf("Variables mismatch.\nExpected: %+v\nActual:   %+v", expected.Variables, actual.Variables)
		}
	})

	// Test Constants parsing
	t.Run("Constants Parsing", func(t *testing.T) {
		expected := &config.Config{
			Constants: []*config.ConstRule{
				{
					Name:    "ext1.MyExternalConstant",
					RuleSet: config.RuleSet{Ignore: []string{"ext1.MyExternalConstant"}},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Constants, expected.Constants) {
			t.Errorf("Constants mismatch.\nExpected: %+v\nActual:   %+v", expected.Constants, actual.Constants)
		}
	})

	// Test Defaults parsing
	t.Run("Defaults Parsing", func(t *testing.T) {
		expected := &config.Config{
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
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Defaults, expected.Defaults) {
			t.Errorf("Defaults mismatch.\nExpected: %+v\nActual:   %+v", expected.Defaults, actual.Defaults)
		}
	})

	// Test Vars parsing
	t.Run("Vars Parsing", func(t *testing.T) {
		expected := &config.Config{
			Vars: []*config.VarEntry{
				{Name: "GlobalVar1", Value: "globalValue1"},
				{Name: "GlobalVar2", Value: "globalValue2"},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Vars, expected.Vars) {
			t.Errorf("Vars mismatch.\nExpected: %+v\nActual:   %+v", expected.Vars, actual.Vars)
		}
	})

	// Test Packages parsing
	t.Run("Packages Parsing", func(t *testing.T) {
		expected := &config.Config{
			Packages: []*config.Package{
				{
					Import: "github.com/my/package/v1",
					Alias:  "mypkg",
					Path:   "./vendor/my/package/v1",
					Vars: []*config.VarEntry{
						{Name: "PackageVar1", Value: "packageValue1"},
					},
					Types: []*config.TypeRule{
						{Name: "MyStructInPackage", Kind: "type", Pattern: "wrap", Methods: []*config.MemberRule{{Name: "DoSomethingInPackage", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingInPackage", To: "DoSomethingNewInPackage"}}}}}, RuleSet: config.RuleSet{}},
					},
					Functions: []*config.FuncRule{
						{Name: "MyFuncInPackage", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "MyFuncInPackage", To: "MyNewFuncInPackage"}}}},
					},
				},
			},
		}
		actual, err := parser.ParseFileDirectives(file, fset)
		if err != nil {
			t.Fatalf("Failed to parse directives: %v", err)
		}
		if !reflect.DeepEqual(actual.Packages, expected.Packages) {
			t.Errorf("Packages mismatch.\nExpected: %+v\nActual:   %+v", expected.Packages, actual.Packages)
		}
	})
}
