package parser

import (
	"fmt"
)

// ContainerFactory defines a function that creates a new instance of a Container.
type ContainerFactory func() Container

// --- Factory ---

type factory struct {
	// The registry is now a slice of factory functions, indexed by RuleType.
	registry []ContainerFactory
}

var defaultFactory = &factory{
	registry: make([]ContainerFactory, 10), // Initial capacity
}

// RegisterContainer registers a factory function for a given RuleType.
// It will resize the registry slice if necessary.
func RegisterContainer(rt RuleType, factoryFunc ContainerFactory) {
	if int(rt) >= len(defaultFactory.registry) {
		// Resize the slice to be large enough.
		newRegistry := make([]ContainerFactory, rt+1)
		copy(newRegistry, defaultFactory.registry)
		defaultFactory.registry = newRegistry
	}
	if defaultFactory.registry[rt] != nil {
		panic(fmt.Sprintf("RegisterContainer: called twice for rule type %d", rt))
	}
	defaultFactory.registry[rt] = factoryFunc
}

// NewContainer creates a new Container instance for a given RuleType.
// It returns nil if the type is not registered or invalid.
func NewContainer(ruleType RuleType) Container {
	if ruleType <= RuleTypeUnknown || int(ruleType) >= len(defaultFactory.registry) || defaultFactory.registry[ruleType] == nil {
		// This should not happen in normal operation as the parser should have
		// already validated the rule type via BuildContainer.
		return invalidRuleInstance
	}
	return defaultFactory.registry[ruleType]()
}

// Container defines the interface for any object that can hold parsed rules
// and participate in the hierarchical configuration structure.
type Container interface {
	// ParseDirective applies a sub-command (e.g., ":rename", ":disabled") to the rule.
	// It takes the builder to interact with the broader parsing state if necessary (e.g., to set an active member).
	ParseDirective(directive *Directive) error

	// AddRule adds a child rule to this container. This is the generic method.
	AddRule(rule any) error

	// AddPackage adds a PackageConfig package configuration to this container.
	AddPackage(pkg *PackageRule) error
	// AddTypeRule adds a TypeRule to this container.
	AddTypeRule(rule *TypeRule) error
	// AddFuncRule adds a FuncRule to this container.
	AddFuncRule(rule *FuncRule) error
	// AddVarRule adds a VarRule to this container.
	AddVarRule(rule *VarRule) error
	// AddConstRule adds a ConstRule to this container.
	AddConstRule(rule *ConstRule) error
	// AddMethodRule adds a MethodRule to this container.
	AddMethodRule(rule *MethodRule) error
	// AddFieldRule adds a FieldRule to this container.
	AddFieldRule(rule *FieldRule) error
	// Finalize performs any post-processing or validation for this container
	// after all its direct rules have been added.
	Finalize(parent Container) error
}

// --- Invalid Rule ---

// InvalidRule is a singleton container returned by the factory when a type is not found.
// Its methods always return an error, allowing for deferred error handling at the call site.
type InvalidRule struct{}

var invalidRuleInstance = &InvalidRule{}

// ParseDirective for an invalid rule always returns an error.
func (i *InvalidRule) ParseDirective(directive *Directive) error {
	return NewParserErrorWithContext(directive, "unrecognized directive command: %s", directive.Command)
}
func (i *InvalidRule) AddRule(rule any) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddPackage(pkg *PackageRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddTypeRule(rule *TypeRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddFuncRule(rule *FuncRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddVarRule(rule *VarRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddConstRule(rule *ConstRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddMethodRule(rule *MethodRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}
func (i *InvalidRule) AddFieldRule(rule *FieldRule) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}

// Finalize for an invalid rule is a no-op.
func (i *InvalidRule) Finalize(parent Container) error {
	return NewParserErrorWithContext(i, "cannot add rule to an invalid container")
}