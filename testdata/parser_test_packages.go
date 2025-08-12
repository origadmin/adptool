package testdata

// Packages configuration
//go:adapter:package github.com/my/package/v1
//go:adapter:package:alias mypkg
//go:adapter:package:path ./vendor/my/package/v1
//go:adapter:package:prop PackageVar1 packageValue1
//go:adapter:package:type MyStructInPackage
//go:adapter:package:type:struct wrap
//go:adapter:package:type:method DoSomethingInPackage
//go:adapter:package:type:method:rename DoSomethingNewInPackage
//go:adapter:package:function MyFuncInPackage
//go:adapter:package:function:rename MyNewFuncInPackage

//go:adapter:done
