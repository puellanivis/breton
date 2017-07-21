package files

import (
	"context"
	"io"
	"io/ioutil"
)

// ReadFrom reads the entire content of an io.ReadCloser and returns the content as a byte slice. It will also Close the reader.
func ReadFrom(r io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err1 := r.Close(); err == nil {
		err = err1
	}
	return b, err
}

// Discard throws away the entire content of an io.ReadCloser and closes the reader.
func Discard(r io.ReadCloser) error {
	if _, err := io.Copy(ioutil.Discard, r); err != nil {
		return err
	}

	return r.Close()
}

// Read takes a Context and a filename and reads the entire content into a byte-slice which it returns.
func Read(ctx context.Context, filename string) ([]byte, error) {
	f, err := Open(ctx, filename)
	if err != nil {
		return nil, err
	}

	return ReadFrom(f)
}

// ReadTo reads the entire content of a filename into an io.Writer.
func ReadTo(ctx context.Context, w io.Writer, filename string) (n int64, err error) {
	f, err := Open(ctx, filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(w, f)
}
