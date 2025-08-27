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
1.  When a function parameter had a type but no name (e.g., `func(int)`), the generated function call would omit this parameter.
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

To ensure robust and accurate testing, the original, invalid test case was discarded. It was replaced with a series of separate, valid Go functions in `testdata/bugfixes/source.go`, with each function targeting a specific scenario:

1.  **`UnnamedParamsTest(int, *CustomType)`**: This function is valid Go syntax and tests the generator's ability to handle purely unnamed parameters.
2.  **`BlankParamTest(a string, _ bool, _ *CustomType)`**: This function tests the correct handling of the blank identifier `_` mixed with named parameters.
3.  **`CollisionTest(p0 string, p1 int)`**: This function explicitly includes `p0` and `p1` in its signature to verify that the generator's name-generation logic correctly avoids collisions.

The `TestBugFixes` test case runs the generator on these functions and compares the output to a golden file containing the correct, expected code for each isolated case. This guarantees the fixes are both correct and independently verified.
