package parser

import "github.com/origadmin/adptool/internal/config"

// Context manages the current parsing context.
// It holds the configuration being built and the state of the parser.
// It acts as a state machine, tracking the current scope (global, package, type, etc.)
// and the active rule being configured.
type Context struct {
	Directive      *Directive
	Config         *config.Config
	CurrentPackage *config.Package // If not nil, new top-level rules are added to this package.

	// activeRule holds the last created top-level rule that implements the Rule interface.
	activeRule Rule

	// activeMemberRule holds the last created member rule (Method, Field)
	// so that sub-directives can be applied to it.
	activeMemberRule *config.MemberRule
}

// NewContext creates a new, initialized parser context.
func NewContext() *Context {
	return &Context{
		Config: config.New(),
	}
}

// Reset clears all scope-specific context, returning to the global scope.
func (c *Context) Reset() {
	c.CurrentPackage = nil
	c.activeRule = nil
	c.activeMemberRule = nil
}

// EndPackageScope clears the package and active rule context.
func (c *Context) EndPackageScope() {
	c.CurrentPackage = nil
	c.ResetActiveRule()
}

// ResetActiveRule clears the context of the currently active rule.
func (c *Context) ResetActiveRule() {
	c.activeRule = nil
	c.activeMemberRule = nil
}

// SetActiveRule sets the main rule being processed and resets any sub-rule context.
func (c *Context) SetActiveRule(rule Rule) {
	c.ResetActiveRule()
	c.activeRule = rule
}

// SetActiveMemberRule sets the member rule being processed.
func (c *Context) SetActiveMemberRule(rule *config.MemberRule) {
	c.activeMemberRule = rule
}

// --- Scope-aware Rule Addition ---

func (c *Context) AddTypeRule(rule *config.TypeRule) {
	if c.CurrentPackage != nil {
		if c.CurrentPackage.Types == nil {
			c.CurrentPackage.Types = make([]*config.TypeRule, 0)
		}
		c.CurrentPackage.Types = append(c.CurrentPackage.Types, rule)
	} else {
		if c.Config.Types == nil {
			c.Config.Types = make([]*config.TypeRule, 0)
		}
		c.Config.Types = append(c.Config.Types, rule)
	}
}

func (c *Context) AddFuncRule(rule *config.FuncRule) {
	if c.CurrentPackage != nil {
		if c.CurrentPackage.Functions == nil {
			c.CurrentPackage.Functions = make([]*config.FuncRule, 0)
		}
		c.CurrentPackage.Functions = append(c.CurrentPackage.Functions, rule)
	} else {
		if c.Config.Functions == nil {
			c.Config.Functions = make([]*config.FuncRule, 0)
		}
		c.Config.Functions = append(c.Config.Functions, rule)
	}
}

func (c *Context) AddVarRule(rule *config.VarRule) {
	if c.CurrentPackage != nil {
		if c.CurrentPackage.Variables == nil {
			c.CurrentPackage.Variables = make([]*config.VarRule, 0)
		}
		c.CurrentPackage.Variables = append(c.CurrentPackage.Variables, rule)
	} else {
		if c.Config.Variables == nil {
			c.Config.Variables = make([]*config.VarRule, 0)
		}
		c.Config.Variables = append(c.Config.Variables, rule)
	}
}

func (c *Context) AddConstRule(rule *config.ConstRule) {
	if c.CurrentPackage != nil {
		if c.CurrentPackage.Constants == nil {
			c.CurrentPackage.Constants = make([]*config.ConstRule, 0)
		}
		c.CurrentPackage.Constants = append(c.CurrentPackage.Constants, rule)
	} else {
		if c.Config.Constants == nil {
			c.Config.Constants = make([]*config.ConstRule, 0)
		}
		c.Config.Constants = append(c.Config.Constants, rule)
	}
}

// --- Active Rule Accessors ---

func (c *Context) ActiveRule() Rule {
	return c.activeRule
}

func (c *Context) ActiveMemberRule() *config.MemberRule {
	return c.activeMemberRule
}
