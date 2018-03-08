package files

import (
	"context"
	"io"
	"time"
)

const defaultBufferSize = 32 * 1024

// Copy is a context aware version of io.Copy.
// Do not use to Discard a reader, as a canceled context would stop the read, and it would not be fully discarded.
func Copy(ctx context.Context, dst io.Writer, src io.Reader, opts ...CopyOption) (written int64, err error) {
	c := new(copyConfig)

	for _, opt := range opts {
		// intentionally throwing away the reverting functions.
		_ = opt(c)
	}

	var observe func(float64)
	if c.bwObserver != nil {
		if c.bwScale < 1 {
			c.bwScale = 1
		}
			
		observe = c.bwObserver.Observe
	}

	if c.buffer == nil {
		// we allocate a buffer to use as a temporary buffer, rather than alloc new every time.
		c.buffer = make([]byte, defaultBufferSize)
	}

	l := int64(len(c.buffer))

	var bwAccum int64
	last := time.Now()
	next := last.Add(time.Second)

	for {
		done := make(chan struct{})

		ctx := ctx          // shadow context intentionally, we might set a timeout later
		cancel := func() {} // noop cancel

		if c.runningTimeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, c.runningTimeout)
		}

		w := &deadlineWriter{
			ctx: ctx,
			w:   dst,
		}
		r := io.LimitReader(src, l)

		var n int64

		go func() {
			defer close(done)

			n, err = io.CopyBuffer(w, r, c.buffer)

			if n < l && err == nil {
				err = io.EOF
			}
		}()

		select {
		case <-done:
			cancel()
			written += n
			bwAccum += n

		case <-ctx.Done():
			cancel()
			return written, ctx.Err()
		}

		if observe != nil {
			if now := time.Now(); now.After(next) {
				dur := now.Sub(last)

				// n and err are valid here because <-done HAPPENS AFTER close(done)
				observe(float64(bwAccum * c.bwScale) / dur.Seconds())

				bwAccum = 0
				last = now
				next = last.Add(time.Second)
			}
		}

		if err != nil {
			break
		}
	}

	if err == io.EOF {
		return written, nil
	}

	return written, err
}
