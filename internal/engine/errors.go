package engine

import "fmt"

// LoaderError represents errors that occur during the loading phase.
type LoaderError struct {
	Op  string
	Err error
}

func (e *LoaderError) Error() string {
	return fmt.Sprintf("loader error in %s: %v", e.Op, e.Err)
}

func (e *LoaderError) Unwrap() error {
	return e.Err
}

// PlanError represents errors that occur during the planning phase.
type PlanError struct {
	Op  string
	Err error
}

func (e *PlanError) Error() string {
	return fmt.Sprintf("plan error in %s: %v", e.Op, e.Err)
}

func (e *PlanError) Unwrap() error {
	return e.Err
}

// ExecutionError represents errors that occur during the execution phase.
type ExecutionError struct {
	Op  string
	Err error
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("execution error in %s: %v", e.Op, e.Err)
}

func (e *ExecutionError) Unwrap() error {
	return e.Err
}