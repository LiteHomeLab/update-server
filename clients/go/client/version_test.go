package client

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		{"equal", "1.0.0", "1.0.0", 0},
		{"v1 greater", "1.2.0", "1.1.0", 1},
		{"v2 greater", "1.0.0", "2.0.0", -1},
		{"with v prefix", "v1.0.0", "1.0.0", 0},
		{"three parts", "1.2.3", "1.2.2", 1},
		{"four parts", "1.2.3.4", "1.2.3.3", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d",
					tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}
