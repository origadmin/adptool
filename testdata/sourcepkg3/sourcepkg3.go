package sourcepkg3

import (
	"context"
	"io"
	"time"
)

// --- Complex Interfaces ---

// ComplexGenericInterface defines an interface with generic types and complex method signatures.
type ComplexGenericInterface[T any, K comparable] interface {
	// MethodWithGenericParamsAndReturns takes a context and a generic struct, returns a generic type and an error.
	MethodWithGenericParamsAndReturns(ctx context.Context, data *InputData[T]) (K, error)
	// MethodWithChannel takes a channel and returns another channel.
	MethodWithChannel(input chan T) chan K
	// MethodWithFunction takes a function as parameter.
	MethodWithFunction(func(T) K) error
	// MethodWithNoReturn takes an io.Writer.
	MethodWithNoReturn(writer io.Writer)
	// MethodWithVariadic takes variadic parameters.
	MethodWithVariadic(items ...T) []K
}

// EmbeddedInterface shows interface embedding.
type EmbeddedInterface interface {
	io.Reader
	io.Writer
	ComplexGenericInterface[string, int]
	// AdditionalMethod adds a method to the embedded interfaces.
	AdditionalMethod() bool
}

// --- Complex Structs ---

// InputData is a generic struct used as a parameter.
type InputData[T any] struct {
	ID      int
	Name    string
	Meta    map[string]interface{}
	Payload T
}

// OutputData is a struct used as a return value with embedded struct.
type OutputData struct {
	Success   bool
	Data      []byte
	Timestamp time.Time
	Details   struct {
		Code    int
		Message string
		Nested  map[string]interface{}
	}
}

// Worker is a struct with methods and embedded fields.
type Worker struct {
	Name      string
	ID        int
	CreatedAt time.Time
	Config    *WorkerConfig
}

// WorkerConfig represents worker configuration.
type WorkerConfig struct {
	MaxTasks   int
	Timeout    time.Duration
	RetryCount int
	Features   []string
}

// GenericWorker is a generic version of Worker.
type GenericWorker[T any] struct {
	Name      string
	Data      T
	Processor func(T) error
}

// --- Complex Methods ---

// Process is a method on Worker with complex parameters.
func (w *Worker) Process(ctx context.Context, reader io.Reader, options ...ProcessOption) (*OutputData, error) {
	// implementation detail
	return &OutputData{Success: true, Timestamp: time.Now()}, nil
}

// ProcessWithOptions processes with a configuration struct.
func (w *Worker) ProcessWithOptions(config ProcessConfig) (*OutputData, error) {
	// implementation detail
	return &OutputData{Success: true, Timestamp: time.Now()}, nil
}

// GetConfig returns worker configuration.
func (w *Worker) GetConfig() *WorkerConfig {
	return w.Config
}

// Process is a method on GenericWorker.
func (gw *GenericWorker[T]) Process() error {
	return gw.Processor(gw.Data)
}

// --- Function Types ---

// ProcessFunc is a function type.
type ProcessFunc func(context.Context, *InputData[interface{}]) (*OutputData, error)

// HandlerFunc is a generic function type.
type HandlerFunc[T any] func(T) error

// --- Complex Functions ---

// NewWorker is a constructor function with options pattern.
func NewWorker(name string, options ...WorkerOption) *Worker {
	worker := &Worker{
		Name:      name,
		ID:        int(time.Now().Unix()),
		CreatedAt: time.Now(),
		Config:    &WorkerConfig{},
	}
	
	for _, option := range options {
		option(worker)
	}
	
	return worker
}

// NewGenericWorker creates a new generic worker.
func NewGenericWorker[T any](name string, data T, processor func(T) error) *GenericWorker[T] {
	return &GenericWorker[T]{
		Name:      name,
		Data:      data,
		Processor: processor,
	}
}

