# adptool

`adptool` is a Go language code generation tool designed to simplify the generation of adapter code for third-party
libraries or internal modules within Go projects. By parsing specific comment directives in Go source code, `adptool`
can automate the creation of type aliases, function proxies, and method adapters, thereby helping developers to:

- **Unify API Interfaces**: Provide a consistent way to access APIs from different sources.
- **Resolve Naming Conflicts**: Avoid naming conflicts with local code or other libraries through custom naming rules.
- **Encapsulate Third-Party Libraries**: Hide implementation details of third-party libraries, reducing coupling.
- **Reduce Boilerplate Code**: Automate the generation of large amounts of patterned adapter code.
- **Improve Maintainability**: When the adapted library changes, simply update the directives and regenerate the code.

## Core Concepts

The core idea behind `adptool` is **"Comments as Directives"** combined with **external configuration**. You define what
to adapt using simple Go comments, and how to adapt it using flexible YAML/JSON configuration files.

## Installation

Please ensure you have Go 1.23 or higher installed in your development environment.

```sh
go install github.com/origadmin/adptool/cmd/adptool@latest
```

## Usage

`adptool` scans your specified Go files, parses the comment directives within them, and generates adapter code based on
these directives and external configuration.

### 1. Define Directive Files

First, you need to create one or more Go files to serve as directive files for `adptool`. These files typically do not
contain executable code themselves; their purpose is to act as input for `adptool`, guiding code generation through
comments.

For example, create a `my_adapters_directives.go` file:

```go
// my_adapters_directives.go

// This file contains directives for adptool.
// It is not meant to be compiled directly.

package mypackage

import (
	// Import the third-party libraries or modules you wish to adapt
	// adptool will automatically parse these import statements to identify source packages.
	yourlib "github.com/your-org/your-library/your-package"
	mylib "github.com/your-org/your-lib"
)

// --- Package Adaptation Directives ---
// Use this directive to tell adptool which source package to adapt.
// The detailed adaptation rules (e.g., prefixes, explicit names) for this package
// are defined in external configuration files (adptool.yaml or file-level config).
//go:adapter:package github.com/your-org/your-library/your-package

//go:adapter:package github.com/your-org/your-lib

// --- Type Adaptation Directives ---
// Use this directive to mark a type from the source package for adaptation.
// The actual type definition is in the source package (e.g., yourlib).
// The generated name and other rules are defined in external configuration.
//go:adapter:type Config # Refers to yourlib.Config
//go:adapter:type Decoder # Refers to yourlib.Decoder

// --- Function Adaptation Directives ---
// Use this directive to mark a function from the source package for adaptation.
// The generated name and other rules are defined in external configuration.
//go:adapter:func New # Refers to yourlib.New
//go:adapter:func WithSource # Refers to yourlib.WithSource

// --- Method Adaptation Directives ---
// Use this directive to mark a method of a type from the source package for adaptation.
// The type must also be marked with //go:adapter:type.
// The generated name and other rules are defined in external configuration.
//go:adapter:method Config.Load # Refers to yourlib.Config.Load
//go:adapter:method Config.Scan # Refers to yourlib.Config.Scan

// --- Ignore Directives ---
// Use this directive to explicitly ignore a type, function, or method from adaptation.
// This overrides any rules defined in external configuration.
//go:adapter:ignore DeprecatedFunc # Instructs adptool to ignore this function
```

### 2. Define Configuration Files

`adptool` uses a simple, layered configuration system to define how adapters are generated. The configuration is defined
in a file named `adptool.yaml`.

#### Configuration Structure

The configuration is now structured to allow specific rules for `types`, `functions`, and `methods`. Rules at the top
level apply globally, while rules inside a `packages` entry apply to a specific package.

Each category (`types`, `functions`, `methods`) can have the following rules:

