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
		{"both inputs are nil", nil, nil, "0.00"},
		{"a is nil, b is non-zero", nil, new(10), "0.00"},
		{"b is nil, a is non-zero", new(10), nil, "0.00"},
		{"b is zero", new(10), new(0), "0.00"},
		{"normal case with positive int values", new(25), new(100), "25.00"},
		{"a and b are equal", new(50), new(50), "100.00"},
		{"a is zero, b is positive", new(0), new(50), "0.00"},
		{"large positive values", new(1000), new(4000), "25.00"},
		{"a greater than b", new(150), new(100), "150.00"},
		{"negative values for a and b", new(-25), new(-50), "50.00"},
		{"a is positive, b is negative", new(25), new(-50), "-50.00"},
		{"a is negative, b is positive", new(-25), new(50), "-50.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := percent(tt.a, tt.b); got != tt.want {
				t.Errorf("percent() = %v, want %v", got, tt.want)
			}
		})
	}
}
