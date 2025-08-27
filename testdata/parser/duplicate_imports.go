package parser

import (
	"fmt"
	"io"
	"net/http"
	"os"

	// 问题2: 导入多个包导致的重复.
	// 场景: 导入不同路径但包名相同的包, 如果不使用别名会导致编译错误.
	"github.com/a/b"
	b_alias "github.com/c/b" // 使用别名解决包名冲突

	// 场景: 后缀名一致的重复测试
	_ "github.com/a/pkg"
	_ "github.com/b/pkg"

	// 问题3: 包名和导入的包路径不一致的问题测试
	// 假设 'my_pkg' 目录下的包声明为 'package mypkg'
	_ "github.com/org/my_pkg"

	// 问题4: 导入的包路径名称非标准的测试(包含无法直接使用的字符,但符合命名的,如_.这些).
	_ "github.com/org/with_underscore"
	_ "github.com/org/with.dot"
)

// UseImports 用于演示如何使用导入的包, 避免 "unused import" 错误
func UseImports() {
	// 使用别名导入的包
	b_alias.DoSomething()
	fmt.Println("Using imports")
}

// MyType 模拟一个与导入包中可能存在的类型名冲突的场景
type MyType http.Client

// MyFunc 模拟一个与导入包中可能存在的函数名冲突的场景
func MyFunc(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello")
}

// 使用 os 包, 避免 "unused import" 错误
var _ = os.Getenv("GOPATH")
