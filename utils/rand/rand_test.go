package rand

import (
	"strings"
	"testing"
)

func TestStringWithMode(t *testing.T) {
	// valid := "bcdfghjklmnpqrstvwxz2456789"
	tests := []struct {
		mode  Mode
		valid string
	}{
		{mode: Number, valid: number},
		{mode: AlphaSpecial, valid: alphaSpecial},
		{mode: Number | AlphaUpper, valid: number + alphaUpper},
		{mode: Number | AlphaUpper | AlphaSpecial, valid: number + alphaUpper + alphaSpecial},
		{mode: Number | AlphaLower, valid: number + alphaLower},
		{mode: Number | AlphaLower | AlphaUpper, valid: number + alphaLower + alphaUpper},

		{mode: Number | AlphaLower | AlphaUpper | AlphaSpecial, valid: number + alphaLower + alphaUpper + alphaSpecial},

		{mode: 10, valid: number + alphaLower + alphaUpper + alphaSpecial},
	}
	for _, d := range tests {
		for _, l := range []int{0, 1, 2, 10, 20, 123} {
			s := StringWithMode(l, d.mode)
			if len(s) != l {
				t.Errorf("expected string of size %d, got %q", l, s)
			}
			for _, c := range s {
				if !strings.ContainsRune(d.valid, c) {
					t.Errorf("expected valid characters, got %v", c)
				}
			}
		}
	}
}
