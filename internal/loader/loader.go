package loader

import (
	"errors"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"

	"github.com/spf13/viper"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/parser"
)

// Load reads the configuration from a file (or searches for one) and unmarshals it into a Config struct.
func Load(filePath string) (*config.Config, error) {
	v := viper.New()

	if filePath != "" {
		// If a specific file path is provided, use it directly.
		v.SetConfigFile(filePath)
	} else {
		// Otherwise, search for a config file named .adptool in the current directory.
		v.AddConfigPath(".")
		v.SetConfigName(".adptool")
		v.SetConfigType("yaml") // Default to yaml, but viper will detect others like json, toml.
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	cfg := config.New() // Initialize with defaults
	if err := v.Unmarshal(cfg, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.TextUnmarshallerHookFunc(),
	))); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// LoadGoFile loads a single Go source file and returns its AST and FileSet.
func LoadGoFile(filePath string) (*go_ast.File, *go_token.FileSet, error) {
	fset := go_token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Go file %s: %w", filePath, err)
	}
	return node, fset, nil
}

// LoadGoFileConfig loads a single Go source file, parses its directives, and returns a Config object
// containing only those directives.
func LoadGoFileConfig(filePath string) (*config.Config, error) {
	file, fset, err := LoadGoFile(filePath)
	if err != nil {
		return nil, err
	}
	// parser.ParseFileDirectives returns a *config.Config object containing only directives from this file.
	return parser.ParseFileDirectives(file, fset)
}

// LoadGoFilesConfigs loads multiple Go source files, parses their directives, and returns a map
// of file paths to their respective Config objects.
func LoadGoFilesConfigs(filePaths []string) (map[string]*config.Config, error) {
	configs := make(map[string]*config.Config)
	for _, filePath := range filePaths {
		cfg, err := LoadGoFileConfig(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config for %s: %w", filePath, err)
		}
		configs[filePath] = cfg
	}
	return configs, nil
}
