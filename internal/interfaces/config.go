package interfaces

// CompiledPackage holds the compiled information for a single source package.
type CompiledPackage struct {
	ImportPath  string
	ImportAlias string
	Types       interface{} // Types defined in this package
	Functions   interface{} // Functions defined in this package
	Variables   interface{} // Variables defined in this package
	Constants   interface{} // Constants defined in this package
}

type PriorityRule struct {
	Rule        RenameRule
	PackageName string
	Priority    int
}

// CompiledConfig holds all the compiled information needed for generation.
type CompiledConfig struct {
	PackageName   string // The name of the package to be generated
	Packages      []*CompiledPackage
	Rules         map[string][]RenameRule   // Compiled rules for generator
	PriorityRules map[string][]PriorityRule // Priority rules for generator
}
