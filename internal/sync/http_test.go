package sync

import (
	"testing"
)

func TestPercent(t *testing.T) {
	tests := []struct {
		name string
		a    *int
		b    *int
		want string
	}{
		{name: "both inputs are nil", a: nil, b: nil, want: "0.00"},
		{name: "a is nil, b is non-zero", a: nil, b: ptr(10), want: "0.00"},
		{name: "b is nil, a is non-zero", a: ptr(10), b: nil, want: "0.00"},
		{name: "b is zero", a: ptr(10), b: ptr(0), want: "0.00"},
		{name: "normal case with positive int values", a: ptr(25), b: ptr(100), want: "25.00"},
		{name: "a and b are equal", a: ptr(50), b: ptr(50), want: "100.00"},
		{name: "a is zero, b is positive", a: ptr(0), b: ptr(50), want: "0.00"},
		{name: "large positive values", a: ptr(1000), b: ptr(4000), want: "25.00"},
		{name: "a greater than b", a: ptr(150), b: ptr(100), want: "150.00"},
		{name: "negative values for a and b", a: ptr(-25), b: ptr(-50), want: "50.00"},
		{name: "a is positive, b is negative", a: ptr(25), b: ptr(-50), want: "-50.00"},
		{name: "a is negative, b is positive", a: ptr(-25), b: ptr(50), want: "-50.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.a, tt.b); got != tt.want {
				t.Errorf("percent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
