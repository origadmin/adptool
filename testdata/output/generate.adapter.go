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
	CommonStruct                                 = custompkg.CommonStruct
	ComplexGenericInterface[T any, K comparable] = custompkg.ComplexGenericInterface[T, K]
	EmbeddedInterface                            = custompkg.EmbeddedInterface
	InputData[T any]                             = custompkg.InputData[T]
	OutputData                                   = custompkg.OutputData
	Worker                                       = custompkg.Worker
	WorkerConfig                                 = custompkg.WorkerConfig
	GenericWorker[T any]                         = custompkg.GenericWorker[T]
	ProcessFunc                                  = custompkg.ProcessFunc
	HandlerFunc[T any]                           = custompkg.HandlerFunc[T]
	ProcessOption                                = custompkg.ProcessOption
	ProcessConfig                                = custompkg.ProcessConfig
	WorkerOption                                 = custompkg.WorkerOption
	Status                                       = custompkg.Status
	Priority                                     = custompkg.Priority
	TimeAlias                                    = custompkg.TimeAlias
	StatusAlias                                  = custompkg.StatusAlias
	IntAlias                                     = custompkg.IntAlias
	User1                                        = pkg1.User
	UserService                                  = pkg1.UserService
	Product1                                     = pkg1.Product
	ProductService                               = pkg1.ProductService
	User2                                        = pkg2.User
	UserService1                                 = pkg2.UserService
	Product2                                     = pkg3.Product
	ProductService1                              = pkg3.ProductService
	Handler                                      = sourcePkg4.Handler
	GenericHandler[T any]                        = sourcePkg4.GenericHandler[T]
	Model                                        = sourcePkg4.Model
	Data                                         = sourcePkg4.Data
	Service                                      = sourcePkg4.Service
	CommonStruct1                                = sourcepkg.CommonStruct
	MyStruct                                     = sourcepkg.MyStruct
	ExportedType                                 = sourcepkg.ExportedType
	ExportedInterface                            = sourcepkg.ExportedInterface
	Product3                                     = sourcepkg1.Product
	ProductService2                              = sourcepkg1.ProductService
	User3                                        = sourcepkg1.User
	UserService2                                 = sourcepkg1.UserService
	CommonStruct2                                = sourcepkg2.CommonStruct
	ComplexInterface                             = sourcepkg2.ComplexInterface
	InputData1                                   = sourcepkg2.InputData
	OutputData1                                  = sourcepkg2.OutputData
	Worker1                                      = sourcepkg2.Worker
	User4                                        = sourcepkg21.User
	UserService3                                 = sourcepkg21.UserService
	User                                         = sourcepkg3.User
	Product                                      = sourcepkg3.Product
	CommonService                                = sourcepkg3.CommonService
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
