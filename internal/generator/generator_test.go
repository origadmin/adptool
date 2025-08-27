package generator

import (
	"bytes"
	"flag"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/origadmin/adptool/internal/compiler"
	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/testutil"
)

var update = flag.Bool("update", false, "update golden files")

// TestMain sets up the test environment, enabling debug logging for slog.
func TestMain(m *testing.M) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))
	os.Exit(m.Run())
}

// TestIssues is a data-driven test that automatically discovers and runs
// regression tests for specific, documented issues.
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
			// Disable the builder's own formatter, as we will format it manually in the test.
			generator := NewGenerator(compiledCfg.PackageName, "", compiler.NewReplacer(compiledCfg), "").WithFormatCode(false)
			generator.builder.writer = outputBuffer

			err = generator.Generate(packageInfos)
			require.NoError(t, err)

			// Manually format the generated code before comparison.
			formatted, err := format.Source(outputBuffer.Bytes())
			require.NoError(t, err, "generated code could not be formatted")

			testutil.CompareWithGoldenFile(t, goldenFilePath, *update, formatted)
		})
	}
}

// TestGenerator_LegacyCases contains the original test suite for basic generator functionality.
// This ensures that our fundamental features remain covered by tests.
func TestGenerator_LegacyCases(t *testing.T) {
	// This helper function is scoped to the legacy tests.
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
		// Disable the builder's own formatter, as we will format it manually in the test.
		generator := NewGenerator(compiledCfg.PackageName, "", compiler.NewReplacer(compiledCfg), "").WithFormatCode(false)
		generator.builder.writer = outputBuffer

		err = generator.Generate(packageInfos)
		require.NoError(t, err)

		// Manually format the generated code before comparison.
		formatted, err := format.Source(outputBuffer.Bytes())
		require.NoError(t, err, "generated code could not be formatted")

		// The legacy tests use the old directory and naming scheme.
		testdataPath := filepath.Join("..", "..", "testdata", "generator")
		testutil.CompareWithGolden(t, testdataPath, *update, formatted)
	}

	t.Run("TestPrefix_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "prefixtest",
			Packages: []*config.Package{{
				Import:    "github.com/origadmin/adptool/testdata/pkgs/source1",
				Alias:     "source",
				Types:     []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
				Variables: []*config.VarRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
				Constants: []*config.ConstRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
				Functions: []*config.FuncRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestConflict_Constants", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "conflicttest",
			Packages: []*config.Package{
				{
					Import: "github.com/origadmin/adptool/testdata/pkgs/source1",
					Alias:  "source1",
				},
				{
					Import: "github.com/origadmin/adptool/testdata/pkgs/source2",
					Alias:  "source2",
				},
			},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestGenerics_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "generictest",
			Packages: []*config.Package{{
				Import:    "github.com/origadmin/adptool/testdata/pkgs/source3",
				Alias:     "source3",
				Types:     []*config.TypeRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "MT"}}},
				Functions: []*config.FuncRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "MF"}}},
				Variables: []*config.VarRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "MV"}}},
				Constants: []*config.ConstRule{{Name: "*", RuleSet: config.RuleSet{Prefix: "MC"}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestIgnores", func(t *testing.T) {
		cfg := &config.Config{
			Ignores:     []string{"ConfigValue", "ExportedVariable"},
			PackageName: "ignoretest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/pkgs/source1",
				Alias:  "source",
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestRegex_Simple", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "regextest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/pkgs/source1",
				Alias:  "source",
				Types: []*config.TypeRule{{Name: "ExportedType", RuleSet: config.RuleSet{Regex: []*config.
				RegexRule{{Pattern: "Exported(.*)", Replace: "My$1"}}}}},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestExplicit_Override", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "overridetest",
			Packages: []*config.Package{{
				Import: "github.com/origadmin/adptool/testdata/pkgs/source1",
				Alias:  "source",
				Types: []*config.TypeRule{
					{Name: "*", RuleSet: config.RuleSet{Prefix: "My"}},
					{Name: "ExportedType", RuleSet: config.RuleSet{Explicit: []*config.ExplicitRule{{From: "ExportedType", To: "CustomType"}}}},
				},
			}},
		}
		runLegacyGoldenTest(t, cfg)
	})

	t.Run("TestNonStandardPackageNames", func(t *testing.T) {
		cfg := &config.Config{
			PackageName: "nonstandardtest",
			Packages: []*config.Package{
				{
					Import: "github.com/origadmin/adptool/testdata/pkgs/source.pkg4",
				},
				{
					Import: "github.com/origadmin/adptool/testdata/pkgs/source-pkg4",
				},
			},
		}
		runLegacyGoldenTest(t, cfg)
	})
}
