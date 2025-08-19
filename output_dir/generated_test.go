package aliaspkg

import (
	"context"

	source "github.com/origadmin/adptool/testdata/sourcepkg"
	source2 "github.com/origadmin/adptool/testdata/sourcepkg2"
)

const (
	ExportedConstant = source.ExportedConstant
	DefaultTimeout   = source2.DefaultTimeout
	Version          = source2.Version
)

var (
	ExportedVariable = source.ExportedVariable
	DefaultWorker    = source2.DefaultWorker
	StatsCounter     = source2.StatsCounter
)

type (
	MyStruct          = source.MyStruct
	ExportedType      = source.ExportedType
	ExportedInterface = source.ExportedInterface
	ComplexInterface  = source2.ComplexInterface
	InputData         = source2.InputData
	OutputData        = source2.OutputData
	Worker            = source2.Worker
)

func ExportedFunction() {
	source.ExportedFunction()
}
func NewWorker(name string) *Worker {
	return source2.NewWorker(name)
}
func Execute(ctx context.Context, api ComplexInterface, input *InputData) (*OutputData, error) {
	return source2.Execute(ctx, api, input)
}
