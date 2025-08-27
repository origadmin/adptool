package source

// This test case is for issue #1.
// It ensures that the `comparable` built-in type is handled correctly
// and not qualified with a package name.

func GenericFuncWithComparable[T comparable](p1 T) T {
	return p1
}
