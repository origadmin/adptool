package parser

import "github.com/origadmin/adptool/internal/config"

// init registers all the top-level container rules with the factory.
func init() {
	RegisterContainer("type", func() Container {
		return &TypeRule{TypeRule: &config.TypeRule{}}
	})
	RegisterContainer("func", func() Container {
		return &FuncRule{FuncRule: &config.FuncRule{}}
	})
	RegisterContainer("var", func() Container { return &VarRule{VarRule: &config.VarRule{}} })
	RegisterContainer("const", func() Container { return &ConstRule{ConstRule: &config.ConstRule{}} })
}

// BuildContainer is a high-level constructor that encapsulates the logic for creating
// top-level rule containers. It resolves a command string (including abbreviations)
// to a canonical name and then calls the underlying factory (`NewContainer`).
//
// For commands that do not correspond to a known top-level container type, it returns invalidRuleInstance.
func BuildContainer(cmd string) func() Container {
	// Return a closure that calls the actual factory.
	return func() Container {
		return NewContainer(cmd)
	}
}
