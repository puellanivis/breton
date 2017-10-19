package width

import (
	"unicode"
	"golang.org/x/text/width"
)

var AmbiguousIsWide = false

func Rune(r rune) int {
	if unicode.In(r, unicode.Cc, unicode.Cf, unicode.Mn) {
		return 0
	}

	switch width.LookupRune(r).Kind() {
	case width.EastAsianAmbiguous:
		if AmbiguousIsWide {
			return 2
		}

	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2

	// case width.EastAsianNarrow, width.EastAsianHalfwidth, width.EastAsianNeutral:
	}

	return 1
}

func String(s string) (n int) {
	for _, r := range s {
		n += Rune(r)
	}

	return n
}
