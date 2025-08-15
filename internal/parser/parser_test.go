package parser

import (
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
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
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
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

	assert.Equal(t, expectedDefaults, cfg.Defaults, "Defaults mismatch")
}

func TestParseProps(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_properties.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
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

	assert.Equal(t, expectedProps, cfg.Props, "Props mismatch")
}

func TestParsePackages(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_packages.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
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

	assert.Equal(t, len(expectedPackages), len(cfg.Packages), "Packages count mismatch")

	for i := range expectedPackages {
		expected := expectedPackages[i]
		actual := cfg.Packages[i]

		assert.Equal(t, expected.Import, actual.Import, "Package %d Import mismatch", i)
		assert.Equal(t, expected.Alias, actual.Alias, "Package %d Alias mismatch", i)
		assert.Equal(t, expected.Path, actual.Path, "Package %d Path mismatch", i)

		// Compare Props
		assert.Equal(t, len(expected.Props), len(actual.Props), "Package %d Props count mismatch", i)
		for j := range expected.Props {
			assert.Equal(t, *expected.Props[j], *actual.Props[j], "Package %d Prop %d mismatch", i, j)
		}

		// Compare Types
		assert.Equal(t, len(expected.Types), len(actual.Types), "Package %d Types count mismatch", i)
		for j := range expected.Types {
			assert.Equal(t, *expected.Types[j], *actual.Types[j], "Package %d Type %d mismatch", i, j)
		}

		// Compare Functions
		assert.Equal(t, len(expected.Functions), len(actual.Functions), "Package %d Functions count mismatch", i)
		for j := range expected.Functions {
			assert.Equal(t, *expected.Functions[j], *actual.Functions[j], "Package %d Function %d mismatch", i, j)
		}

		// Compare Variables
		assert.Equal(t, len(expected.Variables), len(actual.Variables), "Package %d Variables count mismatch", i)
		for j := range expected.Variables {
			assert.Equal(t, *expected.Variables[j], *actual.Variables[j], "Package %d Variable %d mismatch", i, j)
		}

		// Compare Constants
		assert.Equal(t, len(expected.Constants), len(actual.Constants), "Package %d Constants count mismatch", i)
		for j := range expected.Constants {
			assert.Equal(t, *expected.Constants[j], *actual.Constants[j], "Package %d Constant %d mismatch", i, j)
		}
	}
}

func TestParseTypes(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_types.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", parseErrorLog(err))
	}

	// This test now only checks for GLOBAL types. Types defined within a package
	// are tested separately below.
	expectedGlobalTypes := []*config.TypeRule{
		{
			Name:    "*",
			Kind:    "struct",
			Pattern: "wrap",
		},
		{
			Name: "ext1.TypeA",
			Kind: "type",
			Methods: []*config.MemberRule{
				{
					Name:    "DoSomethingA",
					RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "DoSomethingA", To: "DoSomethingA_New"}}},
				},
			},
		},
		{
			Name:    "ext1.TypeB",
			Kind:    "struct",
			Pattern: "copy",
			Fields: []*config.MemberRule{
				{
					Name: "FieldB",
				},
			},
		},
		{
			Name:    "ext1.TypeC",
			Kind:    "struct",
			Pattern: "alias",
		},
		{
			Name:    "ext1.TypeD",
			Kind:    "struct",
			Pattern: "define",
		},
		{
			Name: "github.com/another/pkg/v2.AnotherExternalType",
			Kind: "type",
			Methods: []*config.MemberRule{
				{
					Name: "DoAnother",
				},
			},
		},
	}

	assert.Equal(t, len(expectedGlobalTypes), len(cfg.Types), "Global types count mismatch")

	for i, expected := range expectedGlobalTypes {
		if i >= len(cfg.Types) {
			t.Fatalf("Missing expected global type #%d: %s", i, expected.Name)
		}
		actual := cfg.Types[i]
		assert.Equal(t, expected.Name, actual.Name, "Global type %d Name mismatch", i)
		assert.Equal(t, expected.Kind, actual.Kind, "Global type %d Kind mismatch", i)
		assert.Equal(t, expected.Pattern, actual.Pattern, "Global type %d Pattern mismatch", i)
		assert.Equal(t, len(expected.Methods), len(actual.Methods), "Global type %d Methods count mismatch", i)
		assert.Equal(t, len(expected.Fields), len(actual.Fields), "Global type %d Fields count mismatch", i)
	}

	// Additionally, let's verify the package-scoped types are where they belong.
	expectedPackageTypes := map[string][]*config.TypeRule{
		"github.com/context/pkg/v3": {
			{
				Name: "ctx3.ContextType",
				Kind: "type",
				Methods: []*config.MemberRule{
					{
						Name: "DoSomethingCtx",
					},
				},
			},
			{
				Name: "ctx3.AfterNestedType",
				Kind: "type",
				Methods: []*config.MemberRule{
					{
						Name: "DoSomethingAfterNested",
					},
				},
			},
		},
		"github.com/nested/pkg/v4": {
			{
				Name:    "nested4.NestedType",
				Kind:    "struct",
				Pattern: "copy",
				Fields: []*config.MemberRule{
					{
						Name: "NestedField",
					},
				},
			},
		},
	}

	actualPackageTypes := make(map[string][]*config.TypeRule)
	for _, pkg := range cfg.Packages {
		if len(pkg.Types) > 0 {
			actualPackageTypes[pkg.Import] = pkg.Types
		}
	}

	assert.Equal(t, len(expectedPackageTypes), len(actualPackageTypes), "Count of packages with types mismatch")
	for imp, expectedTypesInPkg := range expectedPackageTypes {
		actualTypesInPkg, ok := actualPackageTypes[imp]
		assert.True(t, ok, "Missing package with types: %s", imp)
		assert.Equal(t, len(expectedTypesInPkg), len(actualTypesInPkg), "Types count in package %s mismatch", imp)
		// A more detailed comparison could be done here if needed.
	}
}