- **`prefix`**: A string added to the beginning of all generated names within this category.
- **`suffix`**: A string added to the end of all generated names within this category.
- **`explicit`**: A map to explicitly rename a specific item. This has the highest priority.
- **`regex`**: A list of regular expression rules for more complex renaming.
- **`ignore`**: A list of names to exclude from adaptation within this category.
- **`inherit_explicit`**: (Boolean, default `false`) If `true`, package-specific `explicit` rules will be merged with
  global `explicit` rules for this category. Package rules override global rules on conflict. If `false`,
  package-specific `explicit` rules completely replace global ones (unless empty, then global is used as fallback).
- **`inherit_regex`**: (Boolean, default `false`) If `true`, package-specific `regex` rules will be appended to global
  `regex` rules for this category.
- **`inherit_ignore`**: (Boolean, default `false`) If `true`, package-specific `ignore` rules will be appended to global
  `ignore` rules for this category.

The top-level `adptool.yaml` can define global rules for `types`, `functions`, and `methods`. The `packages` entry
allows overriding these global rules for specific imported packages.

#### Rule Priority

For each name, `adptool` applies the renaming rules in a strict order:

1. **`//go:adapter:ignore` directives**: These always take precedence and prevent any adaptation.
2. **`explicit`** rules are checked first. If a name matches, it is renamed and no other rules are applied.
3. **`prefix`** is added.
4. **`suffix`** is added.
5. **`regex`** rules are applied in the order they are defined.

#### Example (`adptool.yaml`)

```yaml
# adptool.yaml (Project Root)

# Global rules that apply to all packages unless overridden.

types:
  # Explicitly rename `OldType` to `NewType` everywhere.
  explicit:
    OldType: NewType
  # Add "K" to the beginning of every generated type name.
  prefix: "K"
  # Add "Adapter" to the end of every generated type name.
  suffix: "Adapter"
  # Apply regex rules after all other rules for types.
  regex:
    - pattern: "Type$"
      replace: "TypeV2"
  # Ignore specific types from adaptation.
  ignore:
    - "DeprecatedType"

functions:
  # Explicitly rename `OldFunc` to `NewFunc` everywhere.
  explicit:
    OldFunc: NewFunc
  # Add "Func" to the beginning of every generated function name.
  prefix: "Func"
  # Add "Wrapper" to the end of every generated function name.
  suffix: "Wrapper"
  # Apply regex rules after all other rules for functions.
  regex:
    - pattern: "Service$"
      replace: "ServiceV2"
  # Ignore specific functions from adaptation.
  ignore:
    - "InternalFunc"

methods:
  # Explicitly rename `Config.Load` to `ConfigLoad` everywhere.
  explicit:
    Config.Load: ConfigLoad
  # Add "Method" to the beginning of every generated method name.
  prefix: "Method"
  # Add "Impl" to the end of every generated method name.
  suffix: "Impl"
  # Apply regex rules after all other rules for methods.
  regex:
    - pattern: "Scan$"
      replace: "ScanResult"
  # Ignore specific methods from adaptation.
  ignore:
    - "Config.DeprecatedMethod"

# --- Package-Specific Overrides ---
packages:
  # Rules for the "github.com/your-org/your-lib" package
  - import: "github.com/your-org/your-lib"
    # path is optional. If provided, adptool will load the source code
    # from this local directory, which is useful for vendored modules
    # or projects in a monorepo.
    path: "./vendor/github.com/your-org/your-lib"
    alias: "yourlib"

    # Override global rules for types in this package only.
    types:
      prefix: "YourLibType"
      # Example: Merge package-specific explicit rules with global ones.
      # If not set or false, package explicit rules replace global ones.
      inherit_explicit: true
      explicit:
        AnotherType: MyAnotherType
      ignore:
        - "PackageSpecificTypeToIgnore"

    # Override global rules for functions in this package only.
    functions:
      prefix: "YourLibFunc"
      explicit:
        YourLib.OldFunc: YourLib.NewFunc

    # Override global rules for methods in this package only.
    methods:
      prefix: "YourLibMethod"
      explicit:
        YourLib.Config.Load: YourLib.ConfigLoad
```

#### Configuration Loading

- **Project-level (`adptool.yaml`)**: `adptool` automatically searches for and loads `adptool.yaml` from the current
  directory or a `./configs` subdirectory.
