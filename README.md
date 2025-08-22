# adptool

`adptool` is a highly configurable Go code generation tool that automates the creation of adapter code. It helps you
seamlessly integrate third-party libraries and internal modules by generating type aliases, function wrappers, and
method proxies based on a clear and powerful configuration system.

- **Unify APIs**: Create a consistent, internal-facing API for diverse external libraries.
- **Decouple Dependencies**: Isolate your core logic from specific third-party implementations.
- **Enforce Naming Conventions**: Apply systematic naming rules (prefixes, suffixes, etc.) across your codebase.
- **Eliminate Boilerplate**: Automate the generation of repetitive adapter code.
- **Improve Maintainability**: Adapt to upstream changes by modifying configuration, not manual code.

## Core Concept: Configuration-Driven Generation

`adptool` operates on a simple yet powerful principle: **Your configuration drives the code generation.** You define
*what* to adapt and *how* to adapt it in a single, well-structured YAML file. The tool then parses your Go source code
to apply these rules, generating the necessary adapter code automatically.

## Installation

Ensure you have Go 1.23 or higher installed.

```sh
go install github.com/origadmin/adptool/cmd/adptool@latest
```

## Usage

### 1. Define Directives in Go Source

First, place simple `//go:adapter` directives in your Go files. These directives mark the packages, types, functions,
and methods you intend to adapt. They act as markers for `adptool` to target.

**Example: `directives.go`**

```go
// directives.go
package my_adapters

import (
	// Import the packages you want to adapt. `adptool` uses these imports.
	"github.com/some-org/some-lib"
)

// --- Directives ---
// Mark the entire package for adaptation. Rules are defined in .adptool.yaml
//go:adapter:package github.com/some-org/some-lib

// Mark a specific type for adaptation.
//go:adapter:type some-lib.Client

// Mark a specific function for adaptation.
//go:adapter:func some-lib.NewClient

// Mark a specific method for adaptation.
//go:adapter:method some-lib.Client.Connect

// Explicitly ignore a specific item, overriding all other rules.
//go:adapter:ignore some-lib.DeprecatedFunction
```

### 2. Define Rules in `.adptool.yaml`

This is where the power of `adptool` lies. Create a `.adptool.yaml` file in your project root to define all adaptation
rules. The configuration is hierarchical and maps directly to Go's syntax.

#### Configuration Hierarchy

Rules are applied from the most general to the most specific. A more specific rule set is merged with its parent to
produce a final value.

1. **Top-Level Rules**: These are the global rules that apply to everything.
2. **Package-Level Rules**: Defined inside a `packages` entry, these rules are merged with the top-level rules.

#### Declaration Structure

The configuration structure mirrors Go's syntax. Rules can be defined at the top level (for global rules) or within a
`packages` entry.

- `types`: Rules for `type T ...` declarations.
    - `methods`: Nested rules for the type's methods, `func (t T) M() ...`.
    - `fields`: Nested rules for a `struct`'s fields.
- `functions`: Rules for top-level functions, `func F() ...`.
- `variables`: Rules for `var V ...`.
- `constants`: Rules for `const C ...`.

#### Controlling Merging: The `_mode` Suffix

Every rule (`prefix`, `suffix`, `explicit`, `regex`, `ignore`) can have a `_mode` setting that defines how its value is
calculated based on the parent level.

- The mode for a rule is determined first: a local `_mode` setting is used if present; otherwise, the global default
  from the `defaults.mode` block is used.
- This determined mode is then used to merge the parent's final rule value with the current level's local rule value.

**For `prefix` and `suffix`:**

| Mode      | Behavior                                   | Example (Prefix)                 |
|:----------|:-------------------------------------------|:---------------------------------|
| `append`  | **(Default)** Child is added after parent. | `parent_prefix` + `child_prefix` |
| `prepend` | Child is added before parent.              | `child_prefix` + `parent_prefix` |
| `replace` | Child replaces parent entirely.            | `child_prefix`                   |

**For `explicit`, `regex`, and `ignore` (lists and maps):**

| Mode      | Behavior                                                                                                      |
|:----------|:--------------------------------------------------------------------------------------------------------------|
| `merge`   | **(Default)** Child and parent lists/maps are combined. For `explicit` maps, child keys override parent keys. |
| `replace` | Child list/map replaces the parent's entirely.                                                                |

#### Rule Priority (Within a Single Declaration)

After all rules for a declaration have been merged and calculated, they are applied to a name in this order:

1. **`ignore`**: If the name is in the final `ignore` list, it is skipped.
2. **`explicit`**: If the name is in the final `explicit` map, it is renamed, and **no other rules apply**.
3. **`prefix` & `suffix`**: If not found in `explicit`, the final `prefix` and `suffix` are applied.
4. **`regex`**: The final `regex` rules are applied to the result of the previous step.

### Example: `.adptool.yaml`

```yaml
# .adptool.yaml

# 1. Define global default behaviors
defaults:
  mode:
    prefix: "append"   # Default for all prefixes
    explicit: "merge"  # Default for all explicit maps

# 2. Define top-level (global) rules
functions:
  prefix: "Call_"

types:
  prefix: "T_"
  explicit:
    - from: "Reader"
      to: "SourceReader"
  methods:
    # This inherits its prefix mode ('append') from defaults.mode.prefix
    # It merges with the parent rule from 'functions', resulting in "Call_Method_"
    prefix: "Method_"

# 3. Define package-specific rules that override/merge with global rules
packages:
  - import: "github.com/some-org/some-lib"
    # This package has its own 'types' rules
    types:
      # This prefix merges with the global "T_" prefix, resulting in "T_Lib_"
      prefix: "Lib_"
      methods:
        # This mode overrides the global default. It will not merge.
        prefix_mode: "replace"
        # The final prefix for methods in this package is just "Safe"
        prefix: "Safe"

```

### 3. Run the Tool

Execute `adptool` from your project root. It will find your directives and configuration, and generate the adapter code.

```sh
# Generate code, scanning all .go files in the current directory.
# Output will be in `adapter_generated.go` by default.
adptool generate .

# Specify input files/directories and an output file.
adptool generate -o ./my_adapters.go ./path/to/directives/

# Use a specific config file instead of the default .adptool.yaml
adptool generate -f ./configs/custom.yaml -o ./out.go .
```

## Contributing

Contributions are welcome! Please feel free to submit an Issue or Pull Request.

## Directory Structure

*   `testdata/sourcepkg`: Contains the source package for testing purposes.
*   `output_dir`: The directory where the output of the command is generated.
