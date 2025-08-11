package loader

import (
	"errors"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"

	"github.com/spf13/viper"

	"github.com/origadmin/adptool/internal/config"
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

	// Attempt to read the configuration file.
	if err := v.ReadInConfig(); err != nil {
		// If the error is that the file was not found, it's not a critical error.
		// We can proceed with a default configuration.
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			return config.New(), nil // Return a default config if no file is found.
		}
		// For any other error (e.g., a parsing error), return it.
		return nil, err
	}

	// Create a new default config object to unmarshal into.
	cfg := config.New()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadAdapterFile parses a Go source file and returns its AST.
// It also returns the FileSet used for parsing, which is necessary for position information.
func LoadAdapterFile(filePath string) (*goast.File, *gotoken.FileSet, error) {
	fset := gotoken.NewFileSet()
	node, err := goparser.ParseFile(fset, filePath, nil, goparser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse Go file %s: %w", filePath, err)
	}
	return node, fset, nil
}
