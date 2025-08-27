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
labels: [bug, generator]

# Created At: Optional. Manually record the creation date for reference.
created_at: 

# Remote ID: Do not modify manually. This field is auto-populated by the sync script to link with the GitHub Issue number.
remote_id: 
---

### Description

`adptool` had two related issues when processing function parameters:
1.  When a function parameter had a type but no name (e.g., `func(log.Logger)`), the generated function call would omit this parameter.
2.  When a function parameter used the blank identifier `_` as its name (e.g., `_ context.Context`), the generated function call would also use `_`, which is a syntax error in Go.

### Solution

The root cause for both issues was in the `collectFunctionDeclaration` function in `internal/generator/collector.go`, which failed to handle these special parameter names correctly.

The final solution implements a robust, collision-avoidant parameter name generation strategy:
1.  **Pre-scan**: Before processing parameters, it first iterates through the function signature to record all existing, valid parameter names in a set.
2.  **Safe Generation**: When an unnamed or `_` parameter is encountered, it generates a candidate name (e.g., `p0`).
3.  **Conflict Resolution**: It checks if the generated name already exists in the set. If it does, it increments a counter (`p1`, `p2`...) until a unique name is found.
4.  **Update Set**: The new, unique name is then used in both the function declaration and the function call, and it is added to the set of used names to inform subsequent generations.

This robust approach ensures that all generated function signatures and calls are syntactically correct and free of name collisions.

### Verification Steps

1.  Create a complex function signature as input, including unnamed parameters, `_` parameters, and normally named parameters (e.g., `p0`).
2.  Run `adptool` to generate the adapter code.
3.  Inspect the generated `.adapter.go` file to confirm:
    -   All original unnamed and `_` parameters have been assigned new, non-conflicting, valid names.
    -   The generated function call correctly uses these new names.
    -   The original `p0` parameter name remains unaffected.
4.  Confirm that the generated code compiles successfully with `go build`.
