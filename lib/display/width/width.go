package width

import (
	"unicode"

	"golang.org/x/text/width"
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
//
// Control Characters other than NUL return a value of -1.
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
//
// If any rune in the string returns a width of -1, this function will return -1.
func String(s string) int {
	var n int

	for _, r := range s {
		w := Rune(r)
		if w < 0 {
			return -1
		}

		n += w
	}

	return n
}
