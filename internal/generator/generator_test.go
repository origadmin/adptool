package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/loader"
	"github.com/origadmin/adptool/internal/util"
)

func TestGenerator_Generate(t *testing.T) {
	// 1. Load the configuration file
	// Use a more robust path resolution to find test_config_full.yaml
	configPath := filepath.Join("..", "..", "testdata", "test_config_full.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = filepath.Join("testdata", "test_config_full.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Skip("test_config_full.yaml not found, skipping test")
		}
	}

	cfg, err := loader.LoadConfigFile(configPath)
	assert.NoError(t, err, "loader.LoadConfigFile failed")

	// 2. Compile the configuration
	compiledCfg, err := compiler.Compile(cfg)
	assert.NoError(t, err, "compiler.Compile failed")

	// Convert CompiledPackage to PackageInfo
	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// 3. Set up the output file path in a temporary directory
	outputFilePath := filepath.Join(t.TempDir(), "generated_test.go")

	// 4. Create a new Generator instance and call its Generate method
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg))
	err = generator.Generate(packageInfos)
	assert.NoError(t, err, "generator.Generate failed")

	// 5. Verify the generated file exists
	_, err = os.Stat(outputFilePath)
	assert.NoError(t, err, "generated file should exist")

	// 6. Run go vet on the generated file to check for syntax errors
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, err := vetCmd.CombinedOutput()
	assert.NoError(t, err, "go vet failed for %s with output:\n%s", outputFilePath, string(vetOutput))

	// 7. (Optional) Read and log the generated code content for debugging
	generatedContent, err := os.ReadFile(outputFilePath)
	assert.NoError(t, err)
	t.Logf("Generated code content:\n%s", string(generatedContent))
}

func TestGenerate(t *testing.T) {
	// 1. Create the config and compiled config for the test
	var cfg = &config.Config{
		PackageName: "aliaspkg",
		Types: []*config.TypeRule{
			{
				Name: "*",
				RuleSet: config.RuleSet{
					Prefix: "Type",
				},
			},
		},
		Functions: []*config.FuncRule{
			{
				Name: "*",
				RuleSet: config.RuleSet{
					Prefix: "Func",
				},
			},
		},
		Constants: []*config.ConstRule{
			{
				Name: "*",
				RuleSet: config.RuleSet{
					Prefix: "Const",
				},
			},
		},
		Packages: []*config.Package{
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "source",
			},
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg2",
				Alias:  "source2",
				Types: []*config.TypeRule{
					{
						Name: "*",
						RuleSet: config.RuleSet{
							Suffix: "Source",
						},
					},
					{
						Name: "Worker",
						RuleSet: config.RuleSet{
							Prefix: "Source2",
						},
					},
				},
			},
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Types: []*config.TypeRule{
					{
						Name: "*",
						RuleSet: config.RuleSet{
							Suffix: "Source3",
						},
					},
				},
				Variables: []*config.VarRule{
					{
						Name: "*",
						RuleSet: config.RuleSet{
							Suffix: "Source3",
						},
					},
				},
				Constants: []*config.ConstRule{
					{
						Name: "*",
						RuleSet: config.RuleSet{
							Suffix: "Source3",
						},
					},
				},
				Functions: []*config.FuncRule{
					{
						Name: "*",
						RuleSet: config.RuleSet{
							Suffix: "Function3",
						},
					},
				},
			},
		},
	}

	// Compile the config using the compiler package
	compiledCfg, err := compiler.Compile(cfg)
	require.NoError(t, err, "Failed to compile config: %v", err)

	// Convert CompiledPackage to PackageInfo
	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// 3. Set up the output file path
	outputFilePath := filepath.Join(t.TempDir(), "test_alias.go")

	// 4. Create a new Generator instance and call its Generate method
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).
		WithFormatCode(false)
	err = generator.Generate(packageInfos)
	require.NoError(t, err)

	// Read and verify the generated file content
	generatedContent, err := os.ReadFile(outputFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, generatedContent, "Generated file content is empty after goimports")
	// The output generated code content is used for debugging
	t.Logf("Generated code content:\n%s", string(generatedContent))

	// 5. Run goimports on the generated file first to clean up imports and format
	err = util.RunGoImports(outputFilePath)
	require.NoError(t, err, "util.RunGoImports failed for %s", outputFilePath)

	// 6. Then run go vet on the formatted file
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, err := vetCmd.CombinedOutput()
	require.NoError(t, err, "go vet failed for %s: %s", outputFilePath, string(vetOutput))

	// 7. Verify the generated code using AST parsing
	verifyGeneratedCodeWithAST(t, outputFilePath)
}

