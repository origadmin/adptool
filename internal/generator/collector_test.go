package generator

import (
	"testing"
)

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid package name",
			input:    "valid",
			expected: "valid",
		},
		{
			name:     "Package name with hyphen",
			input:    "source-pkg4",
			expected: "sourcePkg4",
		},
		{
			name:     "Package name starting with number",
			input:    "123package",
			expected: "p123package",
		},
		{
			name:     "Package name with special characters",
			input:    "my-package_test",
			expected: "myPackage_test",
		},
		{
			name:     "Empty package name",
			input:    "",
			expected: "pkg",
		},
		{
			name:     "Package name with only special characters",
			input:    "-_",
			expected: "pkg",
		},
		{
			name:     "Package name that is a Go keyword",
			input:    "range",
			expected: "rangePkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePackageName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizePackageName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}