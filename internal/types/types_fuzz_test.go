package types

import (
	"testing"
)

func FuzzMask(f *testing.F) {
	testcases := []string{"", "a", "ab", "abc", "abcd"}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(_ *testing.T, value string) {
		_ = mask(value)
	})
}
