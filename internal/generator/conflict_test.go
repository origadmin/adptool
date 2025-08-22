package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerator_NameConflicts(t *testing.T) {
	// Create a temporary directory for our test output
	tempDir, err := os.MkdirTemp("", "generator_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "conflict_test.go")

	// Create package info for both packages
	packageInfos := []*PackageInfo{
		{
			ImportPath:  "github.com/origadmin/adptool/testdata/sourcepkg",
			ImportAlias: "source1",
		},
		{
			ImportPath:  "github.com/origadmin/adptool/testdata/sourcepkg2",
			ImportAlias: "source2",
		},
		{
			ImportPath:  "github.com/origadmin/adptool/testdata/sourcepkg3",
			ImportAlias: "source3",
		},
	}

	// Create the generator with default settings
	gen := NewGenerator("conflicttest", outputFile, nil)

	// Generate the code
	err = gen.Generate(packageInfos)
	if err != nil {
		t.Fatalf("Failed to generate code: %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// For now, just check that the file was generated successfully
	// In a real implementation, we would check for the renamed identifiers
	t.Logf("Generated file content:\n%s", string(content))

	// 检查是否正确处理了名称冲突
	output := string(content)
	t.Logf("Generated file content:\n%s", output)

	// 应该包含重命名后的元素
	if !(containsString(output, "MaxRetries") && containsString(output, "MaxRetries1")) {
		t.Errorf("Expected both MaxRetries and MaxRetries1 in output, got:\n%s", output)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