// Execute is a standalone function with multiple arguments and returns.
func Execute(
	ctx context.Context,
	api ComplexGenericInterface[string, int],
	input *InputData[string],
	timeout time.Duration,
) (*OutputData, error) {
	// implementation detail
	result, err := api.MethodWithGenericParamsAndReturns(ctx, input)
	if err != nil {
		return nil, err
	}
	return &OutputData{Data: []byte(string(rune(result))), Timestamp: time.Now()}, nil
}

// ExecuteParallel executes multiple operations in parallel.
func ExecuteParallel(
	ctx context.Context,
	apis []ComplexGenericInterface[string, int],
	input *InputData[string],
) ([]*OutputData, error) {
	// implementation detail
	results := make([]*OutputData, len(apis))
	return results, nil
}

// Map applies a function to all elements in a slice.
func Map[T, U any](ts []T, fn func(T) U) []U {
	result := make([]U, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

// Filter filters elements in a slice based on a predicate.
func Filter[T any](ts []T, fn func(T) bool) []T {
	result := make([]T, 0)
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

// --- Complex Types ---

// ProcessOption is a functional option type.
type ProcessOption func(*ProcessConfig)

// ProcessConfig configures the processing behavior.
type ProcessConfig struct {
	Timeout     time.Duration
	MaxRetries  int
	Concurrency int
	FilterFunc  func(interface{}) bool
}

// WorkerOption configures the worker.
type WorkerOption func(*Worker)

// --- Complex Constants and Variables ---

// Status type represents the status of an operation
type Status int

// Status constants using iota
const (
	StatusUnknown Status = iota // 0
	StatusPending               // 1
	StatusRunning               // 2
	StatusSuccess               // 3
	StatusFailed                // 4
)

// Another set of constants using iota with custom start
const (
	PriorityLow Priority = iota + 1 // 1
	PriorityMedium                  // 2
	PriorityHigh                    // 3
)

// Priority type represents priority levels
type Priority int

// Regular constants without iota
const (
	// DefaultTimeout is a typed constant.
	DefaultTimeout time.Duration = 10 * time.Second
	// Version is a string constant.
	Version = "v1.0.0"
	// MaxRetries defines maximum retry attempts.
	MaxRetries = 3
	// unexportedConstant is a non-exported constant.
	unexportedConstant = "internal"
)

var (
	// DefaultWorker is a default instance of a worker.
	DefaultWorker = &Worker{
		Name:      "default",
		ID:        0,
		CreatedAt: time.Now(),
		Config: &WorkerConfig{
			MaxTasks:   10,
			Timeout:    DefaultTimeout,
			RetryCount: MaxRetries,
		},
	}
	// StatsCounter is a public variable.
	StatsCounter int64
	// Processors is a map of named processors.
	Processors map[string]ProcessFunc
	// unexportedVar is a non-exported variable.
	unexportedVar = "internal"
)

// --- Type Aliases ---

// TimeAlias is an alias for time.Time
type TimeAlias = time.Time

// StatusAlias is an alias for Status
type StatusAlias = Status

// IntAlias is an alias for int
type IntAlias = int

// unexportedAlias is a non-exported alias
type unexportedAlias = string

// --- init function ---

func init() {
	Processors = make(map[string]ProcessFunc)
	Processors["default"] = func(ctx context.Context, data *InputData[interface{}]) (*OutputData, error) {
		return &OutputData{Success: true, Timestamp: time.Now()}, nil
	}
}

// --- Unexported types and functions ---

// unexportedStruct is a non-exported struct
type unexportedStruct struct {
	value string
}

// unexportedFunction is a non-exported function
func unexportedFunction() string {
	return "internal function"
}

// unexportedMethod is a method on unexportedStruct
func (u *unexportedStruct) unexportedMethod() string {
	return u.value
}

// ExportedFunctionWithUnexportedParam takes an unexported type as parameter
func ExportedFunctionWithUnexportedParam(u *unexportedStruct) string {
	return u.value
}