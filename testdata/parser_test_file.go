package main

// This file is for testing the adptool parser's ability to extract directives
// based on the finalized syntax design.

//go:adapter:package github.com/external/pkg/v1 ext1

// Test 1: Global rule setting
//go:adapter:type *
//go:adapter:type:struct wrap

// Test 2: A specific type that should inherit the global 'wrap' pattern
//go:adapter:type ext1.TypeA
//go:adapter:method .DoSomethingA // This should be valid

// Test 3: A specific type that overrides the global pattern
//go:adapter:type ext1.TypeB
//go:adapter:type:struct copy
//go:adapter:field .FieldB // This should be valid
//go:adapter:method .DoSomethingB // This should be invalid as the pattern is 'copy'

// Test 4: A type that explicitly uses the default 'alias' pattern
//go:adapter:type ext1.TypeC
//go:adapter:type:struct alias
//go:adapter:field .FieldC // This should be invalid

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
//go:adapter:method .DoSomethingAfterNested // Should be valid under 'wrap' pattern

//go:adapter:done // End main context

// Test 7: A directive with a full import path, should also use global 'wrap' pattern
//go:adapter:type github.com/another/pkg/v2.AnotherExternalType
//go:adapter:method .DoAnother

// Test 8: Top-level directives for non-struct types
//go:adapter:func ext1.MyExternalFunction
//go:adapter:func:rename MyNewFunction

//go:adapter:var ext1.MyExternalVariable
//go:adapter:var:rename MyNewVariable

//go:adapter:const ext1.MyExternalConstant
//go:adapter:const:ignore true
