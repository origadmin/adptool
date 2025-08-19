package aliaspkg

import (
	"context"

	source "github.com/origadmin/adptool/testdata/sourcepkg"
	source2 "github.com/origadmin/adptool/testdata/sourcepkg2"
)

const ExportedConstant = source.ExportedConstant

var ExportedVariable = source.ExportedVariable

type (
	MyStruct          source.MyStruct
	ExportedType      source.ExportedType
	ExportedInterface source.ExportedInterface
)

func ExportedFunction() {
	source.ExportedFunction()
}

const (
	DefaultTimeout = source2.DefaultTimeout
	Version        = source2.Version
)

var (
	DefaultWorker = source2.DefaultWorker
	StatsCounter  = source2.StatsCounter
)

type (
	ComplexInterface source2.ComplexInterface
	InputData        source2.InputData
	OutputData       source2.OutputData
	Worker           source2.Worker
)

func NewWorker(name string) *source2.Worker {
	return source2.NewWorker(name)
}
func Execute(ctx context.Context, api source2.ComplexInterface, input *source2.InputData) (*source2.OutputData, error) {
	return source2.Execute(ctx, api, input)
}
