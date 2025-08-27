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

1.  In a Go project, create a function that depends on another module and uses a type from an `internal` package of that dependency in its signature.
2.  Run `adptool` on this project.
3.  Confirm that `adptool` prints a warning message in the logs indicating that the function is being skipped.
4.  Confirm that `adptool` does not generate any adapter code for the function that contains the invalid `internal` type.