func TestParseFunctions(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_functions.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", parseErrorLog(err))
	}

	expectedFunctions := []*config.FuncRule{
		{
			Name: "ext1.MyExternalFunction",
			RuleSet: config.RuleSet{
				Explicit: []*config.ExplicitRule{{From: "ext1.MyExternalFunction", To: "MyNewFunction"}},
			},
		},
	}

	assert.Equal(t, len(expectedFunctions), len(cfg.Functions), "Functions count mismatch")

	for i, expected := range expectedFunctions {
		if i >= len(cfg.Functions) {
			t.Fatalf("Missing expected function #%d: %s", i, expected.Name)
		}
		actual := cfg.Functions[i]
		assert.Equal(t, expected.Name, actual.Name, "Function %d Name mismatch", i)
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Function %d RuleSet.Explicit count mismatch", i)
	}
}

func TestParseVariables(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_variables.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
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

	assert.Equal(t, len(expectedVariables), len(cfg.Variables), "Variables count mismatch")

	for i, expected := range expectedVariables {
		if i >= len(cfg.Variables) {
			t.Fatalf("Missing expected variable #%d: %s", i, expected.Name)
		}
		actual := cfg.Variables[i]
		assert.Equal(t, expected.Name, actual.Name, "Variable %d Name mismatch", i)
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Variable %d RuleSet.Explicit count mismatch", i)
	}
}

func TestParseConstants(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_constants.go")
	file, fset, err := loadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	expectedConstants := []*config.ConstRule{
		{
			Name:    "ext1.MyExternalConstant",
			RuleSet: config.RuleSet{Ignores: []string{"ext1.MyExternalConstant"}},
		},
	}

	assert.Equal(t, len(expectedConstants), len(cfg.Constants), "Constants count mismatch")

	for i, expected := range expectedConstants {
		if i >= len(cfg.Constants) {
			t.Fatalf("Missing expected constant #%d: %s", i, expected.Name)
		}
		actual := cfg.Constants[i]
		assert.Equal(t, expected.Name, actual.Name, "Constant %d Name mismatch", i)
		assert.Equal(t, len(expected.RuleSet.Ignores), len(actual.RuleSet.Ignores), "Constant %d RuleSet.Ignores count mismatch", i)
	}
}

// parseErrorLog is a helper function to get a detailed error string for parser errors.
func parseErrorLog(err error) string {
	var pe *parserError
	if errors.As(err, &pe) {
		return pe.String() // Use the String() method for full details
	}
	return err.Error() // Fallback to standard Error() for other errors
}
