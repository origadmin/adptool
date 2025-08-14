package parser

import "github.com/origadmin/adptool/internal/config"

// RuleType is an enum for different container rule types.
type RuleType int

// Enum for RuleType
const (
	RuleTypeUnknown RuleType = iota
	RuleTypePackage
	RuleTypeProp
	RuleTypeType // Renamed to avoid conflict with the type keyword
	RuleTypeFunc
	RuleTypeVar
	RuleTypeConst
)

// init registers all the top-level container rules with the factory.
func init() {
	RegisterContainer(RuleTypePackage, func() Container {
		return &PackageRule{Package: &config.Package{}}
	})
	RegisterContainer(RuleTypeType, func() Container {
		return &TypeRule{TypeRule: &config.TypeRule{}}
	})
	RegisterContainer(RuleTypeFunc, func() Container {
		return &FuncRule{FuncRule: &config.FuncRule{}}
	})
	RegisterContainer(RuleTypeVar, func() Container {
		return &VarRule{VarRule: &config.VarRule{}}
	})
	RegisterContainer(RuleTypeConst, func() Container {
		return &ConstRule{ConstRule: &config.ConstRule{}}
	})
}

// NewContainerFactory resolves a command string (including abbreviations) and returns the
// corresponding RuleType constant.
func NewContainerFactory(cmd RuleType) ContainerFactory {
	return func() Container {
		return NewContainer(cmd)
	}
}
