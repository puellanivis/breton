package files

import (
	"context"
	"io"
)

func WriteAndClose(w io.WriteCloser, data []byte) error {
	n, err := w.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := w.Close(); err == nil {
		err = err1
	}
	return err
}

func WriteFile(ctx context.Context, filename string, data []byte) error {
	f, err := Create(ctx, filename)
	if err != nil {
		return err
	}

	return WriteAndClose(f, data)
}
