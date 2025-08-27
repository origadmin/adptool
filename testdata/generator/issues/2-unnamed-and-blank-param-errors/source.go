package source

// This file contains test cases for issue #2.

// CustomType is a simple struct for testing.
type CustomType struct {
	Field string
}

// Case 2a: For purely unnamed parameters.
func UnnamedParamsTest(int, *CustomType) {}

// Case 2b: For blank identifier `_`.
func BlankParamTest(a string, _ bool, _ *CustomType) string {
	return a
}

// Case 2c: To test for potential name collisions.
func CollisionTest(p0 string, p1 int) string {
	return p0
}
