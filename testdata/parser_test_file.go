package main

// This file is for testing the adptool parser's ability to extract directives
// based on the finalized syntax design, covering all Config parameters.

// --- Top-level Config Directives ---

// Defaults configuration
//go:adapter:defaults:mode:strategy replace
//go:adapter:defaults:mode:prefix append
//go:adapter:defaults:mode:suffix append
//go:adapter:defaults:mode:explicit merge
//go:adapter:defaults:mode:regex merge
//go:adapter:defaults:mode:ignores merge

// Vars configuration
//go:adapter:vars GlobalVar1 globalValue1
//go:adapter:vars GlobalVar2 globalValue2

// Packages configuration
//go:adapter:package github.com/my/package/v1
//go:adapter:package:alias mypkg
//go:adapter:package:path ./vendor/my/package/v1
//go:adapter:package:vars PackageVar1 packageValue1
//go:adapter:package:types MyStructInPackage
//go:adapter:package:types:struct wrap
//go:adapter:package:types:methods DoSomethingInPackage
//go:adapter:package:types:methods:rename DoSomethingNewInPackage
//go:adapter:package:functions MyFuncInPackage
//go:adapter:package:functions:rename MyNewFuncInPackage

// --- Type Directives ---

// Test 1: Global rule setting
//go:adapter:type *
//go:adapter:type:struct wrap
//go:adapter:type:disabled false

// Test 2: A specific type that should inherit the global 'wrap' pattern
//go:adapter:type ext1.TypeA
//go:adapter:method .DoSomethingA
//go:adapter:method:rename DoSomethingA_New

// Test 3: A specific type that overrides the global pattern
//go:adapter:type ext1.TypeB
//go:adapter:type:struct copy
//go:adapter:field .FieldB

// Test 4: A type that explicitly uses the default 'alias' pattern
//go:adapter:type ext1.TypeC
//go:adapter:type:struct alias

// Test 5: A type that uses the 'define' pattern
//go:adapter:type ext1.TypeD
//go:adapter:type:struct define

// Test 6: Context blocks
//go:adapter:context
//go:adapter:package github.com/context/pkg/v3 ctx3

// This type is defined within the context
//go:adapter:type ctx3.ContextType
// It should inherit the global 'wrap' pattern
//go:adapter:method .DoSomethingCtx

// Nested context
//go:adapter:context
//go:adapter:package github.com/nested/pkg/v4 nested4

// This type is in the nested context
//go:adapter:type nested4.NestedType
//go:adapter:type:struct copy // Override pattern inside nested context
//go:adapter:field .NestedField

//go:adapter:done // End nested context

// Back in ctx3 context. Test that the pattern reverts.
//go:adapter:type ctx3.AfterNestedType
//go:adapter:method .DoSomethingAfterNested

//go:adapter:done // End main context

// Test 7: A directive with a full import path, should also use global 'wrap' pattern
//go:adapter:type github.com/another/pkg/v2.AnotherExternalType
//go:adapter:method .DoAnother

// --- Other Top-level Directives ---

// Test 8: Top-level directives for non-struct types
//go:adapter:func ext1.MyExternalFunction
//go:adapter:func:rename MyNewFunction

//go:adapter:var ext1.MyExternalVariable
//go:adapter:var:rename MyNewVariable

//go:adapter:ignores ext1.MyExternalConstant
//go:adapter:const ext1.MyExternalConstant
