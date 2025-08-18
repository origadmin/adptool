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
		t.Fatalf("Failed to parse directives: %v", parseErrorLog(err))
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
		t.Fatalf("Failed to parse directives: %v", parseErrorLog(err))
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
	t.Logf("Show packages: %#v", cfg.Packages[0])
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

	// According to the new rule, all standalone `type` directives are global.
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
		// These types are defined standalone and will be parsed as global.
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
	}

	assert.Equal(t, len(expectedGlobalTypes), len(cfg.Types), "Global types count mismatch")

	// A simple name check for global types for now.
	for i, expected := range expectedGlobalTypes {
		if i >= len(cfg.Types) {
			t.Fatalf("Missing expected global type #%d: %s", i, expected.Name)
		}
		assert.Equal(t, expected.Name, cfg.Types[i].Name, "Global type %d Name mismatch", i)
	}

	// Only types explicitly defined with 'package:type' syntax should be in packages.
	expectedPackageTypes := map[string][]*config.TypeRule{
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
