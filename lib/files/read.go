package files

import (
	"context"
	"io"
	"io/ioutil"
)

// ReadFrom reads the entire content of an io.Reader and returns the content as a byte slice.
// If the Reader also implements io.Closer, it will also Close it.
func ReadFrom(r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)

	if c, ok := r.(io.Closer); ok {
		if err2 := c.Close(); err == nil {
			err = err2
		}
	}

	return b, err
}

// Discard throws away the entire content of an io.Reader.
// If the Reader also implements io.Closer, it will also Close it.
//
// This is specifically not context aware, it is intended to always run to completion.
func Discard(r io.Reader) error {
	_, err := io.Copy(ioutil.Discard, r)

	if c, ok := r.(io.Closer); ok {
		if err2 := c.Close(); err == nil {
			err = err2
		}
	}

	return err
}

// Read reads the entire content of the resource at the given URL into a byte-slice.
func Read(ctx context.Context, url string) ([]byte, error) {
	f, err := Open(ctx, url)
	if err != nil {
		return nil, err
	}

	return ReadFrom(f)
}
