package tables

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

func (d *Divider) writeDivider(wr io.Writer, widths []int) error {
	var line bytes.Buffer

	if d.Left != "" {
		line.WriteString(d.Left)
	}

	if d.Space != "" {
		var cols []string

		for _, width := range widths {
			if width < 0 {
				width = -width
			}

			cols = append(cols, strings.Repeat(d.Space, width+2))
		}

		line.WriteString(strings.Join(cols, d.Bar))
	}

	if d.Right != "" {
		line.WriteString(d.Right)
	}

	if line.Len() > 0 {
		line.WriteByte('\n')

		_, err := io.Copy(wr, &line)
		return err
	}

	return nil
}

func (d *Divider) scale(width int, text string) string {
	if width == 0 || d.Space == "" {
		return text
	}

	// should count display widths...
	l := utf8.RuneCountInString(text)

	if width < 0 {
		padding := strings.Repeat(d.Space, -width-l)
		return fmt.Sprint(padding, text)
	}

	padding := strings.Repeat(d.Space, width-l)
	return fmt.Sprint(text, padding)
}

func (f *Format) writeRow(wr io.Writer, cols []string) error {
	var line bytes.Buffer

	if f.Inner == nil {
		for _, col := range cols {
			line.WriteString(col)
		}

		line.WriteByte('\n')
		_, err := io.Copy(wr, &line)
		return err
	}

	line.WriteString(f.Inner.Left)
	line.WriteString(f.Inner.Space)

	sep := fmt.Sprint(f.Inner.Space, f.Inner.Bar, f.Inner.Space)
	line.WriteString(strings.Join(cols, sep))

	line.WriteString(f.Inner.Space)
	line.WriteString(f.Inner.Right)

	if line.Len() > 0 {
		line.WriteByte('\n')
		_, err := io.Copy(wr, &line)
		return err
	}

	return nil
}

func (f *Format) writeRowScale(wr io.Writer, row []string, widths []int) error {
	var cols []string

	for i, width := range widths {
		cols = append(cols, f.Inner.scale(width, row[i]))
	}

	return f.writeRow(wr, cols)
}

// WriteSimple takes only a 2D slice to fill in the table. It uses the Default format.
func WriteSimple(wr io.Writer, table Table) error {
	return Default.WriteSimple(wr, table)
}

func (f *Format) bind(wr io.Writer, widths []int, fn func() error) error {
	if f.Upper != nil {
		if err := f.Upper.writeDivider(wr, widths); err != nil {
			return err
		}
	}

	if err := fn(); err != nil {
		return err
	}

	if f.Lower != nil {
		if err := f.Lower.writeDivider(wr, widths); err != nil {
			return err
		}
	}

	return nil
}

// WriteSimple takes only a 2D slice to fill in the table.
func (f *Format) WriteSimple(wr io.Writer, table Table) error {
	widths := table.widths(f.Inner.Space != "")

	return f.bind(wr, widths, func() error {
		return f.writeSimple(wr, table, widths)
	})
}

func (f *Format) writeSimple(wr io.Writer, table Table, widths []int) error {
	// number of rows = 0
	if len(table) < 1 {
		return nil
	}

	if f.Inner.Space == "" {
		for _, row := range table {
			if err := f.writeRow(wr, row); err != nil {
				return err
			}
		}

		return nil
	}

	if f.Middle == nil {
		// special case, absolutely no need to check
		// whether to write a divider.

		for _, row := range table {
			if len(row) < len(widths) {
				l := len(widths) - len(row)
				row = append(row, make([]string, l)...)
			}

			if err := f.writeRowScale(wr, row, widths); err != nil {
				return err
			}
		}

		return nil
	}

	for _, row := range table {
		if len(row) < 1 {
			if err := f.Middle.writeDivider(wr, widths); err != nil {
				return err
			}

			continue
		}

		if len(row) < len(widths) {
			l := len(widths) - len(row)
			row = append(row, make([]string, l)...)
		}

		if err := f.writeRowScale(wr, row, widths); err != nil {
			return err
		}
	}

	return nil
}

// WriteMulti writes the data given out it uses the Default format
func WriteMulti(wr io.Writer, table [][][]string) error {
	return Default.WriteMulti(wr, table)
}

func rowHeight(row [][]string) int {
	var height int

	for _, col := range row {
		if height < len(col) {
			height = len(col)
		}
	}

	return height
}

// WriteMulti writes the table given out according to the Format. Each row is checked for height > 1, and if true, it inserts additional lines in a normal 2D Table, populated with available data, and with empty cells for columns where its height is less than the height of the row. It also inserts a new empty row after any row that is height > 1 if the next row is height > 1.
func (f *Format) WriteMulti(wr io.Writer, table [][][]string) error {
	var tbl Table
	var last int

	for _, row := range table {
		height := rowHeight(row)

		if height > 1 || last > 1 {
			tbl = append(tbl, []string{})
		}

		for i := 0; i < height; i++ {
			var r []string

			for _, col := range row {
				var c string

				if i < len(col) {
					c = col[i]
				}

				r = append(r, c)
			}

			tbl = append(tbl, r)
		}

		last = height
	}

	return f.WriteSimple(wr, tbl)
}
