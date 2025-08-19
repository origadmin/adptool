package sourcepkg

type ExportedType struct {
	ExportedField string
	unexportedField string
}

type unexportedType struct {
	Field string
}

type ExportedInterface interface {
	ExportedInterfaceMethod() string
	unexportedInterfaceMethod() int
}

type unexportedInterface interface {
	Method()
}
