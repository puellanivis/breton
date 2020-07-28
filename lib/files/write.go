package files

import (
	"context"
	"io"
)

// WriteTo will write the given data to the io.Writer,
// it will then Close the writer if it implements io.Closer.
//
// Note: this function will return io.ErrShortWrite,
// if the amount written is less than the data given as input.
func WriteTo(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}

	if c, ok := w.(io.Closer); ok {
		if err2 := c.Close(); err == nil {
			err = err2
		}
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
