package utils

import "testing"

func TestIsURLSafe(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"test123", true},
		{"h/m", false},
		{"?test", false},
		{"t est", false},
		{"z.z", false},
		{"\t ", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := IsURLSafe(tt.value); got != tt.expected {
				t.Errorf("IsURLSafe() = %v, want %v", got, tt.expected)
			}
		})
	}
}
