package parser_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

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

	assert.Equal(t, expectedDefaults, cfg.Defaults, "Defaults mismatch")
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

	assert.Equal(t, expectedProps, cfg.Props, "Props mismatch")
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

	// 比较类型数量
	assert.Equal(t, len(expectedTypes), len(cfg.Types), "Types count mismatch")

	// 逐个比较每个类型
	for i := range expectedTypes {
		expected := expectedTypes[i]
		actual := cfg.Types[i]

		assert.Equal(t, expected.Name, actual.Name, "Type %d Name mismatch", i)
		assert.Equal(t, expected.Kind, actual.Kind, "Type %d Kind mismatch", i)
		assert.Equal(t, expected.Pattern, actual.Pattern, "Type %d Pattern mismatch", i)

		// 比较 RuleSet
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Type %d RuleSet.Explicit count mismatch", i)
		assert.Equal(t, len(expected.RuleSet.Ignores), len(actual.RuleSet.Ignores), "Type %d RuleSet.Ignores count mismatch", i)

		// 比较 Methods
		assert.Equal(t, len(expected.Methods), len(actual.Methods), "Type %d Methods count mismatch", i)
		for j := range expected.Methods {
			if j < len(actual.Methods) {
				assert.Equal(t, expected.Methods[j].Name, actual.Methods[j].Name, "Type %d Method %d Name mismatch", i, j)
			}
		}

		// 比较 Fields
		assert.Equal(t, len(expected.Fields), len(actual.Fields), "Type %d Fields count mismatch", i)
		for j := range expected.Fields {
			if j < len(actual.Fields) {
				assert.Equal(t, expected.Fields[j].Name, actual.Fields[j].Name, "Type %d Field %d Name mismatch", i, j)
			}
		}
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

	// 比较函数数量
	assert.Equal(t, len(expectedFunctions), len(cfg.Functions), "Functions count mismatch")

	// 逐个比较每个函数
	for i := range expectedFunctions {
		expected := expectedFunctions[i]
		actual := cfg.Functions[i]

		assert.Equal(t, expected.Name, actual.Name, "Function %d Name mismatch", i)

		// 比较 RuleSet
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Function %d RuleSet.Explicit count mismatch", i)
		for j := range expected.RuleSet.Explicit {
			if j < len(actual.RuleSet.Explicit) {
				assert.Equal(t, expected.RuleSet.Explicit[j].From, actual.RuleSet.Explicit[j].From, "Function %d Explicit %d From mismatch", i, j)
				assert.Equal(t, expected.RuleSet.Explicit[j].To, actual.RuleSet.Explicit[j].To, "Function %d Explicit %d To mismatch", i, j)
			}
		}

		assert.Equal(t, len(expected.RuleSet.Ignores), len(actual.RuleSet.Ignores), "Function %d RuleSet.Ignores count mismatch", i)
		for j := range expected.RuleSet.Ignores {
			if j < len(actual.RuleSet.Ignores) {
				assert.Equal(t, expected.RuleSet.Ignores[j], actual.RuleSet.Ignores[j], "Function %d Ignore %d mismatch", i, j)
			}
		}
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

	// 比较变量数量
	assert.Equal(t, len(expectedVariables), len(cfg.Variables), "Variables count mismatch")

	// 逐个比较每个变量
	for i := range expectedVariables {
		expected := expectedVariables[i]
		actual := cfg.Variables[i]

		assert.Equal(t, expected.Name, actual.Name, "Variable %d Name mismatch", i)

		// 比较 RuleSet
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Variable %d RuleSet.Explicit count mismatch", i)
		for j := range expected.RuleSet.Explicit {
			if j < len(actual.RuleSet.Explicit) {
				assert.Equal(t, expected.RuleSet.Explicit[j].From, actual.RuleSet.Explicit[j].From, "Variable %d Explicit %d From mismatch", i, j)
				assert.Equal(t, expected.RuleSet.Explicit[j].To, actual.RuleSet.Explicit[j].To, "Variable %d Explicit %d To mismatch", i, j)
			}
		}

		assert.Equal(t, len(expected.RuleSet.Ignores), len(actual.RuleSet.Ignores), "Variable %d RuleSet.Ignores count mismatch", i)
		for j := range expected.RuleSet.Ignores {
			if j < len(actual.RuleSet.Ignores) {
				assert.Equal(t, expected.RuleSet.Ignores[j], actual.RuleSet.Ignores[j], "Variable %d Ignore %d mismatch", i, j)
			}
		}
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

	// 比较常量数量
	assert.Equal(t, len(expectedConstants), len(cfg.Constants), "Constants count mismatch")

	// 逐个比较每个常量
	for i := range expectedConstants {
		expected := expectedConstants[i]
		actual := cfg.Constants[i]

		assert.Equal(t, expected.Name, actual.Name, "Constant %d Name mismatch", i)

		// 比较 RuleSet
		assert.Equal(t, len(expected.RuleSet.Explicit), len(actual.RuleSet.Explicit), "Constant %d RuleSet.Explicit count mismatch", i)
		for j := range expected.RuleSet.Explicit {
			if j < len(actual.RuleSet.Explicit) {
				assert.Equal(t, expected.RuleSet.Explicit[j].From, actual.RuleSet.Explicit[j].From, "Constant %d Explicit %d From mismatch", i, j)
				assert.Equal(t, expected.RuleSet.Explicit[j].To, actual.RuleSet.Explicit[j].To, "Constant %d Explicit %d To mismatch", i, j)
			}
		}

		assert.Equal(t, len(expected.RuleSet.Ignores), len(actual.RuleSet.Ignores), "Constant %d RuleSet.Ignores count mismatch", i)
		for j := range expected.RuleSet.Ignores {
			if j < len(actual.RuleSet.Ignores) {
				assert.Equal(t, expected.RuleSet.Ignores[j], actual.RuleSet.Ignores[j], "Constant %d Ignore %d mismatch", i, j)
			}
		}
	}
}

func TestParseIgnores(t *testing.T) {
	filePath := filepath.Join(getModuleRoot(), "testdata", "parser_test_ignores.go")
	file, fset, err := loader.LoadGoFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load Go file %s: %v", filePath, err)
	}

	cfg, err := parser.ParseFileDirectives(file, fset)
	if err != nil {
		t.Fatalf("Failed to parse directives: %v", err)
	}

	// 预期的忽略模式
	expectedIgnores := []string{
		"pattern1", // 来自 ignore 指令
		"pattern2", // 来自 ignores 指令（逗号分隔）
		"pattern3", // 来自 ignores 指令（逗号分隔）
		"pattern4", // 来自 ignores:json 指令（JSON 数组）
		"pattern5", // 来自 ignores:json 指令（JSON 数组）
	}

	// 比较忽略模式数量
	assert.Equal(t, len(expectedIgnores), len(cfg.Ignores), "Ignores count mismatch")

	// 由于忽略模式的顺序可能不同，我们需要检查每个预期的模式是否都存在
	for _, expected := range expectedIgnores {
		found := false
		for _, actual := range cfg.Ignores {
			if expected == actual {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected ignore pattern %q not found", expected)
	}

	// 同样，检查是否有意外的模式
	for _, actual := range cfg.Ignores {
		found := false
		for _, expected := range expectedIgnores {
			if actual == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Unexpected ignore pattern %q found", actual)
	}
}
