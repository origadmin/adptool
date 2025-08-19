package aliaspkg

import source "github.com/origadmin/adptool/testdata/sourcepkg"

const ExportedConstant = source.ExportedConstant

func ExportedFunction() {
	source.ExportedFunction()
}

type MyStruct = source.MyStruct
type ExportedType = source.ExportedType
type ExportedInterface = source.ExportedInterface

var ExportedVariable = source.ExportedVariable
