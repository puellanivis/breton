package width

import (
	"testing"
)

func TestZeroWidth(t *testing.T) {
	if l := Rune('\n'); l != -1 {
		t.Error("Expected that Cc characters are width == -1, but width of NEWLINE gave instead ", l)
	}

	if l := Rune('\u200B'); l != 0 {
		t.Error("Expected that Cf characters are width == 0, but width of ZERO WIDTH SPACE gave instead ", l)
	}

	if l := Rune('\u0304'); l != 0 {
		t.Error("Expected that Mn characters are width == 0, but width of COMBINING MACRON gave instead ", l)

	}
}

func TestSingleWidth(t *testing.T) {
	if l := Rune('a'); l != 1 {
		t.Error("Expected that standard ASCII characters are width == 1, but width of 'a' gave instead ", l)
	}

	if l := Rune('á'); l != 1 {
		t.Error("Expected that extended ASCII characters are width == 1, but width of 'á' gave instead ", l)
	}

	if l := Rune('\u27E6'); l != 1 {
		t.Error("Expected that East_Asian_Narrow characters are width == 1, but width of MATHEMATICAL LEFT WHITE SQUARE BRACKET gave instead ", l)
	}

	if l := Rune('\u20A9'); l != 1 {
		t.Error("Expected that East_Asian_Halfwidth characters are width == 1, but width of WON SIGN gave instead ", l)
	}

	if l := Rune('\u0298'); l != 1 {
		t.Error("Expected that East_Asian_Neutral characters are width == 1, but width of LATIN LETTER BILABIAL CLICK  gave instead ", l)
	}
}

func TestDoubleWidth(t *testing.T) {
	if l := Rune('\uFF01'); l != 2 {
		t.Error("Expected that East_Asian_Fullwidth characters are width == 2, but width of FULLWIDTH EXCLAMATION MARK gave instead ", l)
	}

	if l := Rune('\u30A2'); l != 2 {
		t.Error("Expected that East_Asian_Wide characters are width == 2, but width of KATAKANA LETTER A gave instead ", l)
	}
}

func TestAmbiguousWidth(t *testing.T) {
	if l := Rune('\u0398'); l != 1 {
		t.Error("Expected that GREEK CAPITAL LETTER THETA should default width == 1, but gave instead ", l)
	}

	AmbiguousIsWide = true

	if l := Rune('\u2227'); l != 2 {
		t.Error("Expected that East_Asian_Ambiguous characters are width == 2 when AmibiguousIsWidth == true, but width of LOGICAL AND gave instead ", l)
	}

	AmbiguousIsWide = false

	if l := Rune('\u2227'); l != 1 {
		t.Error("Expected that East_Asian_Ambiguous characters are width == 1 when AmibiguousIsWidth == false, but width of LOGICAL OR gave instead ", l)
	}
}

func TestString(t *testing.T) {
	if l := String("þ\u0300á\u0398"); l != 3 {
		t.Error("Test string had wrong length, expected 3, got ", l)
	}
}
