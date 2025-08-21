package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
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

	cfg, err := config.LoadConfig(configPath)
	assert.NoError(t, err, "config.LoadConfig failed")

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
		OutputPackageName: "aliaspkg",
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
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).WithFormatCode(false)
	err = generator.Generate(packageInfos)
	require.NoError(t, err)

	// 5. Run goimports on the generated file first to clean up imports and format
	err = util.RunGoImports(outputFilePath)
	require.NoError(t, err, "util.RunGoImports failed for %s", outputFilePath)

	// Read and verify the generated file content
	generatedContent, err := os.ReadFile(outputFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, generatedContent, "Generated file content is empty after goimports")

	// The output generated code content is used for debugging
	t.Logf("Generated code content:\n%s", string(generatedContent))

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
			"MyStruct":          "TypeMyStruct",
			"ExportedType":      "TypeExportedType",
			"ExportedInterface": "TypeExportedInterface",
		},
		"source2": {
			"ComplexInterface": "ComplexInterfaceSource",
			"InputData":        "InputDataSource",
			"OutputData":       "OutputDataSource",
			"Worker":           "Source2Worker",
		},
		"sourcepkg3": {
			"MaxRetries":              "MaxRetriesSource3",
			"Processors":              "ProcessorsSource3",
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

	actualTypes := make(map[string]map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
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

							if _, exists := actualTypes[pkgAlias]; !exists {
								actualTypes[pkgAlias] = make(map[string]string)
							}
							actualTypes[pkgAlias][originalName] = newName
						}
					}
				}
			}
		}
		return true
	})

	assert.Equal(t, expectedTypes, actualTypes, "Generated types do not match expected types")
}
