package framer

import (
	"bufio"
	"bytes"
	"io"

	"github.com/pkg/errors"
)

func splitterFunc() bufio.SplitFunc {
	firstFrame := true
	var frameDetect []byte

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if firstFrame {
			if len(data) < 2 {
				return 0, nil, nil
			}

			if data[0] != 0xFF && data[1] & 0xF0 != 0xF0 {
				return 0, nil, errors.New("MPEG sync word missing")
			}

			frameDetect = append([]byte{}, data[:2]...)
		}
		firstFrame = false

		advance = 1

		if len(data) < 1 && atEOF {
			return 0, nil, io.EOF
		}

		for j := 0; ; advance++ {
			j = bytes.Index(data[advance:], frameDetect)
			if j < 0 {
				if atEOF {
					return len(data), data, nil
				}

				// need more data
				return 0, nil, nil
			}

			advance += j
			if data[advance+1] & 0xF0 == 0xF0 {
				return advance, data[:advance], nil
			}
		}
	}
}

func NewScanner(r io.Reader) *bufio.Scanner {
	s := bufio.NewScanner(r)
	s.Split(splitterFunc())
	return s
}
