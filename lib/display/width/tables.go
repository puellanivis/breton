package width

import (
	"unicode"

	"golang.org/x/text/unicode/rangetable"
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
		{Lo: 0x1F030, Hi: 0x1F061, Stride: 1}, // Domino Tiles (horizontal)
		{Lo: 0x1F100, Hi: 0x1F1FF, Stride: 1}, // Enclosed Alphanumeric Supplement
		{Lo: 0x1F200, Hi: 0x1F2FF, Stride: 1}, // Enclosed Ideographic Supplement
		{Lo: 0x1F300, Hi: 0x1F5FF, Stride: 1}, // Miscellaneous Symbols and Pictographs
		{Lo: 0x1F600, Hi: 0x1F64F, Stride: 1}, // Emoticons
		{Lo: 0x1F680, Hi: 0x1F6FF, Stride: 1}, // Transport and Map Symbols
		{Lo: 0x1F900, Hi: 0x1F9FF, Stride: 1}, // Supplemental Symbols and Pictographs
	},
}
