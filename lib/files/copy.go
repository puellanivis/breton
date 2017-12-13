package files

import (
	"context"
	"io"
	"time"
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

const copyBufferSize = 32*1024

// Copy is a context aware version of io.Copy.
// Do not use to Discard a reader, as a canceled context would stop the read, and it would not be fully discarded.
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (written int64, err error) {
	// we allocate a buffer to use as a temporary buffer, rather than alloc new every time.
	buf := make([]byte, copyBufferSize)

	for {
		done := make(chan struct{})

		w := &deadlineWriter{
			ctx: ctx,
			w: dst,
		}

		var n int64
		go func() {
			defer close(done)

			n, err = io.CopyBuffer(w, io.LimitReader(src, copyBufferSize), buf)

			if n < copyBufferSize && err == nil {
				err = io.EOF
			}
		}()

		select {
		case <-done:
		case <-ctx.Done():
			return written, ctx.Err()
		}

		written += n
		if err != nil {
			if err == io.EOF {
				return written, nil
			}

			return written, err
		}
	}

}

// CopyWithRunningTimeout performs a series of io.CopyN calls that each has the given timeout.
// So, this function allows you to copy a continuous stream of data, and yet respond to a disconnect/timeout event.
//
// Example: SHOUTcast streamers use a continuous open HTTP request to transfer data,
// and setting any http.Client.Timeout will limit the whole io.Copy, rather than just each individual Read,
// meaning that eventually the Timeout will be met, and the io.Copy will error with a exceeded deadline.
func CopyWithRunningTimeout(ctx context.Context, dst io.Writer, src io.Reader, timeout time.Duration) (written int64, err error) {
	// we allocate a buffer to use as a temporary buffer, rather than alloc new every time.
	buf := make([]byte, copyBufferSize)

	for {
		done := make(chan struct{})
		ctx, cancel := context.WithTimeout(ctx, timeout)

		w := &deadlineWriter{
			ctx: ctx,
			w: dst,
		}

		var n int64
		go func() {
			defer close(done)

			n, err = io.CopyBuffer(w, io.LimitReader(src, copyBufferSize), buf)

			if n < copyBufferSize && err == nil {
				err = io.EOF
			}
		}()

		select {
		case <-done:
		case <-ctx.Done():
			cancel()
			return written, ctx.Err()
		}

		cancel()
		written += n
		if err != nil {
			if err == io.EOF {
				return written, nil
			}

			return written, err
		}
	}
}
