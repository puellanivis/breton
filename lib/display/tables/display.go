package tables

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func (d *Divider) writeDivider(wr io.Writer, widths []int) (n int64, err error) {
	line := new(bytes.Buffer)

	if d.Left != "" {
		line.WriteString(d.Left)
	}

	switch {
	case d.Bar == "":
		var cols []string

		for _, width := range widths {
			if width < 0 {
				width = -width
			}

			cols = append(cols, strings.Repeat(d.Space, width))
		}

		line.WriteString(strings.Join(cols, d.Space))

	case d.Space != "":
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

	if line.Len() == 0 {
		return 0, nil
	}

	line.WriteByte('\n')
	return io.Copy(wr, line)
}

func (d *Divider) scale(width int, text string, fn func(string) int) string {
	if width == 0 || d.Space == "" {
		return text
	}

	l := len(text)

	if fn != nil {
		l = fn(text)
	}

	if width < 0 {
		padding := strings.Repeat(d.Space, -width-l)
		return fmt.Sprint(padding, text)
	}

	padding := strings.Repeat(d.Space, width-l)
	return fmt.Sprint(text, padding)
}

// writeRow buffers the whole line together, then makes a single wr.Write of the rendered row.
func (f *Format) writeRow(wr io.Writer, cols []string) (n int64, err error) {
	line := new(bytes.Buffer)

	if f.Inner == nil {
		// if there is no defined Inner format, just dump it raw.
		for _, col := range cols {
			line.WriteString(col)
		}

		line.WriteByte('\n')
		return io.Copy(wr, line)
	}

	if f.Inner.Left != "" {
		line.WriteString(f.Inner.Left)
		line.WriteString(f.Inner.Space)
	}

	var sep = f.Inner.Space
	if f.Inner.Bar != "" {
		sep = fmt.Sprint(f.Inner.Space, f.Inner.Bar, f.Inner.Space)
	}
	line.WriteString(strings.Join(cols, sep))

	if f.Inner.Right != "" {
		line.WriteString(f.Inner.Space)
		line.WriteString(f.Inner.Right)
	}

	if line.Len() == 0 {
		// if line is empty, don’t perform any Write at all.
		// Otherwise we would put a newline in.
		return 0, nil
	}

	line.WriteByte('\n')
	return io.Copy(wr, line)
}

func (f *Format) writeRowScale(wr io.Writer, row []string, widths []int) (n int64, err error) {
	var cols []string

	if f.Inner.Right == "" {
		w := make([]int, len(row)-1)
		copy(w, widths)
		widths = w
		widths = append(widths, 0)
	}

	for i, width := range widths {
		cols = append(cols, f.Inner.scale(width, row[i], f.WidthFunc))
	}

	return f.writeRow(wr, cols)
}

// WriteSimple takes only a 2D slice to fill in the table. It uses the Default format.
func WriteSimple(wr io.Writer, table Table) error {
	return Default.WriteSimple(wr, table)
}

// WriteSimple takes only a 2D slice to fill in the table.
func (f *Format) WriteSimple(wr io.Writer, table Table) error {
	widths := table.widths(f.Inner.Space != "", f.WidthFunc)

	if f.Upper != nil {
		if _, err := f.Upper.writeDivider(wr, widths); err != nil {
			return err
		}
	}

	for _, row := range table {
		if len(row) < 1 {
			if f.Middle != nil {
				if _, err := f.Middle.writeDivider(wr, widths); err != nil {
					return err
				}
			}

			continue
		}

		if f.Inner.Space == "" {
			if _, err := f.writeRow(wr, row); err != nil {
				return err
			}

			continue
		}

		if f.Inner.Right != "" && len(row) < len(widths) {
			// in this case, we need to pad the available rows to match the rest of the table.
			l := len(widths) - len(row)
			row = append(row, make([]string, l)...)
		}

		if _, err := f.writeRowScale(wr, row, widths); err != nil {
			return err
		}
	}

	if f.Lower != nil {
		if _, err := f.Lower.writeDivider(wr, widths); err != nil {
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
