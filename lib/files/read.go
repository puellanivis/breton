package files

import (
	"context"
	"io"
	"io/ioutil"
)

// ReadFrom reads the entire content of an io.ReadCloser and returns the content as a byte slice.
// It will also Close the reader.
func ReadFrom(r io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err1 := r.Close(); err == nil {
		err = err1
	}
	return b, err
}

// Discard throws away the entire content of an io.ReadCloser and then closes the reader.
// This is specifically not context aware, it is intended to always run to completion.
func Discard(r io.ReadCloser) error {
	_, err := io.Copy(ioutil.Discard, r)

	if err2 := r.Close(); err == nil {
		err = err2
	}

	return err
}

// Read takes a Context and a filename and reads the entire content into a byte-slice which it returns.
func Read(ctx context.Context, filename string) ([]byte, error) {
	f, err := Open(ctx, filename)
	if err != nil {
		return nil, err
	}

	return ReadFrom(f)
}
