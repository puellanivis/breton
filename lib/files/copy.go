package files

import (
	"context"
	"io"
	"time"
)

const defaultBufferSize = 32*1024

type noopObserver struct{}

func (o *noopObserver) Observe(v float64) { }

// Copy is a context aware version of io.Copy.
// Do not use to Discard a reader, as a canceled context would stop the read, and it would not be fully discarded.
func Copy(ctx context.Context, dst io.Writer, src io.Reader, opts ...CopyOption) (written int64, err error) {
	c := &copyConfig{
		bwObserver: &noopObserver{},
	}

	for _, opt := range opts {
		// intentionally throwing away the reverting functions.
		_ = opt(c)
	}

	if c.buffer == nil {
		// we allocate a buffer to use as a temporary buffer, rather than alloc new every time.
		c.buffer = make([]byte, defaultBufferSize)
	}

	l := int64(len(c.buffer))

	for {
		done := make(chan struct{})

		ctx := ctx // shadow context intentionally, we might set a timeout later
		cancel := func() { } // noop cancel

		if c.runningTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, c.runningTimeout)
		}

		w := &deadlineWriter{
			ctx: ctx,
			w: dst,
		}
		r := io.LimitReader(src, l)

		var n int64
		var dur time.Duration

		go func() {
			defer close(done)

			start := time.Now()

			n, err = io.CopyBuffer(w, r, c.buffer)

			dur = time.Since(start)

			if n < l && err == nil {
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

		// n and err are valid here because <-done HAPPENS AFTER close(done)
		c.bwObserver.Observe(float64(n) / dur.Seconds())

		written += n
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		return written, nil
	}

	return written, err
}

// CopyWithRunningTimeout performs a series of io.CopyN calls that each has the given timeout.
// So, this function allows you to copy a continuous stream of data, and yet respond to a disconnect/timeout event.
//
// Example: SHOUTcast streamers use a continuous open HTTP request to transfer data,
// and setting any http.Client.Timeout will limit the whole io.Copy, rather than just each individual Read,
// meaning that eventually the Timeout will be met, and the io.Copy will error with a exceeded deadline.
func CopyWithRunningTimeout(ctx context.Context, dst io.Writer, src io.Reader, timeout time.Duration) (written int64, err error) {
	return Copy(ctx, dst, src, WithWatchdogTimeout(timeout))
}
