package interfaces

// Context defines the interface for passing context across calls.
// It allows for carrying metadata in a key-value manner.
type Context interface {
	WithValue(key, value interface{}) Context
	Value(key interface{}) interface{}
}

// private type to prevent collisions with other packages
type contextKey string

// contextImpl is the concrete implementation of the Context interface.
type contextImpl struct {
	values map[interface{}]interface{}
}

// NewContext creates a new root context.
func NewContext() Context {
	return &contextImpl{values: make(map[interface{}]interface{})}
}

// WithValue returns a new Context that carries a value associated with a key.
func (c *contextImpl) WithValue(key, value interface{}) Context {
	newCtx := &contextImpl{values: make(map[interface{}]interface{})}
	// Copy parent context values
	for k, v := range c.values {
		newCtx.values[k] = v
	}
	// Set the new value
	newCtx.values[key] = value
	return newCtx
}

// Value returns the value associated with a key, or nil if not found.
func (c *contextImpl) Value(key interface{}) interface{} {
	return c.values[key]
}
