package sourcepkg

import "fmt"

func ExportedFunction() {
	fmt.Println("This is an exported function.")
}

func unexportedFunction() {
	fmt.Println("This is an unexported function.")
}

type MyStruct struct{}

func (ms *MyStruct) ExportedMethod() {
	fmt.Println("This is an exported method.")
}

func (ms *MyStruct) unexportedMethod() {
	fmt.Println("This is an unexported method.")
}