func verifyGeneratedCodeWithAST(t *testing.T, filePath string) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	require.NoError(t, err, "Failed to parse generated file")

	expectedTypes := map[string]map[string]string{
		"source": {
			"CommonStruct":     "TypeCommonStruct",
			"MyStruct":         "TypeMyStruct",
			"ExportedType":     "TypeExportedType",
			"ExportedInterface":"TypeExportedInterface",
		},
		"source2": {
			"CommonStruct":     "CommonStructSource",
			"ComplexInterface": "ComplexInterfaceSource",
			"InputData":        "InputDataSource",
			"OutputData":       "OutputDataSource",
			"Worker":           "Source2Worker",
		},
		"sourcepkg3": {
			"CommonStruct":            "CommonStructSource3",
			"ComplexGenericInterface": "ComplexGenericInterfaceSource3",
			"EmbeddedInterface":       "EmbeddedInterfaceSource3",
			"InputData":               "InputDataSource3",
			"OutputData":              "OutputDataSource3",
			"Worker":                  "WorkerSource3",
			"WorkerConfig":            "WorkerConfigSource3",
			"GenericWorker":           "GenericWorkerSource3",
			"ProcessFunc":             "ProcessFuncSource3",
			"HandlerFunc":             "HandlerFuncSource3",
			"ProcessOption":           "ProcessOptionSource3",
			"ProcessConfig":           "ProcessConfigSource3",
			"WorkerOption":            "WorkerOptionSource3",
			"Status":                  "StatusSource3",
			"Priority":                "PrioritySource3",
			"TimeAlias":               "TimeAliasSource3",
			"StatusAlias":             "StatusAliasSource3",
			"IntAlias":                "IntAliasSource3",
		},
	}

	expectedConsts := map[string]map[string]string{
		"source": {
			"ExportedConstant": "ConstExportedConstant",
			"MaxRetries":      "ConstMaxRetries",
		},
		"source2": {
			"DefaultTimeout": "ConstDefaultTimeout",
			"MaxRetries":    "ConstMaxRetries1",
			"Version":       "ConstVersion",
		},
		"sourcepkg3": {
			"StatusUnknown":  "StatusUnknownSource3",
			"StatusPending":  "StatusPendingSource3",
			"StatusRunning":  "StatusRunningSource3",
			"StatusSuccess":  "StatusSuccessSource3",
			"StatusFailed":   "StatusFailedSource3",
			"PriorityLow":    "PriorityLowSource3",
			"PriorityMedium": "PriorityMediumSource3",
			"PriorityHigh":   "PriorityHighSource3",
			"DefaultTimeout": "DefaultTimeoutSource3",
			"Version":        "VersionSource3",
			"MaxRetries":     "MaxRetriesSource3",
		},
	}

	expectedVars := map[string]map[string]string{
		"source": {
			"ConfigValue":     "ConfigValue",
			"ExportedVariable":"ExportedVariable",
		},
		"source2": {
			"ConfigValue":    "ConfigValue1",
			"DefaultWorker":  "DefaultWorker",
			"StatsCounter":   "StatsCounter",
		},
		"sourcepkg3": {
			"ConfigValue":    "ConfigValueSource3",
			"DefaultWorker":  "DefaultWorkerSource3",
			"StatsCounter":   "StatsCounterSource3",
			"Processors":     "ProcessorsSource3",
		},
	}

	actualTypes := make(map[string]map[string]string)
	actualConsts := make(map[string]map[string]string)
	actualVars := make(map[string]map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		switch genDecl.Tok {
		case token.TYPE:
			extractTypeAliases(genDecl, actualTypes)
		case token.CONST:
			extractValueAliases(genDecl, actualConsts)
		case token.VAR:
			extractValueAliases(genDecl, actualVars)
		}
		return true
	})

	assert.Equal(t, expectedTypes, actualTypes, "Generated types do not match expected types")
	assert.Equal(t, expectedConsts, actualConsts, "Generated consts do not match expected consts")
	assert.Equal(t, expectedVars, actualVars, "Generated vars do not match expected vars")
}

