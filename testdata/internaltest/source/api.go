package source

import (
	"github.com/origadmin/adptool/testdata/internaltest/source/internal/types"
)

// ValidFunc is a perfectly valid function that should be generated.
func ValidFunc(s string) string {
	return s
}

// InvalidFuncWithInternalType uses a type from an internal package in its signature.
// The generator should detect this and skip this function entirely.
func InvalidFuncWithInternalType(it types.InternalType) types.InternalType {
	return it
}
