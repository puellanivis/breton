// Package tables formats multi-dimensional slices of strings to be formated in formated tables.
//
// Defining a table is relatively easy:
//	var t tables.Table
//	t = tables.Append(t, 1, 2, 3, 4)
//	t = tables.Append(t)
//	t = tables.Append(t, "any", "arbitrary", "data", 123)
//	t = tables.Append(t, "a", "b", "c")
//
// tables.Empty will produce:
//	1   2         3    4
//	any arbitrary data 123
//	a   b         c
// N.B. the last elements of every row will not have any ending whitespace added to match other row lengths
//
// tables.ASCII will produce:
//	+-----+-----------+------+-----+
//	| 1   | 2         | 3    | 4   |
//	+-----+-----------+------+-----+
//	| any | arbitrary | data | 123 |
//	| a   | b         | c    |     |
//	+-----+-----------+------+-----+
//
// tables.Unicode will produce:
//	┌─────┬───────────┬──────┬─────┐
//	│ 1   │ 2         │ 3    │ 4   │
//	├─────┼───────────┼──────┼─────┤
//	│ any │ arbitrary │ data │ 123 │
//	│ a   │ b         │ c    │     │
//	└─────┴───────────┴──────┴─────┘
//
// tables.HTML will produce:
//	<table>
//	<tr><td class="first">1</td><td>2</td><td>3</td><td>4</td></tr>
//	<tr><td class="first">any</td><td>arbitrary</td><td>data</td><td>123</td></tr>
//	<tr><td class="first">a</td><td>b</td><td>c</td><td></td></tr>
//	</table>
//
package tables

import (
	"bytes"
	"fmt"
)

// Table defines a 2 dimensional table intended to display.
type Table [][]string

// Append places a row of things onto the table, each argument being converted to a string, placed in a column.
func Append(table Table, a ...interface{}) Table {
	var row []string

	for _, val := range a {
		row = append(row, fmt.Sprint(val))
	}

	return append(table, row)
}

// String converts the Table to a string in a buffer, and returns a string there of.
func (t Table) String() string {
	b := new(bytes.Buffer)

	if err := Default.WriteSimple(b, t); err != nil {
		panic(err)
	}

	return b.String()
}

func (t Table) widths(autoscale bool, fn func(string) int) []int {
	if len(t) < 1 {
		return nil
	}

	if fn == nil {
		fn = func(s string) int {
			return len(s)
		}
	}

	l := 0
	for _, row := range t {
		if l < len(row) {
			l = len(row)
		}
	}

	widths := make([]int, l)

	if !autoscale {
		return widths
	}

	for _, row := range t {
		for i, col := range row {
			l := fn(col)

			if widths[i] < l {
				widths[i] = l
			}
		}
	}

	return widths
}
