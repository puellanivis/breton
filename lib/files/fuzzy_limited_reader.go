package files

import (
	"io"
)

// fuzzyLimitedReader reads at least N-bytes from the underlying reader.
// It does not ensure that it reads only or at-most N-bytes.
// Each call to Read updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type fuzzyLimitedReader struct {
	R io.Reader // underlying reader
	N int64     // stop reading after at least this much
}

func (r *fuzzyLimitedReader) Read(b []byte) (n int, err error) {
	if r.N <= 0 {
		return 0, io.EOF
	}

	n, err = r.R.Read(b)
	r.N -= int64(n)

	return n, err
}
