// Package tables formats multi-dimensional slices of strings to be formated in formated tables.
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

func (t Table) widths(autoscale bool) []int {
	if len(t) < 1 {
		return nil
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
			if widths[i] < len(col) {
				widths[i] = len(col)
			}
		}
	}

	return widths
}
