package source

import (
	"github.com/origadmin/adptool/testdata/generator/issues/4-internal-package-import-error/source/internal/types"
)

// This is a valid function that should be generated.
func ValidFunc(s string) string {
	return s
}

// This function uses a type from an internal package in its signature.
// The generator should detect this and skip this function entirely.
func InvalidFuncWithInternalType(it types.InternalType) types.InternalType {
	return it
}