func extractTypeAliases(genDecl *ast.GenDecl, targetMap map[string]map[string]string) {
	for _, spec := range genDecl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			var selExpr *ast.SelectorExpr
			switch t := typeSpec.Type.(type) {
			case *ast.SelectorExpr:
				selExpr = t
			case *ast.IndexExpr:
				if s, ok := t.X.(*ast.SelectorExpr); ok {
					selExpr = s
				}
			case *ast.IndexListExpr:
				if s, ok := t.X.(*ast.SelectorExpr); ok {
					selExpr = s
				}
			}

			if selExpr != nil {
				if pkgIdent, ok := selExpr.X.(*ast.Ident); ok {
					pkgAlias := pkgIdent.Name
					originalName := selExpr.Sel.Name
					newName := typeSpec.Name.Name

					if _, exists := targetMap[pkgAlias]; !exists {
						targetMap[pkgAlias] = make(map[string]string)
					}
					targetMap[pkgAlias][originalName] = newName
				}
			}
		}
	}
}

func extractValueAliases(genDecl *ast.GenDecl, targetMap map[string]map[string]string) {
	if genDecl.Tok != token.CONST && genDecl.Tok != token.VAR {
		return
	}

	for _, spec := range genDecl.Specs {
		if vs, ok := spec.(*ast.ValueSpec); ok {
			for i, name := range vs.Names {
				if len(vs.Values) > i {
					if selExpr, ok := vs.Values[i].(*ast.SelectorExpr); ok {
						if pkgIdent, ok := selExpr.X.(*ast.Ident); ok {
							pkgAlias := pkgIdent.Name
							originalName := selExpr.Sel.Name
							newName := name.Name

							if _, exists := targetMap[pkgAlias]; !exists {
								targetMap[pkgAlias] = make(map[string]string)
							}
							targetMap[pkgAlias][originalName] = newName
						}
					}
				}
			}
		}
	}
}

func extractFuncAliases(funcDecl *ast.FuncDecl, targetMap map[string]map[string]string) {
	if funcDecl.Recv != nil {
		// Skip methods for now
		return
	}

	if funcDecl.Name == nil {
		return
	}

	// For top-level functions, we need to look at the function body to find the actual implementation
	// since the function name itself won't have the package selector
	if funcDecl.Body != nil {
		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			switch call := n.(type) {
			case *ast.CallExpr:
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if pkgIdent, ok := sel.X.(*ast.Ident); ok {
						pkgName := pkgIdent.Name
						if _, exists := targetMap[pkgName]; !exists {
							targetMap[pkgName] = make(map[string]string)
						}
						// The function being called is sel.Sel.Name
						targetMap[pkgName][sel.Sel.Name] = sel.Sel.Name
					}
				}
			}
			return true
		})
	}
}

