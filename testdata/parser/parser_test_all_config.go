package parser

//go:adapter package_name my_package
//go:adapter ignores ["file1.go", "dir1/file2.go"]

//go:adapter defaults.mode.strategy replace
//go:adapter defaults.mode.prefix append
//go:adapter defaults.mode.suffix prepend
//go:adapter defaults.mode.explicit merge
//go:adapter defaults.mode.regex merge
//go:adapter defaults.mode.ignores merge

//go:adapter props GlobalVar1 globalValue1
//go:adapter props GlobalVar2 globalValue2

//go:adapter package github.com/my/package/v1
//go:adapter package.alias mypkg
//go:adapter package.path ./vendor/my/package/v1
//go:adapter package.props [ { name: PkgVar, value: PkgValue } ]

//go:adapter type MyStruct
//go:adapter type.kind struct
//go:adapter type.pattern wrap
//go:adapter type.disabled true
//go:adapter type.method DoSomething
//go:adapter type.method.prefix Pre
//go:adapter type.method.suffix Post
//go:adapter type.field MyField
//go:adapter type.field.transforms.before (.*)
//go:adapter type.field.transforms.after New$1

//go:adapter function MyFunc
//go:adapter function.disabled false
//go:adapter function.regex [ { pattern: "Old(.*)", replace: "New$1" } ]

//go:adapter variable MyVar
//go:adapter variable.explicit [ { from: "MyVar", to: "NewVar" } ]

//go:adapter const MyConst
//go:adapter const.ignores ["IgnoredConst"]
