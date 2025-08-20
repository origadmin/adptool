package generator

// PackageInfo holds the minimal package information needed by the generator.
type PackageInfo struct {
	ImportPath  string // The import path of the package
	ImportAlias string // The alias for the package import
}
