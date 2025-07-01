package json2csv

import (
	"testing"
)

func TestStringArray_Set(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "single value",
			input:    "test",
			expected: []string{"test"},
		},
		{
			name:     "multiple values",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "values with spaces",
			input:    "a, b , c",
			expected: []string{"a", " b ", " c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sa StringArray
			err := sa.Set(tt.input)
			if err != nil {
				t.Errorf("Set() error = %v", err)
				return
			}
			if len(sa) != len(tt.expected) {
				t.Errorf("Set() got len = %v, want %v", len(sa), len(tt.expected))
				return
			}
			for i := range sa {
				if sa[i] != tt.expected[i] {
					t.Errorf("Set() got[%d] = %v, want %v", i, sa[i], tt.expected[i])
				}
			}
		})
	}
}

func TestStringArray_String(t *testing.T) {
	tests := []struct {
		name     string
		array    StringArray
		expected string
	}{
		{
			name:     "empty array",
			array:    StringArray{},
			expected: "[]",
		},
		{
			name:     "single element",
			array:    StringArray{"test"},
			expected: "[test]",
		},
		{
			name:     "multiple elements",
			array:    StringArray{"a", "b", "c"},
			expected: "[a b c]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.array.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
