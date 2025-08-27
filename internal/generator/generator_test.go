package generator

import (
	"bytes"
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/testutil"
)

var update = flag.Bool("update", false, "update golden files")

// runGoldenTest is a helper function to run a generator test case with a given config.
func runGoldenTest(t *testing.T, cfg *config.Config) {
	t.Helper()

	// Compile the configuration
	compiledCfg, err := compiler.Compile(cfg)
	require.NoError(t, err)

	// Convert CompiledPackage to PackageInfo
	var packageInfos []*PackageInfo
	for _, pkg := range compiledCfg.Packages {
		packageInfos = append(packageInfos, &PackageInfo{
			ImportPath:  pkg.ImportPath,
			ImportAlias: pkg.ImportAlias,
		})
	}

	// Create a new Generator, redirecting output to an in-memory buffer
	outputBuffer := &bytes.Buffer{}
	generator := NewGenerator(compiledCfg.PackageName, "", compiler.NewReplacer(compiledCfg), "")
	generator.builder.writer = outputBuffer

	// Generate the code
	err = generator.Generate(packageInfos)
	require.NoError(t, err)

	// Compare the generated code with the golden file
	testdataPath := filepath.Join("..", "..", "testdata", "generator")
	testutil.CompareWithGolden(t, testdataPath, *update, outputBuffer.Bytes())
}

// TestGenerator holds all granular test cases for the generator.
func TestGenerator(t *testing.T) {
	// Test case for simple prefix renaming
	t.Run("TestPrefix_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "prefixtest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "source",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
			}},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for explicit renaming with override mode
	t.Run("TestExplicit_Override", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "explicittest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "source",
				Types: []*config.TypeRule{{
					Name: "MyStruct",
					RuleSet: config.RuleSet{
						ExplicitMode: "override",
						Explicit:     []*config.ExplicitRule{{From: "MyStruct", To: "MyCustomStruct"}},
					},
				}},
			}},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for regex renaming
	t.Run("TestRegex_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "regextest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg2",
				Alias:  "source2",
				Types: []*config.TypeRule{{
					Name: "*",
					RuleSet: config.RuleSet{
						RegexMode: "override",
						Regex:     []*config.RegexRule{{Pattern: `^(Input|Output)Data$`, Replace: "IO$1"}},
					},
				}},
			}},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for ignoring specific identifiers
	t.Run("TestIgnores", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "ignoretest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Alias:  "source3",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Ignores: []string{"WorkerConfig", "unexportedStruct"}}}},
			}},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for naming conflicts between constants
	t.Run("TestConflict_Constants", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "conflicttest",
			Packages: []*config.Package{
				{Import: "github.com/origadmin/adptool/testdata/sourcepkg", Alias: "source1"},
				{Import: "github.com/origadmin/adptool/testdata/sourcepkg2", Alias: "source2"},
			},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for handling generic types
	t.Run("TestGenerics_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "generictest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Alias:  "source3",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
			}},
		}
		runGoldenTest(t, cfg)
	})

	// Test case for all bug fixes combined
	t.Run("TestBugFixes", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "bugfixestest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/bugfixes",
				Alias:  "source",
			}},
		}
		runGoldenTest(t, cfg)
	})
}
