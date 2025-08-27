package bugfixes

// CustomType is a simple struct for testing.
type CustomType struct {
	Field string
}

// --- Test Cases for Bug Fixes ---

// Case 1: For comparable built-in type.

func GenericFuncWithComparable[T comparable](p1 T) T {
	return p1
}

// Case 2: For purely unnamed parameters.
// All parameters are unnamed, which is valid Go syntax.

func UnnamedParamsTest(int, *CustomType) {}

// Case 3: For blank identifier `_`.
// Parameters can be a mix of named and blank, which is valid Go syntax.

func BlankParamTest(a string, _ bool, _ *CustomType) string {
	return a
}

// Case 4: To test for potential name collisions.
// The generator should not create `p0` or `p1` if they already exist.

func CollisionTest(p0 string, p1 int) string {
	return p0
}
