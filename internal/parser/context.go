package parser

import (
	"github.com/origadmin/adptool/internal/config"
)

// Context is the single stateful object for the parsing process.
// It holds the config being built, the active rule, and the current directive for error reporting.
// It also manages the explicit vs. implicit context state.
type Context struct {
	Config         *config.Config
	Directive      *Directive
	activeRule     interface{}
	activeMember   *config.MemberRule
	currentPackage *config.Package

	// inExplicitContext is true if the parser is within a 'context'... 'done' block.
	inExplicitContext bool
}

// NewContext creates a new, initialized context.
func NewContext() *Context {
	return &Context{
		Config: config.New(),
	}
}

// --- Scope and State Management ---

// StartExplicitContext puts the parser into an explicit context mode.
func (c *Context) StartExplicitContext() {
	c.inExplicitContext = true
	c.activeRule = nil
	c.activeMember = nil
}

// EndScope clears the current scope (package or explicit context).
func (c *Context) EndScope() {
	c.inExplicitContext = false
	c.currentPackage = nil
	c.activeRule = nil
	c.activeMember = nil
}

// SetPackageScope sets the current package for subsequent rule additions.
func (c *Context) SetPackageScope(pkg *config.Package) {
	c.currentPackage = pkg
	c.Config.Packages = append(c.Config.Packages, pkg)
	c.activeRule = nil
	c.activeMember = nil
}

// SetActiveRule sets the main rule being processed.
// In an implicit context, this replaces the previous active rule.
// In an explicit context, it only sets the rule if one is not already active.
func (c *Context) SetActiveRule(rule interface{}) {
	if c.inExplicitContext && c.activeRule != nil {
		return // In an explicit context, the first rule is locked in.
	}
	c.activeRule = rule
	c.activeMember = nil
}

// --- Rule Addition ---

// AddTypeRule adds a type rule to the correct scope(s).
// Per the existing test logic, package-scoped rules are also added to the global list.
func (c *Context) AddTypeRule(rule *config.TypeRule) {
	c.Config.Types = append(c.Config.Types, rule)
	if c.currentPackage != nil {
		c.currentPackage.Types = append(c.currentPackage.Types, rule)
	}
}

// AddFuncRule adds a function rule to the correct scope(s).
func (c *Context) AddFuncRule(rule *config.FuncRule) {
	c.Config.Functions = append(c.Config.Functions, rule)
	if c.currentPackage != nil {
		c.currentPackage.Functions = append(c.currentPackage.Functions, rule)
	}
}

// AddVarRule adds a variable rule to the correct scope(s).
func (c *Context) AddVarRule(rule *config.VarRule) {
	c.Config.Variables = append(c.Config.Variables, rule)
	if c.currentPackage != nil {
		c.currentPackage.Variables = append(c.currentPackage.Variables, rule)
	}
}

// AddConstRule adds a constant rule to the correct scope(s).
func (c *Context) AddConstRule(rule *config.ConstRule) {
	c.Config.Constants = append(c.Config.Constants, rule)
	if c.currentPackage != nil {
		c.currentPackage.Constants = append(c.currentPackage.Constants, rule)
	}
}

// --- Active Rule Accessors ---

func (c *Context) ActiveRule() interface{} {
	return c.activeRule
}

func (c *Context) SetActiveMember(member *config.MemberRule) {
	c.activeMember = member
}

func (c *Context) ActiveMember() *config.MemberRule {
	return c.activeMember
}
