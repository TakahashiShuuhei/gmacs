package util

import "unicode"

// RuneWidth returns the display width of a rune in terminal
// Based on East Asian Width properties:
// - ASCII: 1 column
// - East Asian Full Width: 2 columns
// - East Asian Half Width: 1 column
// - Other characters: 1 column (default)
func RuneWidth(r rune) int {
	if r < 0x20 {
		return 0 // Control characters
	}
	
	if r < 0x7F {
		return 1 // ASCII
	}
	
	// East Asian Full Width characters
	if isEastAsianFullWidth(r) {
		return 2
	}
	
	// Default to 1 for other characters
	return 1
}

// StringWidth returns the total display width of a string
func StringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}

// StringWidthUpTo returns the display width up to a specific byte position
func StringWidthUpTo(s string, bytePos int) int {
	if bytePos <= 0 {
		return 0
	}
	if bytePos >= len(s) {
		return StringWidth(s)
	}
	
	substr := s[:bytePos]
	return StringWidth(substr)
}

// isEastAsianFullWidth checks if a rune is East Asian Full Width
func isEastAsianFullWidth(r rune) bool {
	// Common CJK ranges that are typically full-width
	return (r >= 0x1100 && r <= 0x115F) || // Hangul Jamo
		(r >= 0x2E80 && r <= 0x2EFF) ||    // CJK Radicals Supplement
		(r >= 0x2F00 && r <= 0x2FDF) ||    // Kangxi Radicals
		(r >= 0x2FF0 && r <= 0x2FFF) ||    // Ideographic Description Characters
		(r >= 0x3000 && r <= 0x303F) ||    // CJK Symbols and Punctuation
		(r >= 0x3040 && r <= 0x309F) ||    // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) ||    // Katakana
		(r >= 0x3100 && r <= 0x312F) ||    // Bopomofo
		(r >= 0x3130 && r <= 0x318F) ||    // Hangul Compatibility Jamo
		(r >= 0x3190 && r <= 0x319F) ||    // Kanbun
		(r >= 0x31A0 && r <= 0x31BF) ||    // Bopomofo Extended
		(r >= 0x31C0 && r <= 0x31EF) ||    // CJK Strokes
		(r >= 0x31F0 && r <= 0x31FF) ||    // Katakana Phonetic Extensions
		(r >= 0x3200 && r <= 0x32FF) ||    // Enclosed CJK Letters and Months
		(r >= 0x3300 && r <= 0x33FF) ||    // CJK Compatibility
		(r >= 0x3400 && r <= 0x4DBF) ||    // CJK Unified Ideographs Extension A
		(r >= 0x4E00 && r <= 0x9FFF) ||    // CJK Unified Ideographs
		(r >= 0xA000 && r <= 0xA48F) ||    // Yi Syllables
		(r >= 0xA490 && r <= 0xA4CF) ||    // Yi Radicals
		(r >= 0xAC00 && r <= 0xD7AF) ||    // Hangul Syllables
		(r >= 0xF900 && r <= 0xFAFF) ||    // CJK Compatibility Ideographs
		(r >= 0xFE10 && r <= 0xFE1F) ||    // Vertical Forms
		(r >= 0xFE30 && r <= 0xFE4F) ||    // CJK Compatibility Forms
		(r >= 0xFE50 && r <= 0xFE6F) ||    // Small Form Variants
		(r >= 0xFF01 && r <= 0xFF60) ||    // Fullwidth ASCII variants
		(r >= 0xFFE0 && r <= 0xFFE6) ||    // Fullwidth symbol variants
		unicode.Is(unicode.Han, r) ||      // Additional Han characters
		unicode.Is(unicode.Hiragana, r) || // Additional Hiragana
		unicode.Is(unicode.Katakana, r)    // Additional Katakana
}