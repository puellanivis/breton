package m3u8

import (
	"bytes"
	"fmt"
	"strconv"
)

// Resolution represents a video resolution of Width x Height.
type Resolution struct {
	Width  int
	Height int
}

// TextUnmarshal implements encoding.TextUnmarshaler.
func (r *Resolution) TextUnmarshal(value []byte) error {
	fields := bytes.Split(value, []byte{'x'})

	if len(fields) != 2 {
		return fmt.Errorf("invalid resolution: %q", value)
	}

	width, err := strconv.Atoi(string(fields[0]))
	if err != nil {
		return fmt.Errorf("invalid width: %q: %v", fields[0], err)
	}

	height, err := strconv.Atoi(string(fields[1]))
	if err != nil {
		return fmt.Errorf("invalid height: %q: %v", fields[1], err)
	}

	r.Width = width
	r.Height = height

	return nil
}

// TextMarshal implements encoding.TextMarshaler.
func (r *Resolution) TextMarshal() ([]byte, error) {
	return []byte(r.String()), nil
}

// String implements fmt.Stringer.
func (r *Resolution) String() string {
	return fmt.Sprintf("%dx%d", r.Width, r.Height)
}
