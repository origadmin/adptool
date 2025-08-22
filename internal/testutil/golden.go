package testutil

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/origadmin/adptool/internal/util"
	"github.com/pmezard/go-difflib/difflib"
)

// Update is a flag to update golden files.
var Update = flag.Bool("update", false, "update golden files")

// CompareWithGolden compares generated content with a golden file.
// It handles goimports formatting, diffing, and updating the golden file.
// testdataDir should be the path to the directory containing the golden files.
func CompareWithGolden(t *testing.T, testdataDir string, gotBytes []byte) {
	t.Helper()

	// Create a temporary file to run goimports on.
	tempFile, err := ioutil.TempFile(t.TempDir(), "*.go")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file

	if _, err := tempFile.Write(gotBytes); err != nil {
		t.Fatalf("failed to write to temporary file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("failed to close temporary file: %v", err)
	}

	// Run goimports on the temporary file.
	if err := util.RunGoImports(tempFile.Name()); err != nil {
		t.Fatalf("failed to format generated code with goimports: %v", err)
	}

	// Read the formatted content back from the temp file.
	formattedBytes, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("failed to read formatted temp file: %v", err)
	}

	// Determine the golden file path from the test name.
	goldenFile := filepath.Join(testdataDir, strings.ReplaceAll(t.Name(), "/", "_")+ ".golden")

	// If the -update flag is set, write the formatted content to the golden file.
	if *Update {
		if err := os.MkdirAll(filepath.Dir(goldenFile), 0755); err != nil {
			t.Fatalf("failed to create directory for golden file: %v", err)
		}
		if err := ioutil.WriteFile(goldenFile, formattedBytes, 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return
	}

	// Read the golden file.
	wantBytes, err := ioutil.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	// Compare the formatted generated content with the golden file content.
	if !bytes.Equal(formattedBytes, wantBytes) {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(wantBytes)),
			B:        difflib.SplitLines(string(formattedBytes)),
			FromFile: "golden:" + goldenFile,
			ToFile:   "got",
			Context:  3,
		}
		diffStr, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			t.Fatalf("failed to generate diff: %v", err)
		}
		t.Errorf("generated output does not match golden file (-golden +got):\n%s", diffStr)
	}
}