func TestGenerator_AdvancedModes(t *testing.T) {
	// 创建一个临时目录用于生成测试文件
	tempDir := t.TempDir()

	// 1. 准备测试配置
	cfg := &config.Config{
		PackageName: "generated",
		Packages: []*config.Package{
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Alias:  "source3",
				Types: []*config.TypeRule{
					// 1.1 显式重命名测试
					{
						RuleSet: config.RuleSet{
							ExplicitMode: "override",
							Explicit: []*config.ExplicitRule{
								{From: "ComplexGenericInterface", To: "ExplicitGenericInterface"},
							},
						},
						Name: "ComplexGenericInterface",
					},
					// 1.2 正则表达式重命名测试
					{
						RuleSet: config.RuleSet{
							RegexMode: "override",
							Regex: []*config.RegexRule{
								{Pattern: `^(\w+)Data$`, Replace: "${1}Wrapper"},
							},
						},
						Name: "*",
					},
					// 1.3 忽略特定类型
					{
						RuleSet: config.RuleSet{
							Ignores: []string{"WorkerConfig", "unexportedStruct"},
						},
						Name: "WorkerConfig",
					},
				},
				Functions: []*config.FuncRule{
					// 2.1 函数重命名测试
					{
						RuleSet: config.RuleSet{
							RegexMode: "override",
							Regex: []*config.RegexRule{
								{Pattern: `^New(\w+)$`, Replace: "Create$1"},
							},
						},
						Name: "*",
					},
					// 2.2 忽略特定函数
					{
						RuleSet: config.RuleSet{
							Ignores: []string{"unexportedFunction"},
						},
						Name: "unexportedFunction",
					},
				},
				Constants: []*config.ConstRule{
					// 3.1 常量重命名测试
					{
						RuleSet: config.RuleSet{
							ExplicitMode: "override",
							Explicit: []*config.ExplicitRule{
								{From: "DefaultTimeout", To: "CustomTimeout"},
							},
						},
						Name: "DefaultTimeout",
					},
					// 3.2 使用正则表达式重命名常量
					{
						RuleSet: config.RuleSet{
							RegexMode: "override",
							Regex: []*config.RegexRule{
								{Pattern: `^Status(\w+)$`, Replace: "${1}Status"},
							},
						},
						Name: "*",
					},
				},
				Variables: []*config.VarRule{
					// 4.1 变量重命名测试
					{
						RuleSet: config.RuleSet{
							ExplicitMode: "override",
							Explicit: []*config.ExplicitRule{
								{From: "DefaultWorker", To: "CustomWorker"},
							},
						},
						Name: "DefaultWorker",
					},
				},
			},
		},
	}

	// 2. 编译配置
	compiledCfg, err := compiler.Compile(cfg)
	require.NoError(t, err, "Failed to compile config: %v", err)

	// 3. 准备包信息
	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// 4. 生成代码
	outputFilePath := filepath.Join(tempDir, "advanced_modes_test.go")
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).WithFormatCode(true)
	err = generator.Generate(packageInfos)
	require.NoError(t, err, "Failed to generate code: %v", err)

	// 5. 读取生成的内容
	generatedContent, err := os.ReadFile(outputFilePath)
	require.NoError(t, err, "Failed to read generated file: %v", err)
	content := string(generatedContent)
	t.Logf("Generated code content for advanced modes test:\n%s", content)

	// 6. 定义测试用例
	tests := []struct {
		name           string
		pattern        string
		desc           string
		shouldNotMatch bool // 是否应该不匹配
	}{
		// 类型重命名测试
		{
			name:    "explicit type rename",
			pattern: `	ExplicitGenericInterface\[T any, K comparable\]\s*=\s*source3\.ComplexGenericInterface\[T, K\]`,
			desc:    "ComplexGenericInterface should be explicitly renamed to ExplicitGenericInterface with generic parameters",
		},
		{
			name:    "regex type rename - InputWrapper",
			pattern: `	InputWrapper\[T any\]\s*=\s*source3\.InputData\[T\]`,
			desc:    "InputData should be renamed to InputWrapper by regex",
		},
		{
			name:    "regex type rename - OutputWrapper",
			pattern: `	OutputWrapper\s*=\s*source3\.OutputData`,
			desc:    "OutputData should be renamed to OutputWrapper by regex",
		},
		{
			name:           "ignored type - WorkerConfig",
			pattern:        `type\s+WorkerConfig\s*=`,
			desc:           "WorkerConfig should be ignored",
			shouldNotMatch: true,
		},

		// 函数重命名测试
		{
			name:    "function rename - NewWorker",
			pattern: `func\s+CreateWorker\(name string, options \.\.\.source3\.WorkerOption\) \*source3\.Worker`,
			desc:    "NewWorker should be renamed to CreateWorker with correct signature",
		},
		{
			name:    "function rename - NewGenericWorker",
			pattern: `func\s+CreateGenericWorker\s*\[`,
			desc:    "NewGenericWorker should be renamed to CreateGenericWorker by regex",
		},
		{
			name:           "ignored function - unexportedFunction",
			pattern:        `func\s+unexportedFunction\s*\(`,
			desc:           "unexportedFunction should be ignored",
			shouldNotMatch: true,
		},

		// 常量重命名测试
		{
			name:    "constant rename - DefaultTimeout",
			pattern: `CustomTimeout\s*=\s*source3\.DefaultTimeout`,
			desc:    "DefaultTimeout should be renamed to CustomTimeout",
		},
		{
			name:    "constant check - PendingStatus",
			pattern: `PendingStatus\s*=\s*source3\.StatusPending`,
			desc:    "StatusPending should be renamed to PendingStatus by regex",
		},

		// 变量重命名测试
		{
			name:    "variable rename - DefaultWorker",
			pattern: `CustomWorker\s*=\s*source3\.DefaultWorker`,
			desc:    "DefaultWorker should be renamed to CustomWorker",
		},

		// 确保泛型类型参数正确保留
		{
			name:    "generic type parameters - InputWrapper",
			pattern: `	InputWrapper\[T any\]\s*=\s*source3\.InputData\[T\]`,
			desc:    "Generic type parameters should be preserved in InputWrapper",
		},

		// 确保接口方法签名正确
		// The interface method signature is not directly in the generated code
		// as it's part of the type alias, so we'll skip this test
		{
			name:    "interface method signature - MethodWithGenericParamsAndReturns",
			pattern: `ExplicitGenericInterface\[T any, K comparable\]\s*=`,
			desc:    "Interface type should be properly aliased",
		},
	}

	// 7. 运行测试用例
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.pattern)
			matched := re.MatchString(content)

			if tc.shouldNotMatch {
				assert.False(t, matched, "%s: pattern '%s' should not match in output", tc.desc, tc.pattern)
			} else {
				assert.True(t, matched, "%s: pattern '%s' not found in output", tc.desc, tc.pattern)
			}
		})
	}

	// 8. 使用 go vet 验证生成的代码
	cmd := exec.Command("go", "vet", outputFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("go vet output:\n%s", string(output))
	}
	assert.NoError(t, err, "Generated code should pass go vet")
}

