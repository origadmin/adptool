package generator

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/origadmin/adptool/internal/config"
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
			},
		},
		Replacer: replacer,
	}

	// 3. Define the output path
	outputFilePath := "../../output_dir/generated_test.go"

	// 4. Call the new Generate function
	err := Generate(compiledCfg, outputFilePath)
	assert.NoError(t, err)

	// Here we would ideally parse the generated file and check its contents.
	// For now, we are just checking if the function runs without errors
	// and manually inspecting the printed output.
}
