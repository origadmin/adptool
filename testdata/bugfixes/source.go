package bugfixes

// GenericFuncWithComparable tests that the comparable built-in is handled correctly.
func GenericFuncWithComparable[T comparable](p1 T) T {
	return p1
}

// ParamsTest tests unnamed, blank, and potentially conflicting parameter names.
func ParamsTest(string, _ int, p0 bool) (string, bool) {
	return "hello", p0
}
