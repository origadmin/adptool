package testdata

// go:adapter:ignore pattern1
// go:adapter:ignores pattern2,pattern3
// go:adapter:ignores:json ["pattern4", "pattern5"]

// This file is used to test the parsing of ignore and ignores directives.
// It contains various formats of ignore directives:
// 1. Single pattern using ignore
// 2. Multiple patterns using ignores with comma separation
// 3. Multiple patterns using ignores:json with JSON array format
