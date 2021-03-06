package files

import (
	"context"
	"io"
)

// WriteTo writes the entire content of data to an io.Writer.
// If the Writer also implements io.Closer, it will also Close it.
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

// Write writes the entire content of data to the resource at the given URL.
func Write(ctx context.Context, url string, data []byte) error {
	f, err := Create(ctx, url)
	if err != nil {
		return err
	}

	return WriteTo(f, data)
}