- **File-level (`-f` flag)**: You can provide a specific configuration file using the `-f` flag (e.g.,
  `adptool generate -f my_config.yaml ...`). If specified, this file\'s configuration **completely replaces** any
  project-level `adptool.yaml`.

### 3. Run `adptool`

From your project root directory, run the `adptool` command:

```bash
# Basic usage: scans directive files and uses adptool.yaml (if present)
adptool generate ./my_adapters_directives.go
# Output will be generated to ./my_adapters_directives.adapter.go by default.

# Specify output file explicitly:
adptool generate -o ./generated/adapters.go ./my_adapters_directives.go

# With a specific file-level configuration:
adptool generate -o ./generated/adapters.go -f ./my_file_config.yaml ./my_adapters_directives.go

# Scan multiple directive files or directories:
adptool generate -o ./generated/adapters.go ./my_adapters_directives.go ./another_directives.go
adptool generate -o ./generated/adapters.go ./my_directives_dir/...
```

- `generate`: The generation subcommand for `adptool`.
- `-o <output_file_path>`: Optional. Specifies the output file path for the generated adapter code. If not provided, the
  output file name defaults to `<input_file_name>.adapter.go` for single input files. For multiple inputs, `-o` is
  required.
- `-f <file_level_config.yaml>`: Optional. Specifies a file-level configuration that completely replaces the
  project-level `adptool.yaml`.
- `<directive_files_or_dirs>`: Specifies the Go files or directories containing the directive comments.

### 4. Inspect Generated Code

`adptool` will generate Go adapter code in the specified output file based on your directives and configuration. Please
inspect the generated file and make any necessary adjustments.

## Directive Comment Syntax

`adptool`\'s directive comments follow the `//go:adapter:<category> <value>` format.

- **`//go:adapter:package <import_path>`**
    * **Description**: Specifies a source Go package whose types, functions, and methods are to be adapted. `adptool`
      will automatically parse the `import` statements in the directive file to identify the source package\'s alias (if
      any).
    * **Example**: `//go:adapter:package github.com/your-org/your-library/your-package`

- **`//go:adapter:type <OriginalType>`**
    * **Description**: Marks a specific type from the source package for adaptation. Adaptation rules (e.g., naming,
      methods to include) are defined in external configuration.
    * **Example**: `//go:adapter:type Config` (refers to `yourlib.Config` if `yourlib` is imported)

- **`//go:adapter:func <OriginalFunc>`**
    * **Description**: Marks a specific function from the source package for adaptation. Adaptation rules are defined in
      external configuration.
    * **Example**: `//go:adapter:func New` (refers to `yourlib.New`)

- **`//go:adapter:method <OriginalType>.<OriginalMethod>`**
    * **Description**: Marks a specific method of a type from the source package for adaptation. The type must also be
      marked with `//go:adapter:type`. Adaptation rules are defined in external configuration.
    * **Example**: `//go:adapter:method Config.Load` (refers to `yourlib.Config.Load`)

- **`//go:adapter:ignore <OriginalName>`**
    * **Description**: Explicitly instructs `adptool` to ignore `<OriginalName>` (which can be a type, function, or
      method), overriding any rules defined in external configuration.
    * **Example**: `//go:adapter:ignore DeprecatedFunc`

## Configuration Priority

`adptool` applies configuration rules with the following priority (highest to lowest):

1. **`//go:adapter:ignore` directives**: These always take precedence and prevent any adaptation.
2. **Rules from File-level Configuration (`-f` flag)**: If a file-level config is provided, its rules completely replace
   the project-level `adptool.yaml`.
3. **Rules from Project-level Global Configuration (`adptool.yaml`)**: Default rules for the entire project.
4. **`adptool`'s Built-in Defaults**: If no configuration is provided, `adptool` uses its internal default behavior (
   e.g., no prefixes, direct naming).

## Contributing

Contributions to `adptool` are welcome! If you have any questions, suggestions, or find bugs, feel free to submit an
Issue or Pull Request.
