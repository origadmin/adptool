package parser

import (
	"fmt"
)

// ContainerFactory defines a function that creates a new instance of a Container.
type ContainerFactory func() Container

// --- Factory ---

type factory struct {
	registry map[string]ContainerFactory
}

var defaultFactory = &factory{
	registry: make(map[string]ContainerFactory),
}

// RegisterContainer makes a container type available by name.
// This is the public registration function that delegates to the internal factory.
func RegisterContainer(name string, factoryFunc ContainerFactory) {
	if _, dup := defaultFactory.registry[name]; dup {
		panic("RegisterContainer: called twice for container " + name)
	}
	defaultFactory.registry[name] = factoryFunc
}

// NewContainer creates a new Container instance from its registered name.
// If the name is not found, it returns a special InvalidRule singleton.
func NewContainer(name string) Container {
	factoryFunc, ok := defaultFactory.registry[name]
	if !ok {
		return invalidRuleInstance
	}
	return factoryFunc()
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
	Finalize() error
}

// --- Invalid Rule ---

// InvalidRule is a singleton container returned by the factory when a type is not found.
// Its methods always return an error, allowing for deferred error handling at the call site.
type InvalidRule struct{}

var invalidRuleInstance = &InvalidRule{}

// ParseDirective for an invalid rule always returns an error.
func (i *InvalidRule) ParseDirective(directive *Directive) error {
	return fmt.Errorf("unrecognized directive command: %s", directive.Command)
}
func (i *InvalidRule) AddRule(rule any) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddPackage(pkg *PackageRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddTypeRule(rule *TypeRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddFuncRule(rule *FuncRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddVarRule(rule *VarRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddConstRule(rule *ConstRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddMethodRule(rule *MethodRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}
func (i *InvalidRule) AddFieldRule(rule *FieldRule) error {
	return fmt.Errorf("cannot add rule to an invalid container")
}

// Finalize for an invalid rule is a no-op.
func (i *InvalidRule) Finalize() error {
	return nil
}
