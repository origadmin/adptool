package generator

import (
	"github.com/origadmin/adptool/internal/config"
)

// Directive represents a parsed //go:adapter directive.
type Directive struct {
	Type  string // e.g., "package", "type", "func", "method", "ignore"
	Value string // The value associated with the directive
	Line  int    // The line number where the directive was found
}

// AdapterSpec holds all the information needed to generate an adapter file.
type AdapterSpec struct {
	PackageName  string            // The package name for the generated file
	Imports      map[string]string // Map of import path to alias
	AdaptedItems []AdaptedItem
}

// AdaptedItem represents a single item (type, func, etc.) to be adapted.
type AdaptedItem struct {
	OriginalName string              // e.g., "Config", "New"
	TargetName   string              // The final generated name after applying rules
	ItemType     string              // "type", "func", "method"
	Rules        []config.RenameRule // List of rules to apply

	// More fields will be needed, e.g., for method's receiver
}
