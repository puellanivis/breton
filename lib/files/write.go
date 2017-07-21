package files

import (
	"context"
	"io"
)

// WriteTo will write the given data to the io.WriteCloser and Close the writer.
func WriteTo(w io.WriteCloser, data []byte) error {
	n, err := w.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := w.Close(); err == nil {
		err = err1
	}
	return err
}

// Write will Create the given filename with the Context, and write the given data to it.
func Write(ctx context.Context, filename string, data []byte) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteTo(f, data)
}

// WriteFrom writes the entire contents of an io.Reader into the filename.
func WriteFrom(ctx context.Context, filename string, r io.Reader) (n int64, err error) {
	f, err := Create(ctx, filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(f, r)
}
