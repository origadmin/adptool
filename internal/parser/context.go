// Package parser provides tools for parsing Go source files to extract adapter configurations.
package parser

import (
	"fmt"
)

// Context represents a node in the parsing state hierarchy.
// It holds the configuration for a specific scope (e.g., a file or a specific struct)
// and maintains links to its parent and children contexts, forming a tree structure.
type Context struct {
	// explicit is true if this context was created by an explicit directive,
	// such as //go:adapter:context.
	explicit bool
	// container holds the actual configuration data for this scope,
	// which can be either a RootConfig or a PackageConfig.
	container Container
	// parent points to the parent context in the hierarchy.
	parent *Context

	// active indicates whether this context is currently active. An inactive context
	// is typically ignored during processing.
	active bool
	// activeContexts holds a list of child contexts. This is used to manage scopes
	// where only one child can be active at a time.
	activeContexts []*Context
}

// NewContext creates a new root Context node.
// It takes the data container (Container) for this scope and whether it was explicitly created.
func NewContext(container Container, explicit bool) *Context {
	return &Context{
		explicit:       explicit,
		container:      container,
		active:         true, // A new context is active by default.
		activeContexts: make([]*Context, 0),
	}
}

// Container returns the data container associated with this context.
func (c *Context) Container() Container {
	return c.container
}

// IsExplicit returns true if this context was created by an explicit directive.
func (c *Context) IsExplicit() bool {
	return c.explicit
}

// SetExplicit sets the explicit flag for the context and returns the context.
func (c *Context) SetExplicit(explicit bool) *Context {
	c.explicit = explicit
	return c
}

// IsActive returns true if the context is currently active.
func (c *Context) IsActive() bool {
	return c.active
}

// SetActivate sets the active status of the context.
func (c *Context) SetActivate(active bool) {
	c.active = active
}

// Parent returns the parent of the current context.
func (c *Context) Parent() *Context {
	return c.parent
}

// StartOrActiveContext gets an active child context or creates a new one.
// It first checks if an active child context already exists and returns it.
// If not, it creates a new one by calling the provided factory function.
func (c *Context) StartOrActiveContext(ruleType RuleType) (*Context, error) {
	if active := c.ActiveContext(); active != nil {
		return active, nil
	}
	// Execute the factory function only when a new containerFactory is needed.
	containerFactory := NewContainerFactory(ruleType)
	container := containerFactory()
	if container.Type() == RuleTypeUnknown {
		return nil, NewParserError("unknown rule type: %s", ruleType.String())
	}
	return c.StartContext(container)
}

// ActiveContext finds and returns the currently active child context from the activeContexts.
// It returns nil if no child context is active.
func (c *Context) ActiveContext() *Context {
	// Iterate in reverse to find the most recently added active context first.
	for i := len(c.activeContexts) - 1; i >= 0; i-- {
		stack := c.activeContexts[i]
		if stack.active {
			return stack
		}
	}
	return nil

}

// StartContext creates a new child context, makes it the sole active context among
// its siblings, and returns it. Before starting the new context, it ensures
// any previously active sibling context is properly ended by calling EndContext.
func (c *Context) StartContext(container Container) (*Context, error) {
	// End any currently active sibling context to ensure its container is finalized.
	if activeChild := c.ActiveContext(); activeChild != nil {
		if err := activeChild.EndContext(); err != nil {
			return nil, fmt.Errorf("failed to end previous context before starting new one: %w", err)
		}
	}

	activeContext := &Context{
		explicit:       false,
		active:         true,
		container:      container,
		parent:         c,
		activeContexts: make([]*Context, 0),
	}

	c.activeContexts = append(c.activeContexts, activeContext)
	return activeContext, nil
}

// EndContext deactivates the current context, finalizes its container, and returns
// its parent context. This process is recursive, ensuring that all active child
// contexts are also ended and finalized from the bottom up.
func (c *Context) EndContext() error {
	// 1. Recursively end all active children first. This ensures their container
	// data is finalized and bubbles up to the current container before it is finalized.
	for _, child := range c.activeContexts {
		if child.IsActive() {
			if err := child.EndContext(); err != nil {
				return fmt.Errorf("failed to recursively end child context: %w", err)
			}
		}
	}

	// 2. Deactivate the current context now that its children are handled.
	c.active = false

	// 3. Finalize the current container's data into its parent's container.
	if c.parent != nil {
		currentContainer := c.Container()
		parentContainer := c.parent.Container()
		if err := currentContainer.Finalize(parentContainer); err != nil {
			return NewParserError("failed to finalize container: %w", err)
		}
	}

	return nil
}
