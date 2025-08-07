package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config represents the overall configuration for adptool.
type Config struct {
	GlobalPrefix string                   `mapstructure:"global_prefix,omitempty"`
	Packages     map[string]PackageConfig `mapstructure:"packages,omitempty"`
}

// PackageConfig represents configuration for a specific Go package.
type PackageConfig struct {
	Alias        string                    `mapstructure:"alias,omitempty"`
	GlobalPrefix string                    `mapstructure:"global_prefix,omitempty"`
	RenameRules  []RenameRule              `mapstructure:"rename_rules,omitempty"`
	Explicit     map[string]string         `mapstructure:"explicit,omitempty"`
	Types        map[string]TypeConfig     `mapstructure:"types,omitempty"`
	Functions    map[string]FunctionConfig `mapstructure:"functions,omitempty"`
	Methods      map[string]MethodConfig   `mapstructure:"methods,omitempty"`
	Ignore       []string                  `mapstructure:"ignore,omitempty"`
}

// TypeConfig represents configuration for a specific type within a package.
type TypeConfig struct {
	Name        string                  `mapstructure:"name,omitempty"`
	RenameRules []RenameRule            `mapstructure:"rename_rules,omitempty"`
	Explicit    map[string]string       `mapstructure:"explicit,omitempty"`
	Methods     map[string]MethodConfig `mapstructure:"methods,omitempty"`
}

// FunctionConfig represents configuration for a specific function within a package.
type FunctionConfig struct {
	Name        string            `mapstructure:"name,omitempty"`
	RenameRules []RenameRule      `mapstructure:"rename_rules,omitempty"`
	Explicit    map[string]string `mapstructure:"explicit,omitempty"`
}

// MethodConfig represents configuration for a specific method within a type.
type MethodConfig struct {
	Name        string            `mapstructure:"name,omitempty"`
	RenameRules []RenameRule      `mapstructure:"rename_rules,omitempty"`
	Explicit    map[string]string `mapstructure:"explicit,omitempty"`
}

// RenameRule defines a rule for renaming.
type RenameRule struct {
	Type    string `mapstructure:"type"` // prefix, suffix, regex, explicit
	Value   string `mapstructure:"value,omitempty"`
	Pattern string `mapstructure:"pattern,omitempty"`
	Replace string `mapstructure:"replace,omitempty"`
}

// LoadConfig loads the adptool configuration from the specified path.
// It handles the priority: fileConfigPath (if provided) > adptool.yaml.
func LoadConfig(fileConfigPath string) (*Config, error) {
	v := viper.New()

	// Set config file name and type
	v.SetConfigName("adptool")   // name of config file (without extension)
	v.AddConfigPath(".")         // path to look for the config file in the current directory
	v.AddConfigPath("./configs") // optionally add a configs directory

	log.Printf("Attempting to load project config from: %s", v.ConfigFileUsed())
	// Read project-level global config (adptool.yaml)
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; ignore error
			log.Println("Project-level adptool.yaml not found, proceeding without it.")
		}
	}

	log.Printf("Project config loaded. Settings: %+v", v.AllSettings())

	// Load file-level config (if provided), which completely replaces project config
	if fileConfigPath != "" {
		log.Printf("Attempting to load file-level config from: %s", fileConfigPath)
		v.SetConfigFile(fileConfigPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read file-level config %s: %w", fileConfigPath, err)
		}
		log.Printf("File-level config loaded. Settings: %+v", v.AllSettings())
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Printf("Final unmarshaled config: %+v", cfg)

	return cfg, nil
}
