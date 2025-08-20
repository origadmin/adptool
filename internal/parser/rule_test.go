// Package parser implements the functions, types, and interfaces for the module.
package parser

import (
	"strings"

	"github.com/stretchr/testify/mock"

	"github.com/origadmin/adptool/internal/interfaces"
)

// MockContainer is a mock implementation of the Container interface for testing.
// (Copied from rule_const_test.go to avoid circular dependency issues if needed)
type MockContainer struct {
	mock.Mock
}

func (m *MockContainer) Type() interfaces.RuleType {
	args := m.Called()
	return args.Get(0).(interfaces.RuleType)
}

func (m *MockContainer) ParseDirective(directive *Directive) error {
	args := m.Called(directive)
	return args.Error(0)
}

func (m *MockContainer) AddPackage(pkg *PackageRule) error {
	args := m.Called(pkg)
	return args.Error(0)
}

func (m *MockContainer) AddTypeRule(rule *TypeRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddFuncRule(rule *FuncRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddVarRule(rule *VarRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddConstRule(rule *ConstRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddMethodRule(rule *MethodRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddFieldRule(rule *FieldRule) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) AddRule(rule any) error {
	args := m.Called(rule)
	return args.Error(0)
}

func (m *MockContainer) Finalize(parent Container) error {
	args := m.Called(parent)
	return args.Error(0)
}

func decodeTestDirective(directiveString string) Directive {
	if !strings.HasPrefix(directiveString, directivePrefix) {
		return Directive{}
	}

	rawDirective := strings.TrimPrefix(directiveString, directivePrefix)
	commentStart := strings.Index(rawDirective, "//")
	if commentStart != -1 {
		rawDirective = strings.TrimSpace(rawDirective[:commentStart])
	}

	return extractDirective(rawDirective, 0)
}
