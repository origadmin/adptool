package testdata

// --- Type Directives ---

// Test 1: Global rule setting
//go:adapter:type *
//go:adapter:type:struct wrap
//go:adapter:type:disabled false

// Test 2: A specific type that should inherit the global 'wrap' pattern
//go:adapter:type ext1.TypeA
//go:adapter:type:method DoSomethingA
//go:adapter:type:method:rename DoSomethingA_New

// Test 3: A specific type that overrides the global pattern
//go:adapter:type ext1.TypeB
//go:adapter:type:struct copy
//go:adapter:type:field FieldB

// Test 4: A type that explicitly uses the default 'alias' pattern
//go:adapter:type ext1.TypeC
//go:adapter:type:struct alias

// Test 5: A type that uses the 'define' pattern
//go:adapter:type ext1.TypeD
//go:adapter:type:struct define

// Test 6: Context blocks
//go:adapter:context
//go:adapter:package github.com/context/pkg/v3 ctx3
//go:adapter:done

// This type is defined within the context
//go:adapter:type ctx3.ContextType
// It should inherit the global 'wrap' pattern
//go:adapter:type:method DoSomethingCtx

// Nested context
//go:adapter:context
//go:adapter:package github.com/nested/pkg/v4 nested4
//go:adapter:done

// This type is in the nested context
//go:adapter:context
//go:adapter:type nested4.NestedType
//go:adapter:type:struct copy // Override pattern inside nested context
//go:adapter:type:field NestedField
//go:adapter:done // End nested context

// Back in ctx3 context. Test that the pattern reverts.
//go:adapter:type ctx3.AfterNestedType
//go:adapter:type:method DoSomethingAfterNested

// Test 7: A directive with a full import path, should also use global 'wrap' pattern
//go:adapter:type github.com/another/pkg/v2.AnotherExternalType
//go:adapter:type:method DoAnother
