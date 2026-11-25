// Package norm folds names into a form where spelling variants compare equal,
// so "العفاسي", "العَفاسي" and "Al-Afasy" can all be matched by search.
package norm

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var arabicFold = strings.NewReplacer(
	"أ", "ا", "إ", "ا", "آ", "ا", "ٱ", "ا",
	"ة", "ه", "ى", "ي", "ئ", "ي", "ؤ", "و",
	"ک", "ك", "ی", "ي", // Persian/Urdu letter forms
)

// Fold lowercases, strips diacritics and punctuation, and unifies Arabic letter
// variants. Words stay separated by single spaces.
func Fold(s string) string {
	s = arabicFold.Replace(norm.NFD.String(s))
	var b strings.Builder
	space := true
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r) || r == 'ـ': // harakat, latin accents, tatweel
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			space = false
		default:
			if !space {
				b.WriteByte(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// Compact is Fold without spaces, for comparing names written as one or two
// words ("عبدالباسط" vs "عبد الباسط", "Al-Afasy" vs "Alafasy").
func Compact(s string) string { return strings.ReplaceAll(Fold(s), " ", "") }

// Match reports whether every word of query occurs in the folded haystack,
// either as a substring or (for Latin names) with "al"/"el" prefixes ignored.
func Match(foldedHaystack, query string) bool {
	q := Fold(query)
	if q == "" {
		return true
	}
	h := foldedHaystack
	hc := strings.ReplaceAll(h, " ", "")
	for _, w := range strings.Fields(q) {
		if strings.Contains(h, w) || strings.Contains(hc, w) {
			continue
		}
		return false
	}
	return true
}
