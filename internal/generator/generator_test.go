package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
		Constants: []*config.ConstRule{
			{
				Name:     "*",
				Disabled: false,
				RuleSet: config.RuleSet{
					Prefix: "Const",
				},
			},
		},
		Types: []*config.TypeRule{
			{
				Name:     "*",
				Disabled: false,
				RuleSet: config.RuleSet{
					Prefix: "Type",
				},
			},
		},
		Variables: []*config.VarRule{
			{
				Name:     "*",
				Disabled: false,
				RuleSet: config.RuleSet{
					Prefix: "Var",
				},
			},
		},
		Functions: []*config.FuncRule{
			{
				Name:     "*",
				Disabled: false,
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
						Name:     "*",
						Disabled: false,
						RuleSet: config.RuleSet{
							Suffix: "Source",
						},
					},
				},
			},
		},
	}

	// Compile the config using the compiler package
	compiledCfg, err := compiler.Compile(cfg)
	if err != nil {
		t.Fatalf("Failed to compile config: %v", err)
	}

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
	// 禁用自动格式化，因为测试中会手动调用goimports
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).WithFormatCode(false)
	err = generator.Generate(packageInfos)
	assert.NoError(t, err)

	// 5. Run goimports on the generated file first to clean up imports and format
	err = util.RunGoImports(outputFilePath)
	assert.NoError(t, err, "util.RunGoImports failed for %s", outputFilePath)

	// 6. Then run go vet on the formatted file
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, err := vetCmd.CombinedOutput()
	if err != nil {
		t.Errorf("go vet failed for %s: %s", outputFilePath, string(vetOutput))
	}

	// Read and verify the generated file content
	generatedContent, err := os.ReadFile(outputFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, generatedContent, "Generated file content is empty after goimports")

	// 输出生成的代码内容用于调试
	t.Logf("Generated code content:\n%s", string(generatedContent))

	// Verify basic expected content
	content := string(generatedContent)
	assert.Contains(t, content, "package aliaspkg", "Missing package declaration")
	assert.Contains(t, content, "import (", "Missing imports section")
	assert.Contains(t, content, `"github.com/origadmin/adptool/testdata/sourcepkg"`, "Missing sourcepkg import")
	assert.Contains(t, content, `"github.com/origadmin/adptool/testdata/sourcepkg2"`, "Missing sourcepkg2 import")

	// Verify renamed types from source package
	assert.Contains(t, content, "	TypeMyStruct           = source.MyStruct", "Missing renamed type TypeMyStruct")
	assert.Contains(t, content, "	TypeExportedType       = source.ExportedType", "Missing renamed type TypeExportedType")
	assert.Contains(t, content, "	TypeExportedInterface  = source.ExportedInterface", "Missing renamed type TypeExportedInterface")

	// Verify renamed types from source2 package
	assert.Contains(t, content, "	ComplexInterfaceSource = source2.ComplexInterface", "Missing renamed type ComplexInterfaceSource")
	assert.Contains(t, content, "	InputDataSource        = source2.InputData", "Missing renamed type InputDataSource")
	assert.Contains(t, content, "	OutputDataSource       = source2.OutputData", "Missing renamed type OutputDataSource")
	assert.Contains(t, content, "	WorkerSource           = source2.Worker", "Missing renamed type WorkerSource")

	// Clean up - don't remove the generated file!!!
	//err = os.Remove(outputFilePath)
	//assert.NoError(t, err)
}
