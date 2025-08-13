package parser

// Context is a node in the parsing state hierarchy.
// It points to its scope's actual data container (Container).
type Context struct {
	explicit  bool      // True if this is an explicit //go:adapter:context
	container Container // The actual data container (RootConfig or PackageConfig)
	parent    *Context  // Parent context in the linked list
}

// NewContext creates a new Context node.
// It takes the actual data container (Container) for this scope.
func NewContext(Container Container, explicit bool) *Context {
	return &Context{
		explicit:  explicit,
		container: Container,
	}
}

// Container returns the data container for this context.
func (c *Context) Container() Container {
	return c.container
}

// IsExplicit returns true if this context was explicitly declared.
func (c *Context) IsExplicit() bool {
	return c.explicit
}

func (c *Context) Parent() *Context {
	return c.parent
}

func (c *Context) StartContext() *Context {
	return &Context{
		explicit: true,
		parent:   c,
	}
}

func (c *Context) SetContainer(container Container) {
	c.container = container
}

func (c *Context) StartChildContext(container Container, explicit bool) *Context {
	return &Context{
		explicit:  explicit,
		container: container,
		parent:    c,
	}

}

func (c *Context) EndContext() *Context {
	return c.parent
}
