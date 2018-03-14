package m3u8

import (
	"fmt"
	"strconv"
	"strings"
)

type Resolution struct {
	Width  int
	Height int
}

func (r *Resolution) UnmarshalString(value string) error {
	fields := strings.Split(value, "x")

	if len(fields) != 2 {
		return fmt.Errorf("invalid resolution: %q", value)
	}

	width, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("invalid width: %q: %v", fields[0], err)
	}

	height, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("invalid height: %q: %v", fields[1], err)
	}

	r.Width = width
	r.Height = height

	return nil
}

func (r *Resolution) String() string {
	return fmt.Sprintf("%dx%d", r.Width, r.Height)
}
