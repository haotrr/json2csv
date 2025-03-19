package json2csv

import (
	"testing"
)

func TestStringArray_Set(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			want:    []string{""},
			wantErr: false,
		},
		{
			name:    "single value",
			input:   "test",
			want:    []string{"test"},
			wantErr: false,
		},
		{
			name:    "multiple values",
			input:   "a,b,c",
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "values with spaces",
			input:   "a, b, c",
			want:    []string{"a", " b", " c"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var arr StringArray
			err := arr.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArray.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(arr) != len(tt.want) {
				t.Errorf("StringArray.Set() got length = %v, want length %v", len(arr), len(tt.want))
				return
			}

			for i := range tt.want {
				if arr[i] != tt.want[i] {
					t.Errorf("StringArray.Set() got[%d] = %v, want[%d] = %v", i, arr[i], i, tt.want[i])
				}
			}
		})
	}
}

func TestStringArray_String(t *testing.T) {
	tests := []struct {
		name string
		arr  StringArray
		want string
	}{
		{
			name: "empty array",
			arr:  StringArray{},
			want: "[]",
		},
		{
			name: "single element",
			arr:  StringArray{"test"},
			want: "[test]",
		},
		{
			name: "multiple elements",
			arr:  StringArray{"a", "b", "c"},
			want: "[a b c]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.arr.String(); got != tt.want {
				t.Errorf("StringArray.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
