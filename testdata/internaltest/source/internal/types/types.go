package types

// InternalType is a type defined within an internal package.
// Go's tooling should prevent this type from being used outside of the
// parent package ('source').
type InternalType struct {
	ID string
}
