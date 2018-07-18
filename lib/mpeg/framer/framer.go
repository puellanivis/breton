package framer

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

// Various MIME-Types for Audo/Video types.
const (
	AAC  = "audio/x-aac"
	MP3  = "audio/mp3"
	H264 = "video/h264"
)

// Scanner implemnets bufio.Scanner breaking up certain Audio/Video types into packet frames.
type Scanner struct {
	mediaType   string
	frameDetect []byte

	*bufio.Scanner
}

// MediaType returns what MIME-Type was detected by the Scanner.
func (s *Scanner) MediaType() string {
	if len(s.frameDetect) <= 0 {
		return "unknown/unknown"
	}

	return s.mediaType
}

// GetPCR returns the current PCR from the Scanner, extracting the information from the MPEG frame header.
//
// TODO: currently unimplemented.
func (s *Scanner) GetPCR() int64 {
	return 0
}

var (
	syncWordMP3nocrc = []byte{0xFF, 0xFA}
	syncWordMP3      = []byte{0xFF, 0xFB}

	syncWordAACnocrc = []byte{0xFF, 0xF0}
	syncWordAAC      = []byte{0xFF, 0xF1}

	syncWordNAL = []byte{0x00, 0x00, 0x00, 0x01}
)

// DetectContentType is a dropin replacement and wrapper for http.DetectContentType,
// which additionally detects certain Audio/Video types.
//
// Before defering to http.DetectContentType, it looks for certain MPEG sync words supported by this library.
func DetectContentType(data []byte) string {
	switch {
	// MPEG ADTS wrapped MP3
	case bytes.Equal(data[:2], syncWordMP3):
		return MP3

	// MPEG ADTS wrapped AAC
	case bytes.Equal(data[:2], syncWordAAC):
		return AAC

	// H264 byte stream.
	case bytes.Equal(data[:4], syncWordNAL):
		return H264
	}

	return http.DetectContentType(data)
}

func (s *Scanner) splitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(s.frameDetect) <= 0 {
		if len(data) < 4 {
			return 0, nil, nil
		}

		switch {
		// MPEG ADTS wrapped MP3
		case bytes.Equal(data[:2], syncWordMP3), bytes.Equal(data[:2], syncWordMP3nocrc):
			s.mediaType = MP3
			s.frameDetect = append([]byte{}, data[:2]...)

		// MPEG ADTS wrapped AAC
		case bytes.Equal(data[:2], syncWordAAC), bytes.Equal(data[:2], syncWordAACnocrc):
			s.mediaType = AAC
			s.frameDetect = append([]byte{}, data[:2]...)

		// H264 byte stream.
		case bytes.Equal(data[:4], syncWordNAL):
			s.mediaType = H264
			s.frameDetect = append([]byte{}, data[:4]...)

		default:
			return 0, nil, errors.New("unable to detect type")
		}
	}

	advance = 1

	if len(data) < 1 && atEOF {
		return 0, nil, io.EOF
	}

	for j := 0; ; advance++ {
		j = bytes.Index(data[advance:], s.frameDetect)
		if j < 0 {
			if atEOF {
				return len(data), data, nil
			}

			// need more data
			return 0, nil, nil
		}

		advance += j
		return advance, data[:advance], nil
	}
}

// NewScanner returns a new Scanner, which implements bufio.Scanner.
// The Scanner reads from the given io.Reader, and chunks the data into packet frames.
func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{
		Scanner: bufio.NewScanner(r),
	}
	s.Split(s.splitter)
	return s
}
