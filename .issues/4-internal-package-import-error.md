---
# Status: Required. Tracks the current progress of the issue.
# Options:
#   - open: The issue is recorded but not yet being worked on.
#   - in-progress: The issue is actively being worked on.
#   - resolved: The issue has been resolved in the codebase.
#   - wont-fix: After evaluation, a decision was made not to fix this issue.
status: resolved

# Assignee: Optional. Your GitHub username.
assignee: 

# Labels: Optional. Used to categorize the issue, separate multiple labels with a comma.
# Example: [bug, enhancement, documentation, performance]
labels: [bug, generator, validation]

# Created At: Optional. Manually record the creation date for reference.
created_at: 

# Remote ID: Do not modify manually. This field is auto-populated by the sync script to link with the GitHub Issue number.
remote_id: 
---

### Description

`adptool` incorrectly handled types from `internal` packages of other Go modules. It would generate code that imported these `internal` packages, which is disallowed by the Go language, leading to compilation errors.

### Solution

The solution was to introduce a stricter type validation mechanism in `internal/generator/collector.go`. This mechanism leverages the `go/types` package to inspect the full import path of every type in a function's signature.

The `containsInvalidTypes` function in `internal/generator/utils.go` was updated to check if a type's import path contains `/internal`. If the `internal` package does not belong to the module currently being processed, `adptool` will skip that function and not generate any code for it, thus preventing the invalid import.

### Verification Steps

A dedicated test case (`TestGenerator_InternalPackageSkip`) was added to validate this specific fix:

1.  **Test Setup**: A special test package was created at `testdata/internaltest/source`. This package contains an `internal/types` subdirectory, mimicking a real-world Go module structure.
2.  **Test Functions**: The source package defines two functions:
    -   `ValidFunc()`: A normal function that should always be processed.
    -   `InvalidFuncWithInternalType()`: A function whose signature uses a type from the forbidden `internal/types` directory.
3.  **Test Execution**: The test runs the generator on this package and captures the output in memory.
4.  **Assertions**: The test then asserts two conditions:
    -   The generated code **must contain** the wrapper for `ValidFunc`.
    -   The generated code **must NOT contain** any reference to `InvalidFuncWithInternalType`.

This test provides definitive proof that the generator correctly identifies and skips functions with invalid `internal` types, ensuring the fix is effective and preventing future regressions.
