package generator

import (
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

func TestGenerateVariadic(t *testing.T) {
	// 1. Create the config and compiled config for the test
	var cfg = &config.Config{
		OutputPackageName: "aliaspkg",
		Packages: []*config.Package{
			{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Functions: []*config.FuncRule{
					{
						Name: "NewWorker",
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
	generator := NewGenerator(compiledCfg.PackageName, outputFilePath, compiler.NewReplacer(compiledCfg)).WithFormatCode(false)
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

	// Verify the generated function call
	assert.Contains(t, string(generatedContent), "return sourcepkg3.NewWorker(name, options...)", "Generated code should pass variadic args correctly")
}
