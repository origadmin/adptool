package interfaces

// ContextInfo holds information about the current processing context.
type ContextInfo struct {
	NodeType string // e.g., "const", "var", "type", "func"
}

// RuleType is an enum for different container rule types.
type RuleType int

func (t RuleType) String() string {
	switch t {
	case RuleTypeRoot:
		return "root"
	case RuleTypePackage:
		return "package"
	case RuleTypeType:
		return "type"
	case RuleTypeFunc:
		return "func"
	case RuleTypeVar:
		return "var"
	case RuleTypeConst:
		return "const"
	case RuleTypeMethod:
		return "method"
	case RuleTypeField:
		return "field"
	default:
		return "unknown"
	}
}

const (
	RuleTypeUnknown RuleType = iota
	RuleTypeRoot
	RuleTypePackage
	RuleTypeType
	RuleTypeFunc
	RuleTypeVar
	RuleTypeConst
	RuleTypeMethod
	RuleTypeField
)
