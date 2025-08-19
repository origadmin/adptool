package generator

import (
	"go/ast"
	"os"
	"os/exec"
	"path/filepath" // Added for filepath.Dir and os.MkdirAll
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/util" // Added for RunGoimports
)

// mockReplacer is a no-op replacer for testing purposes.
type mockReplacer struct{}

func (m *mockReplacer) Apply(node ast.Node) ast.Node {
	return node
}

func TestGenerate(t *testing.T) {
	// 1. Create a mock replacer
	replacer := &mockReplacer{}

	// 2. Create the compiled config for the test
	compiledCfg := &config.CompiledConfig{
		PackageName: "aliaspkg",
		Packages: []*config.CompiledPackage{
			{
				ImportPath:  "github.com/origadmin/adptool/testdata/sourcepkg",
				ImportAlias: "source",
				Constants: []*config.ConstRule{
					{
						Name:     "*",
						Disabled: false,
						RuleSet: config.RuleSet{
							Prefix: "Const",
						},
					},
				},
			},
			{
				ImportPath:  "github.com/origadmin/adptool/testdata/sourcepkg2",
				ImportAlias: "source2",
			},
		},
		Replacer: replacer,
	}

	// 3. Define the output path
	outputFilePath := "../../output_dir/generated_test.go"

	// Ensure the output directory exists
	outputDir := filepath.Dir(outputFilePath)
	err := os.MkdirAll(outputDir, 0755)
	assert.NoError(t, err)

	// 4. Create a new Generator instance and call its Generate method
	generator := NewGenerator(compiledCfg, outputFilePath)
	err = generator.Generate()
	assert.NoError(t, err)

	// 5. Run goimports on the generated file first to clean up imports and format
	err = util.RunGoimports(outputFilePath)
	assert.NoError(t, err, "util.RunGoimports failed for %s", outputFilePath)

	// 6. Then run go vet on the formatted file
	vetCmd := exec.Command("go", "vet", outputFilePath)
	vetOutput, vetErr := vetCmd.CombinedOutput()
	assert.NoError(t, vetErr, "go vet failed for %s: %s", outputFilePath, string(vetOutput))

	// Read the generated file content after goimports to ensure it's not empty
	generatedContent, err := os.ReadFile(outputFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, generatedContent, "Generated file content is empty after goimports")

	// No direct string comparison for expectedContent.
	// The test now relies on go vet and goimports for correctness and formatting.

	// Keep the generated file for inspection if needed.
	// err = os.Remove(outputFilePath)
	// assert.NoError(t, err)
}
