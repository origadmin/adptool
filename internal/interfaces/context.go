package interfaces

// ContextKey is the type for context keys.
type ContextKey string

// PackagePathContextKey is the context key for the package path.
const PackagePathContextKey = ContextKey("packagePath")

// Context defines the interface for passing context across calls.
// It allows for carrying metadata in a key-value manner and managing a stack of node types.
type Context interface {
	WithValue(key, value interface{}) Context
	Value(key interface{}) interface{}
	Push(nodeType RuleType) Context
	CurrentNodeType() RuleType
}

// contextImpl is the concrete implementation of the Context interface.
type contextImpl struct {
	values        map[interface{}]interface{}
	nodeTypeStack []RuleType
}

// NewContext creates a new root context.
func NewContext() Context {
	return &contextImpl{
		values:        make(map[interface{}]interface{}),
		nodeTypeStack: make([]RuleType, 0),
	}
}

// WithValue returns a new Context that carries a value associated with a key.
func (c *contextImpl) WithValue(key, value interface{}) Context {
	newCtx := &contextImpl{
		values:        make(map[interface{}]interface{}),
		nodeTypeStack: make([]RuleType, len(c.nodeTypeStack)),
	}
	// Copy parent context values
	for k, v := range c.values {
		newCtx.values[k] = v
	}
	// Copy parent node type stack
	copy(newCtx.nodeTypeStack, c.nodeTypeStack)
	// Set the new value
	newCtx.values[key] = value
	return newCtx
}

// Value returns the value associated with a key, or nil if not found.
func (c *contextImpl) Value(key interface{}) interface{} {
	return c.values[key]
}

// Push adds a node type to the context stack.
func (c *contextImpl) Push(nodeType RuleType) Context {
	newStack := make([]RuleType, len(c.nodeTypeStack)+1)
	copy(newStack, c.nodeTypeStack)
	newStack[len(c.nodeTypeStack)] = nodeType

	newCtx := &contextImpl{
		values:        make(map[interface{}]interface{}),
		nodeTypeStack: newStack,
	}
	// Copy parent context values
	for k, v := range c.values {
		newCtx.values[k] = v
	}
	return newCtx
}

// CurrentNodeType returns the current node type from the context stack.
func (c *contextImpl) CurrentNodeType() RuleType {
	if len(c.nodeTypeStack) > 0 {
		return c.nodeTypeStack[len(c.nodeTypeStack)-1]
	}
	return RuleTypeUnknown
}
