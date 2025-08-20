package interfaces

// RenameRule defines a single renaming rule.
type RenameRule struct {
	Type    string // e.g., "prefix", "suffix", "explicit", "regex"
	Value   string // For prefix/suffix
	From    string // For explicit
	To      string // For explicit
	Pattern string // For regex
	Replace string // For regex
}