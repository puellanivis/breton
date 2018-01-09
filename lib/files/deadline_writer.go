package files

import (
	"context"
	"io"
)

type deadlineWriter struct {
	ctx context.Context
	w   io.Writer
}

func (w *deadlineWriter) Write(b []byte) (n int, err error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}

	return w.w.Write(b)
}
