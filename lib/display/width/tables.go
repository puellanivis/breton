package width

import (
	"golang.org/x/text/unicode/rangetable"
	"unicode"
)

// AdditionalZeroWidth is a Unicode Range of expected zero-width glyphs outside of Cf, Mn, and Me.
var AdditionalZeroWidth = &unicode.RangeTable{
	R16: []unicode.Range16{
		{0x1160, 0x11FF, 1}, // Hangul Jamo medial vowels and final consonants.
	},
}

// ZeroWidth is a Unicode Range that can be expected to be zero-width glyphs on a terminal.
var ZeroWidth = rangetable.Merge(unicode.Cf, unicode.Mn, unicode.Me, AdditionalZeroWidth)

// DoubleWidth is a Unicode Range that can be safely assumed to be Wide, even if not EastAsian{Wide,Fullwidth}.
var DoubleWidth = &unicode.RangeTable{
	R32: []unicode.Range32{
		{0x1F030, 0x1F061, 1}, // Domino Tiles (horizontal)
		{0x1F100, 0x1F1FF, 1}, // Enclosed Alphanumeric Supplement
		{0x1F200, 0x1F2FF, 1}, // Enclosed Ideographic Supplement
		{0x1F300, 0x1F5FF, 1}, // Miscellaneous Symbols and Pictographs
		{0x1F600, 0x1F64F, 1}, // Emoticons
		{0x1F680, 0x1F6FF, 1}, // Transport and Map Symbols
		{0x1F900, 0x1F9FF, 1}, // Supplemental Symbols and Pictographs
	},
}
