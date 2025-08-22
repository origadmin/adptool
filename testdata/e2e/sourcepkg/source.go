package sourcepkg

import "fmt"

const ExportedConstant = "Hello, World!"

var ExportedVariable = 123

type ExportedInterface interface {
	DoSomething(val int) error
}

type ExportedStruct struct {
	Name string
	Value int
}

func ExportedFunction(s string) {
	fmt.Println(s)
}
