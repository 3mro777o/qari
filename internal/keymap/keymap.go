// Package keymap lets single-key shortcuts work on the Arabic keyboard layout.
package keymap

// arabicKeys maps what the standard Arabic layout (xkb "ara") types to the
// US key in the same position, so shortcuts work without switching layouts.
// Terminals only report characters, not key positions, hence the table.
var arabicKeys = map[rune]rune{
	'ض': 'q', 'ص': 'w', 'ث': 'e', 'ق': 'r', 'ف': 't', 'غ': 'y', 'ع': 'u', 'ه': 'i', 'خ': 'o', 'ح': 'p', 'ج': '[', 'د': ']',
	'ش': 'a', 'س': 's', 'ي': 'd', 'ب': 'f', 'ل': 'g', 'ا': 'h', 'ت': 'j', 'ن': 'k', 'م': 'l', 'ك': ';', 'ط': '\'',
	'ئ': 'z', 'ء': 'x', 'ؤ': 'c', 'ر': 'v', 'ى': 'n', 'ة': 'm', 'و': ',', 'ز': '.', 'ظ': '/',
	// Shifted keys that differ from US: shift+a, shift+g, shift+/, shift+, and shift+.
	'ِ': 'A', 'ﻷ': 'G', '؟': '?', ',': '<', '.': '>',
	'١': '1', '٢': '2', '٣': '3', '٤': '4', '٥': '5',
}

// LatinKey returns the US-layout rune for r, or r itself.
func LatinKey(r rune) rune {
	if l, ok := arabicKeys[r]; ok {
		return l
	}
	return r
}
