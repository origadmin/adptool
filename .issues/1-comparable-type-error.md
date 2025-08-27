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

When processing the built-in Go type `comparable`, `adptool` incorrectly treats it as `[package_name].comparable`, leading to generated code that fails to compile.

### Solution

The root cause was that the `isBuiltinType` function in `internal/generator/utils.go` was missing `comparable` in its list of built-in types.

By adding `"comparable": true` to the `builtinTypes` map within the `isBuiltinType` function, the generator now correctly identifies `comparable` as a built-in type that should not be qualified with a package name, thus resolving the issue.

### Verification Steps

1.  Create a Go source file with a `comparable` type parameter as input for `adptool`.
2.  Run `adptool` to generate the adapter code.
3.  Inspect the generated `.adapter.go` file to confirm that the `comparable` type is not prefixed with any package name.
4.  Confirm that the generated code compiles successfully with `go build`.
