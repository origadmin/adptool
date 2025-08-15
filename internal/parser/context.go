// Package parser provides tools for parsing Go source files to extract adapter configurations.
package parser

import (
	"fmt"
	"log/slog"
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
	// activeStacks holds a list of child contexts. This is used to manage scopes
	// where only one child can be active at a time.
	activeStacks []*Context
}

// NewContext creates a new root Context node.
// It takes the data container (Container) for this scope and whether it was explicitly created.
func NewContext(container Container, explicit bool) *Context {
	return &Context{
		explicit:     explicit,
		container:    container,
		active:       true, // A new context is active by default.
		activeStacks: make([]*Context, 0),
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
func (c *Context) StartOrActiveContext(factory ContainerFactory) *Context {
	if active := c.ActiveContext(); active != nil {
		return active
	}
	// Execute the factory function only when a new container is needed.
	container := factory()
	return c.StartContext(container)
}

// ActiveContext finds and returns the currently active child context from the activeStacks.
// It returns nil if no child context is active.
func (c *Context) ActiveContext() *Context {
	// Iterate in reverse to find the most recently added active context first.
	for i := len(c.activeStacks) - 1; i >= 0; i-- {
		stack := c.activeStacks[i]
		if stack.active {
			if deepContext := stack.ActiveContext(); deepContext != nil {
				return deepContext
			}
			return stack
		}
	}
	return nil
}

// StartContext creates a new child context, makes it the sole active context among
// its siblings, and returns it.
func (c *Context) StartContext(container Container) *Context {
	activeContext := &Context{
		explicit:     false,
		active:       true,
		container:    container,
		parent:       c,
		activeStacks: make([]*Context, 0),
	}

	// Deactivate all other sibling contexts to ensure only the new one is active.
	actives := 0
	for _, stack := range c.activeStacks {
		if stack.active {
			stack.active = false
			actives++
		}
	}
	// This is a sanity check. In a correct flow, there should be at most one active context.
	if actives > 1 {
		slog.Warn("more than one active stack was found and deactivated", "count", actives)
	}

	c.activeStacks = append(c.activeStacks, activeContext)
	return activeContext
}

// EndContext deactivates the current context and returns its parent.
// This is used to exit a scope.
// EndContext deactivates the current context, finalizes its container,
// and returns its parent context. This is used to exit a scope.
func (c *Context) EndContext() error {
	c.active = false // Deactivate the current context

	// Finalize the current container and pass its data to the parent
	if c.parent != nil { // Only finalize if there's a parent to pass data to
		currentContainer := c.Container()
		parentContainer := c.parent.Container()
		if err := currentContainer.Finalize(parentContainer); err != nil {
			return fmt.Errorf("failed to finalize container: %w", NewParserError("failed to finalize container"))
		}
		return nil
	}
	return NewParserError("no parent context found for EndContext call")
}
