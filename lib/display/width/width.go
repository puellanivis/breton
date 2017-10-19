package width

import (
	"golang.org/x/text/width"
	"unicode"
)

// AmbiguousIsWide sets if an unknown ambiguous character is assumed to be width == 2.
var AmbiguousIsWide = false

func guessWidth(r rune) int {
	if unicode.Is(unicode.Cc, r) {
		// NUL is width 0, all other C0/C1 codes shall return -1
		if r == 0 {
			return 0
		}

		return -1
	}

	// assume SOFT HYPHEN has a width of 1 (because monospaced-fonts)
	if r == '\u00AD' {
		return 1
	}

	if unicode.Is(ZeroWidth, r) {
		return 0
	}

	if AmbiguousIsWide || unicode.Is(DoubleWidth, r) {
		return 2
	}

	return 1
}

// Rune returns the expected display width of a rune. The logic largely matches what PuTTY does.
func Rune(r rune) int {
	switch width.LookupRune(r).Kind() {
	case width.EastAsianAmbiguous, width.Neutral:
		return guessWidth(r)

	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2

		// case width.EastAsianNarrow, width.EastAsianHalfwidth, width.EastAsianNeutral:
	}

	return 1
}

// String returns the sum display length expected of the string given.
func String(s string) (n int) {
	for _, r := range s {
		n += Rune(r)
	}

	return n
}
