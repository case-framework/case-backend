package utils

import (
	"testing"
	"time"
)

func TestParseDurationString(t *testing.T) {
	tests := []struct {
		input      string
		expected   time.Duration
		shouldFail bool
	}{
		{"", 0, true},
		{"1", 0, true},
		{"1s", time.Second, false},
		{"1m", time.Minute, false},
		{"1h", time.Hour, false},
		{"1d", 0, true}, // not supported
		{"1w", 0, true}, // not supported
		{"1y", 0, true}, // not supported
		{"1ms", time.Millisecond, false},
		{"1us", time.Microsecond, false},
		{"1ns", time.Nanosecond, false},
	}

	for _, test := range tests {
		result, err := ParseDurationString(test.input)
		if test.shouldFail {
			if err == nil {
				t.Errorf("expected error for input %s, but got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("expected no error for input %s, but got %s", test.input, err)
			}
			if result != test.expected {
				t.Errorf("expected %s for input %s, but got %s", test.expected, test.input, result)
			}
		}
	}
}
