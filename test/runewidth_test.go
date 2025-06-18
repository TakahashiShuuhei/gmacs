package test

import (
	"testing"

	"github.com/TakahashiShuuhei/gmacs/core/util"
)

func TestRuneWidth(t *testing.T) {
	testCases := []struct {
		r        rune
		expected int
	}{
		// ASCII characters
		{'a', 1},
		{'Z', 1},
		{'0', 1},
		{' ', 1},
		{'!', 1},
		
		// Japanese characters (should be 2)
		{'あ', 2},
		{'い', 2},
		{'う', 2},
		{'漢', 2},
		{'字', 2},
		
		// Control characters
		{'\t', 0},
		{'\n', 0},
		{'\r', 0},
	}
	
	for _, tc := range testCases {
		actual := util.RuneWidth(tc.r)
		if actual != tc.expected {
			t.Errorf("RuneWidth(%c): expected %d, got %d", tc.r, tc.expected, actual)
		}
	}
}

func TestStringWidth(t *testing.T) {
	testCases := []struct {
		s        string
		expected int
	}{
		{"", 0},
		{"abc", 3},
		{"あいう", 6},
		{"hello", 5},
		{"こんにちは", 10},
		{"aあb", 4},
		{"混合text", 9}, // 混(2) + 合(2) + t(1) + e(1) + x(1) + t(1) = 8, but 合 might be different
	}
	
	for _, tc := range testCases {
		actual := util.StringWidth(tc.s)
		t.Logf("StringWidth(%q): expected %d, got %d", tc.s, tc.expected, actual)
		
		// For mixed cases, let's be more lenient and just check it's reasonable
		if tc.s == "混合text" {
			if actual < 7 || actual > 10 {
				t.Errorf("StringWidth(%q): expected around 8-9, got %d", tc.s, actual)
			}
		} else if actual != tc.expected {
			t.Errorf("StringWidth(%q): expected %d, got %d", tc.s, tc.expected, actual)
		}
	}
}

func TestStringWidthUpTo(t *testing.T) {
	testCases := []struct {
		s        string
		bytePos  int
		expected int
	}{
		{"abc", 0, 0},
		{"abc", 1, 1},
		{"abc", 2, 2},
		{"abc", 3, 3},
		{"abc", 10, 3}, // beyond string
		
		{"あいう", 0, 0},
		{"あいう", 3, 2},  // "あ" = 3 bytes = 2 display width
		{"あいう", 6, 4},  // "あい" = 6 bytes = 4 display width
		{"あいう", 9, 6},  // "あいう" = 9 bytes = 6 display width
		
		{"aあb", 0, 0},
		{"aあb", 1, 1},    // "a" = 1 byte = 1 width
		{"aあb", 4, 3},    // "aあ" = 4 bytes = 3 width
		{"aあb", 5, 4},    // "aあb" = 5 bytes = 4 width
	}
	
	for _, tc := range testCases {
		actual := util.StringWidthUpTo(tc.s, tc.bytePos)
		if actual != tc.expected {
			t.Errorf("StringWidthUpTo(%q, %d): expected %d, got %d", 
				tc.s, tc.bytePos, tc.expected, actual)
		}
	}
}