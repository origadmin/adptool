package aliaspkg

import (
	pkg1 "github.com/origadmin/adptool/testdata/duplicate/pkg1"
	pkg2 "github.com/origadmin/adptool/testdata/duplicate/pkg2"
	pkg3 "github.com/origadmin/adptool/testdata/duplicate/pkg3"
	sourcepkg1 "github.com/origadmin/adptool/testdata/duplicate/sourcepkg"
	sourcepkg21 "github.com/origadmin/adptool/testdata/duplicate/sourcepkg2"
	sourcepkg3 "github.com/origadmin/adptool/testdata/duplicate/sourcepkg3"
	sourcePkg4 "github.com/origadmin/adptool/testdata/source-pkg4"
	sourcepkg "github.com/origadmin/adptool/testdata/sourcepkg"
	sourcepkg2 "github.com/origadmin/adptool/testdata/sourcepkg2"
	custompkg "github.com/origadmin/adptool/testdata/sourcepkg3"
)

const (
	MaxRetries       = custompkg.MaxRetries
	StatusUnknown    = custompkg.StatusUnknown
	StatusPending    = custompkg.StatusPending
	StatusRunning    = custompkg.StatusRunning
	StatusSuccess    = custompkg.StatusSuccess
	StatusFailed     = custompkg.StatusFailed
	PriorityLow      = custompkg.PriorityLow
	PriorityMedium   = custompkg.PriorityMedium
	PriorityHigh     = custompkg.PriorityHigh
	DefaultTimeout   = custompkg.DefaultTimeout
	Version          = custompkg.Version
	MaxRetries1      = sourcepkg.MaxRetries
	ExportedConstant = sourcepkg.ExportedConstant
	MaxRetries2      = sourcepkg2.MaxRetries
	DefaultTimeout1  = sourcepkg2.DefaultTimeout
	Version1         = sourcepkg2.Version
)

var (
	ConfigValue      = custompkg.ConfigValue
	DefaultWorker    = custompkg.DefaultWorker
	StatsCounter     = custompkg.StatsCounter
	Processors       = custompkg.Processors
	ConfigValue1     = sourcepkg.ConfigValue
	ExportedVariable = sourcepkg.ExportedVariable
	ConfigValue2     = sourcepkg2.ConfigValue
	DefaultWorker1   = sourcepkg2.DefaultWorker
	StatsCounter1    = sourcepkg2.StatsCounter
)

type (
	CommonStruct1                                 = custompkg.CommonStruct
	ComplexGenericInterface1[T any, K comparable] = custompkg.ComplexGenericInterface[T, K]
	EmbeddedInterface1                            = custompkg.EmbeddedInterface
	InputData1[T any]                             = custompkg.InputData[T]
	OutputData1                                   = custompkg.OutputData
	Worker1                                       = custompkg.Worker
	WorkerConfig1                                 = custompkg.WorkerConfig
	GenericWorker1[T any]                         = custompkg.GenericWorker[T]
	ProcessFunc1                                  = custompkg.ProcessFunc
	HandlerFunc1[T any]                           = custompkg.HandlerFunc[T]
	ProcessOption1                                = custompkg.ProcessOption
	ProcessConfig1                                = custompkg.ProcessConfig
	WorkerOption1                                 = custompkg.WorkerOption
	Status1                                       = custompkg.Status
	Priority1                                     = custompkg.Priority
	TimeAlias1                                    = custompkg.TimeAlias
	StatusAlias1                                  = custompkg.StatusAlias
	IntAlias1                                     = custompkg.IntAlias
	User1                                         = pkg1.User
	UserService1                                  = pkg1.UserService
	Product1                                      = pkg1.Product
	ProductService1                               = pkg1.ProductService
	User3                                         = pkg2.User
	UserService2                                  = pkg2.UserService
	Product2                                      = pkg3.Product
	ProductService11                              = pkg3.ProductService
	Handler1                                      = sourcePkg4.Handler
	GenericHandler1[T any]                        = sourcePkg4.GenericHandler[T]
	Model1                                        = sourcePkg4.Model
	Data1                                         = sourcePkg4.Data
	Service1                                      = sourcePkg4.Service
	CommonStruct11                                = sourcepkg.CommonStruct
	MyStruct1                                     = sourcepkg.MyStruct
	ExportedType1                                 = sourcepkg.ExportedType
	ExportedInterface1                            = sourcepkg.ExportedInterface
	Product3                                      = sourcepkg1.Product
	ProductService2                               = sourcepkg1.ProductService
	User4                                         = sourcepkg1.User
	UserService3                                  = sourcepkg1.UserService
	CommonStruct2                                 = sourcepkg2.CommonStruct
	ComplexInterface1                             = sourcepkg2.ComplexInterface
	InputData11                                   = sourcepkg2.InputData
	OutputData11                                  = sourcepkg2.OutputData
	Worker11                                      = sourcepkg2.Worker
	User11                                        = sourcepkg21.User
	UserService11                                 = sourcepkg21.UserService
	User2                                         = sourcepkg3.User
	Product11                                     = sourcepkg3.Product
	CommonService1                                = sourcepkg3.CommonService
)

func CommonFunction() string {
	return custompkg.CommonFunction()
}
func NewWorker(name string, options ...custompkg.WorkerOption) *custompkg.Worker {
	return custompkg.NewWorker(name, options...)
}
func NewGenericWorker[T any](name string, data T, processor func(T) error) *custompkg.GenericWorker[T] {
	return custompkg.NewGenericWorker[T](name, data, processor)
}
func Map[T, U any](ts []T, fn func(T) U) []U {
	return custompkg.Map[T, U](ts, fn)
}
func Filter[T any](ts []T, fn func(T) bool) []T {
	return custompkg.Filter[T](ts, fn)
}
func CommonFunction1() string {
	return sourcepkg.CommonFunction()
}
func ExportedFunction() {
	sourcepkg.ExportedFunction()
}
func CommonFunction2() string {
	return sourcepkg2.CommonFunction()
}
func NewWorker1(name string) *sourcepkg2.Worker {
	return sourcepkg2.NewWorker(name)
}
