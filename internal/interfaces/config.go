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
	IsWildcard  bool
}

// CompiledConfig holds all the compiled information needed for generation.
type CompiledConfig struct {
	PackageName   string // The name of the package to be generated
	Packages      []*CompiledPackage
	Rules         map[string][]RenameRule   // Compiled rules for generator
	PriorityRules map[string][]PriorityRule // Priority rules for generator

	// 按类型分类的全局规则，提高运行时效率
	TypeRules     map[string][]PriorityRule
	FuncRules     map[string][]PriorityRule
	VarRules      map[string][]PriorityRule
	ConstRules    map[string][]PriorityRule

	// 按包和类型分类的规则，进一步提高运行时效率
	PackageTypeRules  map[string]map[string][]PriorityRule
	PackageFuncRules  map[string]map[string][]PriorityRule
	PackageVarRules   map[string]map[string][]PriorityRule
	PackageConstRules map[string]map[string][]PriorityRule
}
