package loader

import (
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"log/slog"

	"github.com/spf13/viper"

	"github.com/origadmin/adptool/internal/config"
	"github.com/origadmin/adptool/internal/parser"
)

var configPaths = []string{
	".", "configs",
}

// LoadConfigFile reads the configuration from a file (or searches for one) and unmarshals it into a Config struct.
func LoadConfigFile(filePath string) (*config.Config, error) {
	v := viper.New()

	if filePath != "" {
		// If a specific file path is provided, use it directly.
		v.SetConfigFile(filePath)
	} else {
		// Otherwise, search for a config file named .adptool in standard paths.
		v.SetConfigName(".adptool")
		v.SetConfigType("yaml") // Explicitly set type for search
		for _, path := range configPaths {
			v.AddConfigPath(path)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		// If the config file is not found, and no specific file was provided, it's not a fatal error.
		// We can proceed with a default/empty config.
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && filePath == "" {
			slog.Debug("No config file found, using default configuration.")
			return config.New(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := config.New() // Initialize with defaults
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	slog.Info("Loaded config from file", "path", v.ConfigFileUsed())
	return cfg, nil
}

// LoadGoFile loads a single Go source file and returns its AST and FileSet.
func LoadGoFile(filePath string) (*goast.File, *gotoken.FileSet, error) {
	fset := gotoken.NewFileSet()
	node, err := goparser.ParseFile(fset, filePath, nil, goparser.ParseComments)
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
	// Create a new config for this file's directives
	cfg := config.New()
	// parser.ParseFileDirectives will update this cfg object
	_, err = parser.ParseFileDirectives(cfg, file, fset)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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