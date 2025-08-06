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

The core idea behind `adptool` is **"Comments as Directives"**. You can guide `adptool` on how to generate adapter code
by adding specially formatted comments to your Go source code.

## Installation

Please ensure you have Go 1.23 or higher installed in your development environment.

```sh
go install github.com/origadmin/adptool/cmd/adptool@latest
```

## Usage

`adptool` scans your specified Go files, parses the comment directives within them, and generates adapter code based on
these directives.

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

// --- Global Directives ---
//go:adapter:prefix:global K # Adds "K" prefix to all types and functions not explicitly prefixed

// --- Type Adaptation Directives ---
//go:adapter:type:name OriginalType as MyCustomType # Explicitly specifies the generated type name
type OriginalType struct{} // This is a placeholder; adptool will process the actual OriginalType

//go:adapter:type:prefix My # Adds "My" prefix to the immediately following type
type AnotherType struct{}

// Generates: type MyAnotherType struct{}

//go:adapter:type OriginalStruct # Generates an adapter for OriginalStruct
//go:adapter:method OriginalStruct.MethodA as AdapterMethodA # Generates an adapter method for OriginalStruct.MethodA
//go:adapter:method OriginalStruct.MethodB # Generates an adapter method for OriginalStruct.MethodB, using default naming

// --- Function Adaptation Directives ---
//go:adapter:func:name OriginalFunc as MyCustomFunc # Explicitly specifies the generated function name
func OriginalFunc() {} // This is a placeholder

//go:adapter:func:prefix Adapter # Adds "Adapter" prefix to the immediately following function
func AnotherFunc() {}

// Generates: func AdapterAnotherFunc() {}

// --- Ignore Directives ---
//go:adapter:ignore DeprecatedFunc # Instructs adptool to ignore this function, no adapter will be generated
func DeprecatedFunc() {}

```

### 2. Run `adptool`

From your project root directory or the directory containing your directive files, run the `adptool` command:

```sh
adptool generate -o ./generated/adapters.go ./my_adapters_directives.go
```

- `generate`: The generation subcommand for `adptool`.
- `-o ./generated/adapters.go`: Specifies the output file path for the generated adapter code.
- `./my_adapters_directives.go`: Specifies the Go file containing the directive comments. You can specify multiple files
  or use wildcards (e.g., `./...`).

### 3. Inspect Generated Code

`adptool` will generate Go adapter code in the specified output file based on your directives. Please inspect the
generated file and make any necessary adjustments.

## Directive Comment Syntax

`adptool`'s directive comments follow the `//go:adapter:<category>:<property> <value>` format.

### Global Directives

These directives are typically placed at the top of a directive file and affect the generation behavior for the entire
file.

- `//go:adapter:prefix:global <prefix>`
    * **Description**: Sets a global default prefix to be applied to all types and functions not explicitly prefixed.
    * **Example**: `//go:adapter:prefix:global K`

### Type Adaptation Directives

These directives guide `adptool` on how to generate adapters for Go types (structs, interfaces, etc.). They are usually
placed immediately after `import` statements or before type definitions.

- `//go:adapter:type <OriginalType>`
    * **Description**: Instructs `adptool` to generate a type alias or adapter for `<OriginalType>`.
    * **Example**: `//go:adapter:type Config`
- `//go:adapter:type:name <OriginalType> as <GeneratedName>`
    * **Description**: Explicitly specifies the generated name for `<OriginalType>`, taking the highest precedence.
    * **Example**: `//go:adapter:type:name Decoder as MyCustomDecoder`
- `//go:adapter:type:prefix <prefix>`
    * **Description**: Sets a local prefix for the immediately following type, overriding any global prefix.
    * **Example**: `//go:adapter:type:prefix My`

### Function Adaptation Directives

These directives guide `adptool` on how to generate proxies or adapters for Go functions. They are usually placed
immediately after `import` statements or before function definitions.

- `//go:adapter:func <OriginalFunc>`
    * **Description**: Instructs `adptool` to generate a function proxy for `<OriginalFunc>`.
    * **Example**: `//go:adapter:func New`
- `//go:adapter:func:name <OriginalFunc> as <GeneratedName>`
    * **Description**: Explicitly specifies the generated name for `<OriginalFunc>`, taking the highest precedence.
    * **Example**: `//go:adapter:func:name WithSource as SetSource`
- `//go:adapter:func:prefix <prefix>`
    * **Description**: Sets a local prefix for the immediately following function, overriding any global prefix.
    * **Example**: `//go:adapter:func:prefix Adapter`

### Method Adaptation Directives

These directives guide `adptool` on how to generate adapters for methods of Go types. They are usually placed
immediately after type definitions or before method definitions.

- `//go:adapter:method <OriginalType>.<OriginalMethod>`
    * **Description**: Instructs `adptool` to generate a method adapter for `<OriginalType>`'s `<OriginalMethod>`.
    * **Example**: `//go:adapter:method Config.Load`
- `//go:adapter:method:name <OriginalType>.<OriginalMethod> as <GeneratedName>`
    * **Description**: Explicitly specifies the generated name for `<OriginalMethod>`, taking the highest precedence.
    * **Example**: `//go:adapter:method:name Config.Scan as ConfigScanValues`
- `//go:adapter:method:prefix <prefix>`
    * **Description**: Sets a local prefix for the immediately following method, overriding any global prefix.
    * **Example**: `//go:adapter:method:prefix Adapter`

### Ignore Directives

- `//go:adapter:ignore <OriginalName>`
    * **Description**: Instructs `adptool` to ignore `<OriginalName>` (which can be a type, function, or method), and
      not generate any adapter for it.
    * **Example**: `//go:adapter:ignore DeprecatedFunc`

### External Configuration Anchor

- `//go:adapter:config <path/to/config.yaml>`
    * **Description**: Instructs `adptool` to load more complex generation rules from the specified YAML/JSON file.
    * **Example**: `//go:adapter:config kratos_config_rules.yaml`

## Contributing

Contributions to `adptool` are welcome! If you have any questions, suggestions, or find bugs, feel free to submit an
Issue or Pull Request.

---