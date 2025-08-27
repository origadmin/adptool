package sourcepkg2

import (
	"context"
	"io"
)

// --- Interfaces ---

// ComplexInterface defines an interface with various method signatures.
type ComplexInterface interface {
	// MethodWithParamsAndReturns takes a context and a struct, returns a string and an error.
	MethodWithParamsAndReturns(ctx context.Context, data *InputData) (string, error)
	// MethodWithNoReturn takes an io.Writer.
	MethodWithNoReturn(writer io.Writer)
}

// --- Structs ---

// InputData is a struct used as a parameter.
type InputData struct {
	ID   int
	Name string
	Meta map[string]interface{}
}

// OutputData is a struct used as a return value.
type OutputData struct {
	Success bool
	Data    []byte
}

// Worker is a struct with methods.
type Worker struct {
	Name string
}

// --- Methods ---

// Process is a method on Worker with complex parameters.
func (w *Worker) Process(ctx context.Context, reader io.Reader) (*OutputData, error) {
	// implementation detail
	return &OutputData{Success: true}, nil
}

// --- Functions ---

// NewWorker is a constructor function.
func NewWorker(name string) *Worker {
	return &Worker{Name: name}
}

// Execute is a standalone function with multiple arguments and returns.
func Execute(
	ctx context.Context,
	api ComplexInterface,
	input *InputData,
) (*OutputData, error) {
	// implementation detail
	result, err := api.MethodWithParamsAndReturns(ctx, input)
	if err != nil {
		return nil, err
	}
	return &OutputData{Data: []byte(result)}, nil
}

// --- Constants and Variables ---

const (
	// DefaultTimeout is a typed constant.
	DefaultTimeout = 10
	// Version is a string constant.
	Version = "v1.0.0"
)

var (
	// DefaultWorker is a default instance of a worker.
	DefaultWorker = &Worker{Name: "default"}
	// StatsCounter is a public variable.
	StatsCounter int64
)
