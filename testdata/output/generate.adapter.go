package aliaspkg

import (
	"github.com/origadmin/adptool/testdata/duplicate/pkg1"
	"github.com/origadmin/adptool/testdata/duplicate/pkg2"
	"github.com/origadmin/adptool/testdata/duplicate/pkg3"
	sourcepkg1 "github.com/origadmin/adptool/testdata/duplicate/sourcepkg"
	sourcepkg21 "github.com/origadmin/adptool/testdata/duplicate/sourcepkg2"
	"github.com/origadmin/adptool/testdata/duplicate/sourcepkg3"
	sourcePkg4 "github.com/origadmin/adptool/testdata/source-pkg4"
	"github.com/origadmin/adptool/testdata/sourcepkg"
	"github.com/origadmin/adptool/testdata/sourcepkg2"
	custompkg "github.com/origadmin/adptool/testdata/sourcepkg3"
)

const (
	MaxRetries2      = custompkg.MaxRetries
	StatusUnknown    = custompkg.StatusUnknown
	StatusPending    = custompkg.StatusPending
	StatusRunning    = custompkg.StatusRunning
	StatusSuccess    = custompkg.StatusSuccess
	StatusFailed     = custompkg.StatusFailed
	PriorityLow      = custompkg.PriorityLow
	PriorityMedium   = custompkg.PriorityMedium
	PriorityHigh     = custompkg.PriorityHigh
	DefaultTimeout1  = custompkg.DefaultTimeout
	Version1         = custompkg.Version
	MaxRetries       = sourcepkg.MaxRetries
	ExportedConstant = sourcepkg.ExportedConstant
	MaxRetries1      = sourcepkg2.MaxRetries
	DefaultTimeout   = sourcepkg2.DefaultTimeout
	Version          = sourcepkg2.Version
)

var (
	ConfigValue2     = custompkg.ConfigValue
	DefaultWorker1   = custompkg.DefaultWorker
	StatsCounter1    = custompkg.StatsCounter
	Processors       = custompkg.Processors
	ConfigValue      = sourcepkg.ConfigValue
	ExportedVariable = sourcepkg.ExportedVariable
	ConfigValue1     = sourcepkg2.ConfigValue
	DefaultWorker    = sourcepkg2.DefaultWorker
	StatsCounter     = sourcepkg2.StatsCounter
)

type (
	CommonStruct2                                = custompkg.CommonStruct
	ComplexGenericInterface[T any, K comparable] = custompkg.ComplexGenericInterface[T, K]
	EmbeddedInterface                            = custompkg.EmbeddedInterface
	InputData1[T any]                            = custompkg.InputData[T]
	OutputData1                                  = custompkg.OutputData
	Worker1                                      = custompkg.Worker
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
	User3                                        = pkg1.User
	UserService2                                 = pkg1.UserService
	Product3                                     = pkg1.Product
	ProductService2                              = pkg1.ProductService
	User                                         = pkg2.User
	UserService                                  = pkg2.UserService
	Product                                      = pkg3.Product
	ProductService                               = pkg3.ProductService
	Handler                                      = sourcePkg4.Handler
	GenericHandler[T any]                        = sourcePkg4.GenericHandler[T]
	Model                                        = sourcePkg4.Model
	Data                                         = sourcePkg4.Data
	Service                                      = sourcePkg4.Service
	CommonStruct                                 = sourcepkg.CommonStruct
	MyStruct                                     = sourcepkg.MyStruct
	ExportedType                                 = sourcepkg.ExportedType
	ExportedInterface                            = sourcepkg.ExportedInterface
	Product1                                     = sourcepkg1.Product
	ProductService1                              = sourcepkg1.ProductService
	User1                                        = sourcepkg1.User
	UserService1                                 = sourcepkg1.UserService
	CommonStruct1                                = sourcepkg2.CommonStruct
	ComplexInterface                             = sourcepkg2.ComplexInterface
	InputData                                    = sourcepkg2.InputData
	OutputData                                   = sourcepkg2.OutputData
	Worker                                       = sourcepkg2.Worker
	User4                                        = sourcepkg21.User
	UserService3                                 = sourcepkg21.UserService
	User2                                        = sourcepkg3.User
	Product2                                     = sourcepkg3.Product
	CommonService                                = sourcepkg3.CommonService
)

func CommonFunction2() string {
	return custompkg.CommonFunction()
}
func NewWorker1(name string, options ...custompkg.WorkerOption) *custompkg.Worker {
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
func CommonFunction() string {
	return sourcepkg.CommonFunction()
}
func ExportedFunction() {
	sourcepkg.ExportedFunction()
}
func CommonFunction1() string {
	return sourcepkg2.CommonFunction()
}
func NewWorker(name string) *sourcepkg2.Worker {
	return sourcepkg2.NewWorker(name)
}
