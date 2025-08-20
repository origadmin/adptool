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
	cfg, err := config.LoadConfig("../testdata/test_config_full.yaml")
	assert.NoError(t, err, "config.LoadConfig failed")

	// 2. Compile the configuration
	compiledCfg, err := compiler.Compile(cfg)
	assert.NoError(t, err, "compiler.Compile failed")

	// 3. Create a mock replacer
	replacer := compiler.NewReplacer(compiledCfg)

	// Convert CompiledPackage to PackageInfo
	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// Ensure the output directory exists
	outputDir := "../output_dir"
	err = os.MkdirAll(outputDir, 0755)
	assert.NoError(t, err)

	outputFilePath := filepath.Join(outputDir, "generated_test.go")

	// 4. Create a new Generator instance and call its Generate method
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, replacer)
	err = generator.Generate(packageInfos)
	assert.NoError(t, err)

	// 5. Run goimports on the generated file first to clean up imports and format
	err = util.RunGoImports(outputFilePath)
	assert.NoError(t, err, "util.RunGoImports failed for %s", outputFilePath)

	// 6. Then run go vet on the formatted file
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, err := vetCmd.CombinedOutput()
	assert.NoError(t, err, "go vet failed for %s with output:\n%s", outputFilePath, string(vetOutput))

	// 7. Finally, try to build the file to make sure it's valid Go code
	buildCmd := exec.Command("go", "build", outputFilePath)
	buildOutput, err := buildCmd.CombinedOutput()
	assert.NoError(t, err, "go build failed for %s with output:\n%s", outputFilePath, string(buildOutput))
}

func TestGenerate(t *testing.T) {
	// 1. Create a mock replacer, removed for now. use the compiler package instead

	// 2. Create the config and compiled config for the test
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
			},
		},
	} // Compile the config using the compiler package
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
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg))
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

	// Clean up - don't remove the generated file!!!
	//err = os.Remove(outputFilePath)
	//assert.NoError(t, err)
}
