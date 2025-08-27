package parser

//go:adapter:ignores:json ["file1.go", "dir1/file2.go"]

//go:adapter:default:mode:strategy replace
//go:adapter:default:mode:prefix append
//go:adapter:default:mode:suffix prepend
//go:adapter:default:mode:explicit merge
//go:adapter:default:mode:regex merge
//go:adapter:default:mode:ignores merge

//go:adapter:property GlobalVar1 globalValue1
//go:adapter:property GlobalVar2 globalValue2

//go:adapter:package github.com/my/package/v1
//go:adapter:package:alias mypkg
//go:adapter:package:path ./vendor/my/package/v1
//go:adapter:package:property PkgVar PkgValue

//go:adapter:type MyStruct
//go:adapter:type:struct wrap
//go:adapter:type:disabled true
//go:adapter:type:method DoSomething
//go:adapter:type:method:prefix Pre
//go:adapter:type:method:suffix Post
//go:adapter:type:field MyField
//go:adapter:type:field:transform:before (.*)
//go:adapter:type:field:transform:after New$1

//go:adapter:function MyFunc
//go:adapter:function:disabled false
//go:adapter:function:regex Old(.*)=New$1

//go:adapter:variable MyVar
//go:adapter:variable:explicit MyVar=NewVar

//go:adapter:const MyConst
//go:adapter:const:ignores IgnoredConst
