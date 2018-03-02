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
			if len(data) < 4 {
				return 0, nil, nil
			}

			switch {
			// MPEG ADTS.
			case data[0] == 0xFF && data[1]&0xF0 == 0xF0:
				frameDetect = append([]byte{}, data[:2]...)

			// H264 byte stream.
			case data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x00 && data[3] == 0x01:
				frameDetect = append([]byte{}, data[:4]...)

			default:
				return 0, nil, errors.New("MPEG sync word missing")
			}

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
			if frameDetect[0] != 0xFF || data[advance+1]&0xF0 == 0xF0 {
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
