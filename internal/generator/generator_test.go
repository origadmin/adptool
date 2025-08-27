package generator

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/testutil"
)

var update = flag.Bool("update", false, "update golden files")

// TestIssues is a data-driven test that automatically discovers and runs
// regression tests for specific, documented issues.
// It works by scanning the `testdata/generator/issues` directory.
func TestIssues(t *testing.T) {
	issuesDir := filepath.Join("..", "..", "testdata", "generator", "issues")

	dirs, err := filepath.Glob(filepath.Join(issuesDir, "*"))
	require.NoError(t, err)

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		require.NoError(t, err)

		if !info.IsDir() {
			continue
		}

		testCaseName := filepath.Base(dir)

		// Dynamically create a sub-test for each issue directory.
		t.Run(testCaseName, func(t *testing.T) {
			importPath := "github.com/origadmin/adptool/testdata/generator/issues/" + testCaseName + "/source"
			goldenFilePath := filepath.Join(dir, "test.golden")

			cfg := &config.Config{
				PackageName: "test",
				Packages: []*config.Package{{
					Import: importPath,
					Alias:  "source",
				}},
			}

			compiledCfg, err := compiler.Compile(cfg)
			require.NoError(t, err)

			var packageInfos []*PackageInfo
			for _, pkg := range compiledCfg.Packages {
				packageInfos = append(packageInfos, &PackageInfo{
					ImportPath:  pkg.ImportPath,
					ImportAlias: pkg.ImportAlias,
				})
			}

			outputBuffer := &bytes.Buffer{}
			generator := NewGenerator(compiledCfg.PackageName, "", compiler.NewReplacer(compiledCfg), "")
			generator.builder.writer = outputBuffer

			err = generator.Generate(packageInfos)
			require.NoError(t, err)

			testutil.CompareWithGoldenFile(t, goldenFilePath, *update, outputBuffer.Bytes())
		})
	}
}

// TestGenerator_LegacyCases contains the original test suite for basic generator functionality.
// This ensures that existing, fundamental tests are preserved.
func TestGenerator_LegacyCases(t *testing.T) {
	// runLegacyGoldenTest is a helper specific to the legacy test structure.
	runLegacyGoldenTest := func(t *testing.T, cfg *config.Config) {
		t.Helper()

		compiledCfg, err := compiler.Compile(cfg)
		require.NoError(t, err)

		var packageInfos []*PackageInfo
		for _, pkg := range compiledCfg.Packages {
			packageInfos = append(packageInfos, &PackageInfo{
				ImportPath:  pkg.ImportPath,
				ImportAlias: pkg.ImportAlias,
			})
		}

		outputBuffer := &bytes.Buffer{}
		generator := NewGenerator(compiledCfg.PackageName, "", compiler.NewReplacer(compiledCfg), "")
		generator.builder.writer = outputBuffer

		err = generator.Generate(packageInfos)
		require.NoError(t, err)

		// The legacy helper determines the golden file path from the test name.
		testdataPath := filepath.Join("..", "..", "testdata", "generator")
		testutil.CompareWithGolden(t, testdataPath, *update, outputBuffer.Bytes())
	}

	t.Run("TestPrefix_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "prefixtest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg",
				Alias:  "source",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

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
		runLegacyGoldenTest(t, cfg)
	})

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
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestIgnores", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "ignoretest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Alias:  "source3",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Ignores: []string{"WorkerConfig", "unexportedStruct"}}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestConflict_Constants", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "conflicttest",
			Packages: []*config.Package{
				{Import: "github.com/origadmin/adptool/testdata/sourcepkg", Alias: "source1"},
				{Import: "github.com/origadmin/adptool/testdata/sourcepkg2", Alias: "source2"},
			},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestGenerics_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "generictest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/sourcepkg3",
				Alias:  "source3",
				Types:  []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})
}
