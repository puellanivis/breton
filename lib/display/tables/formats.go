package tables

import (
	"os"
	"strings"
)

// Divider defines a set of dividers to be used when printing a single specific row.
type Divider struct {
	Left  string // The left-side of the row e.g. "|" for ASCII
	Space string // The character used for spacing.
	Bar   string // The separator between two colums, e.g. " |" for ASCII
	Right string // The right-side of the row e.g. "|" for ASCII
}

// Format defines a set of dividers for all types of rows.
type Format struct {
	Upper  *Divider // The top    e.g. ┌---+---+---┐
	Inner  *Divider // Content    e.g. | x | y | z |
	Middle *Divider // The middle e.g. ├---+---+---┤
	Lower  *Divider // the bottom e.g. └---+---+---┘
}

var (
	// Empty does nothing but put a single space between
	// each column, and then autoscale each column to line up.
	Empty = &Format{
		Inner: &Divider{
			Bar: " ",
		},
	}

	// ASCII uses ASCII line-drawing characters, i.e. |, +, and -
	ASCII = &Format{
		Upper: &Divider{
			Left:  "+",
			Space: "-",
			Bar:   "+",
			Right: "+",
		},
		Inner: &Divider{
			Left:  "|",
			Space: " ",
			Bar:   "|",
			Right: "|",
		},
		Middle: &Divider{
			Left:  "+",
			Space: "-",
			Bar:   "+",
			Right: "+",
		},
		Lower: &Divider{
			Left:  "+",
			Space: "-",
			Bar:   "+",
			Right: "+",
		},
	}

	// Unicode uses Unicode line-drawing characters.
	Unicode = &Format{
		Upper: &Divider{
			Left:  "┌",
			Space: "─",
			Bar:   "┬",
			Right: "┐",
		},
		Inner: &Divider{
			Left:  "│",
			Space: " ",
			Bar:   "│",
			Right: "│",
		},
		Middle: &Divider{
			Left:  "├",
			Space: "─",
			Bar:   "┼",
			Right: "┤",
		},
		Lower: &Divider{
			Left:  "└",
			Space: "─",
			Bar:   "┴",
			Right: "┘",
		},
	}

	// HTML outputs a _very_ simple HTML table.
	HTML = &Format{
		Inner: &Divider{
			Left:  "<tr><td class=\"first\">",
			Bar:   "</td><td>",
			Right: "</td></tr>",
		},
		Upper: &Divider{
			Left: "<table>",
		},
		Lower: &Divider{
			Left: "</table>",
		},
	}
)

var (
	// Default is the default Format to use. This starts out defaulting to
	// tables.ASCII, but if environment variables LC_ALL or LANG ends
	// with .UTF-8, at runtime, then it switches to tables.Unicode
	// by default. (LC_ALL overrides LANG)
	Default = ASCII
)

func init() {
	if lang := os.Getenv("LC_ALL"); lang != "" {
		if strings.HasSuffix(lang, ".UTF-8") {
			Default = Unicode
		}

		return
	}

	if lang := os.Getenv("LANG"); strings.HasSuffix(lang, ".UTF-8") {
		Default = Unicode
	}
}
