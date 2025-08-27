---
title: 'Refactor testdata directory for clarity and consistency'
status: 'open'
---

## Summary

The `testdata` directory has become disorganized over time, containing a mix of top-level Go files, inconsistently named source packages, and various test case directories. This makes it difficult to navigate and maintain our test suite.

This issue proposes a comprehensive cleanup and reorganization of the `testdata` directory to establish a clear and consistent structure.

## Current Problems

1.  **Top-level Go files:** A large number of `parser_test_*.go` files are located at the root of `testdata`, creating clutter.
2.  **Inconsistent source packages:** Source packages used for testing (e.g., `sourcepkg`, `sourcepkg2`, `source-pkg4`) are scattered at the top level with inconsistent naming.
3.  **Unclear structure:** The purpose of directories like `duplicate` and `output` is not immediately obvious without deeper inspection.
4.  **Legacy files:** Some files and directories may no longer be relevant to the current test suite.

## Proposed Reorganization

To address these issues, I propose the following new structure for the `testdata` directory:

```
testdata/
├── e2e/                  # End-to-end tests
│   └── ...
├── generator/            # Golden files for generator tests
│   ├── issues/
│   └── ...
├── parser/               # Test source files specifically for the parser
│   ├── constants.go
│   ├── defaults.go
│   └── ...
└── sources/              # All source packages used as input for tests
    ├── source1/
    ├── source2/
    └── ...
```

### Key Changes

1.  **Create `parser/` directory:** All `parser_test_*.go` files will be moved into this directory and renamed to reflect their content (e.g., `parser_test_constants.go` -> `constants.go`).
2.  **Create `sources/` directory:** All source packages (`sourcepkg`, `sourcepkg2`, etc.) will be consolidated under this directory with consistent naming (`source1`, `source2`, etc.).
3.  **Review and Relocate:** The contents of `config`, `output`, and `duplicate` will be reviewed. Relevant files will be moved to the appropriate new directory, and obsolete files will be removed.
4.  **Update tests:** All tests that rely on the `testdata` directory will be updated to reflect the new file paths.

This refactoring will make our test data easier to manage, understand, and extend in the future.