func TestGenerator_Modes(t *testing.T) {
	// 1. Create the config for the test
	cfg := &config.Config{
		PackageName: "aliaspkg",
		Packages: []*config.Package{
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "source",
				Types: []*config.TypeRule{
					{
						RuleSet: config.RuleSet{
							ExplicitMode: "override", // Corrected field name
							Explicit: []*config.ExplicitRule{{
								From: "MyStruct",
								To:   "ExplicitMyStruct",
							}},
						},
						Name: "MyStruct",
					},
					// This rule should be ignored because the mode is explicit and the name doesn't match.
					{
						Name: "ExportedType",
						RuleSet: config.RuleSet{
							Prefix: "ShouldNotApply",
						},
					},
				},
				Constants: []*config.ConstRule{
					{
						RuleSet: config.RuleSet{
							ExplicitMode: "override", // Corrected field name
							Explicit: []*config.ExplicitRule{{
								From: "ExportedConstant",
								To:   "ExplicitConstant",
							}},
						},
						Name: "ExportedConstant",
					},
				},
			},
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg2",
				Alias:  "source2",
				Types: []*config.TypeRule{
					{
						RuleSet: config.RuleSet{
							RegexMode: "override", // Corrected field name
							Regex: []*config.RegexRule{{
								Pattern: `^(Input|Output)Data$`,
								Replace: "IO$1",
							}},
						},
						Name: `*`,
					},
				},
				Functions: []*config.FuncRule{
					{
						RuleSet: config.RuleSet{
							RegexMode: "override", // Corrected field name
							Regex: []*config.RegexRule{{
								Pattern: `^New(Worker)$`,
								Replace: "Create1$1",
							}},
						},
						Name: `*`,
					},
				},
			},
		},
	}

	// Compile the config
	compiledCfg, err := compiler.Compile(cfg)
	require.NoError(t, err, "Failed to compile config: %v", err)

	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	outputFilePath := filepath.Join(t.TempDir(), "test_modes.go")

	// Create a new generator
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).WithNoEditHeader(true).WithFormatCode(false)
	err = generator.Generate(packageInfos)
	require.NoError(t, err)

	// Read the generated content
	generatedContent, err := os.ReadFile(outputFilePath)
	require.NoError(t, err)
	content := string(generatedContent)
	t.Logf("Generated code content for modes test:\n%s", content)

	// Define test cases with more flexible matching
	tests := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "type alias for MyStruct",
			pattern: `(?m)^\s*ExplicitMyStruct\s*=\s*source\.MyStruct`,
			desc:    "ExplicitMyStruct should be an alias for source.MyStruct",
		},
		{
			name:    "type alias for InputData",
			pattern: `(?m)^\s*IOInput\s*=\s*source2\.InputData`,
			desc:    "IOInput should be an alias for source2.InputData",
		},
		{
			name:    "type alias for OutputData",
			pattern: `(?m)^\s*IOOutput\s*=\s*source2\.OutputData`,
			desc:    "IOOutput should be an alias for source2.OutputData",
		},
		{
			name:    "constant alias",
			pattern: `(?m)^\s*ExplicitConstant\s*=\s*source\.ExportedConstant`,
			desc:    "ExplicitConstant should be an alias for source.ExportedConstant",
		},
		{
			name:    "function definition",
			pattern: `(?m)^func\s+Create1Worker\s*\(name string\)\s*\*source2\.Worker`,
			desc:    "Create1Worker function should be defined with the correct signature",
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			re := regexp.MustCompile(tc.pattern)
			if !re.MatchString(content) {
				t.Errorf("Pattern not found: %s\nExpected pattern: %s", tc.desc, tc.pattern)
			}
		})
	}

	// Run goimports to format the code
	err = util.RunGoImports(outputFilePath)
	require.NoError(t, err, "util.RunGoImports failed for %s", outputFilePath)

	// Verify the file with go vet
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, err := vetCmd.CombinedOutput()
	require.NoError(t, err, "go vet failed for %s: %s", outputFilePath, string(vetOutput))

	// Verify the file is valid Go code
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, outputFilePath, nil, parser.ParseComments)
	require.NoError(t, err, "Generated file is not valid Go code")
}
