package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/origadmin/adptool/internal/config"
)

// Engine is the main engine for adptool.
type Engine struct {
	logger *slog.Logger
}

// Config holds the engine configuration.
type Config struct {
	// Add configuration options here
}

// Result holds the result of the engine execution.
type Result struct {
	// Add result fields here
}

// Option is a function that configures the Engine.
type Option func(*Engine)

// New creates a new Engine.
func New(opts ...Option) *Engine {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	engine := &Engine{
		logger: logger,
	}
	
	// Apply options
	for _, opt := range opts {
		opt(engine)
	}
	
	return engine
}

// WithLogger sets the logger for the engine.
func WithLogger(logger *slog.Logger) Option {
	return func(e *Engine) {
		e.logger = logger
	}
}

// Execute processes the input and generates output.
func (e *Engine) Execute(ctx context.Context, cfg *Config) (*Result, error) {
	e.logger.Info("Starting execution")

	// Create components
	loader := NewLoader(
		os.DirFS("."),
		NewFileSystemParser(),
		&config.Config{},
		e.logger,
	)

	compiler := NewRealCompiler()
	generator := NewRealGenerator(e.logger)
	
	planner := NewPlanner(
		&config.Config{},
		&loggerAdapter{logger: e.logger},
		compiler,
		generator,
	)

	executor := NewExecutor(
		generator,
		compiler,
		&loggerAdapter{logger: e.logger},
	)

	// 1. Load phase
	loadCtx, err := loader.Load(ctx, []string{"."})
	if err != nil {
		return nil, fmt.Errorf("failed to load files: %w", err)
	}

	// 2. Plan phase
	plan, err := planner.Plan(loadCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to create plan: %w", err)
	}

	// 3. Execute phase
	if err := executor.Execute(ctx, plan); err != nil {
		return nil, fmt.Errorf("failed to execute plan: %w", err)
	}

	e.logger.Info("Execution completed successfully")
	return &Result{}, nil
}

// ExecuteFile processes a single Go file and generates its adapter.
func (e *Engine) ExecuteFile(filePath string, cfg *config.Config) error {
	e.logger.Info("Processing file", "file", filePath)

	// Create a context
	ctx := context.Background()

	// 1. Load phase
	loader := NewLoader(
		os.DirFS("."),
		NewFileSystemParser(),
		cfg,
		e.logger,
	)

	loadCtx, err := loader.Load(ctx, []string{filePath})
	if err != nil {
		return fmt.Errorf("failed to load files: %w", err)
	}

	// 2. Plan phase
	// For now, we'll use a simplified planner
	compiler := NewRealCompiler()
	generator := NewRealGenerator(e.logger)
	
	planner := NewPlanner(cfg, &loggerAdapter{logger: e.logger}, compiler, generator)
	plan, err := planner.Plan(loadCtx)
	if err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	// 3. Execute phase
	// For now, we'll use a simplified executor
	executor := NewExecutor(generator, compiler, &loggerAdapter{logger: e.logger})
	if err := executor.Execute(ctx, plan); err != nil {
		return fmt.Errorf("failed to execute plan: %w", err)
	}

	e.logger.Info("Successfully processed file", "file", filePath)
	return nil
}

// loggerAdapter adapts slog.Logger to the Logger interface
type loggerAdapter struct {
	logger *slog.Logger
}

func (l *loggerAdapter) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

func (l *loggerAdapter) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

func (l *loggerAdapter) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}