package files

import (
	"context"
	"errors"
	"io"
	"time"
)

const defaultBufferSize = 64 * 1024

// ErrWatchdogExpired is returned by files.Copy, if the watchdog time expires during a read.
var ErrWatchdogExpired error = watchdogExpiredError{}

type watchdogExpiredError struct{}

func (watchdogExpiredError) Error() string   { return "watchdog expired" }
func (watchdogExpiredError) Timeout() bool   { return true }
func (watchdogExpiredError) Temporary() bool { return true }

// Copy is a context aware version of io.Copy.
// Do not use to Discard a reader, as a canceled context would stop the read, and it would not be fully discarded.
func Copy(ctx context.Context, dst io.Writer, src io.Reader, opts ...CopyOption) (written int64, err error) {
	if dst == nil {
		return 0, errors.New("nil io.Writer passed to files.Copy")
	}

	c := new(copyConfig)

	for _, opt := range opts {
		// intentionally throwing away the reverting functions.
		_ = opt(c)
	}

	if c.buffer == nil {
		// we allocate a buffer to use as a temporary buffer, rather than alloc new every time.
		c.buffer = make([]byte, defaultBufferSize)
	}
	buflen := int64(len(c.buffer))

	var keepingMetrics bool

	var total func(float64)
	if c.bwLifetime != nil {
		total = c.bwLifetime.Observe
		keepingMetrics = true
	}

	if c.bwScale <= 0 {
		c.bwScale = 1
	}

	if c.bwInterval <= 0 {
		c.bwInterval = 1 * time.Second
	}

	type bwSnippet struct {
		n int64
		d time.Duration
	}
	var bwWindow []bwSnippet

	var running func(float64)
	if c.bwRunning != nil {
		if c.bwCount < 1 {
			c.bwCount = 1
		}

		running = c.bwRunning.Observe
		keepingMetrics = true
		bwWindow = make([]bwSnippet, c.bwCount)
	}

	// Prevent an accidental write outside of returning from this function.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	w := &deadlineWriter{
		ctx: ctx,
		w:   dst,
	}

	r := &fuzzyLimitedReader{
		R: src,
		N: buflen,
	}

	t := time.NewTimer(c.runningTimeout)
	if c.runningTimeout <= 0 {
		if !t.Stop() {
			<-t.C
		}
	}

	start := time.Now()

	var bwAccum int64
	last := start
	next := last.Add(c.bwInterval)

	for {
		r.N = buflen // reset io.LimitedReader

		if c.runningTimeout > 0 {
			if !t.Stop() {
				<-t.C
			}
			t.Reset(c.runningTimeout)
		}

		var n int64
		done := make(chan struct{})
		go func() {
			defer close(done)

			n, err = io.CopyBuffer(w, r, c.buffer)

			if n < buflen && err == nil {
				err = io.EOF
			}
		}()

		select {
		case <-done:

		case <-t.C:
			return written, ErrWatchdogExpired

		case <-ctx.Done():
			return written, ctx.Err()
		}

		// n and err are valid here because <-done HAPPENS AFTER close(done)
		written += n
		bwAccum += n
		if err != nil {
			break
		}

		if keepingMetrics {
			if now := time.Now(); now.After(next) {
				if total != nil {
					dur := now.Sub(start)
					total(float64(written) * c.bwScale / dur.Seconds())
				}

				if running != nil {
					dur := now.Sub(last)

					copy(bwWindow, bwWindow[1:])
					bwWindow[len(bwWindow)-1].n = bwAccum
					bwWindow[len(bwWindow)-1].d = dur

					var n int64
					var d time.Duration
					for i := range bwWindow {
						n += bwWindow[i].n
						d += bwWindow[i].d
					}

					running(float64(n) * c.bwScale / d.Seconds())
				}

				bwAccum = 0
				last = now
				next = last.Add(time.Second)
			}
		}
	}

	if err == io.EOF {
		return written, nil
	}

	return written, err
}
