package files

import (
	"os"
	"time"
)

// Option is a function that applies a specific option to a files.File, it
// returns an Option and and error. If error is not nil, then the Option
// returned will revert the option that was set. Since errors returned by
// Option arguments are discarded by Open(), and Create(), if you
// care about the error status of an Option you must apply it yourself
// after Open() or Create()
type Option func(File) (Option, error)

// WithFileMode returns an Option that will set the files.File.Stat().FileMode() to the given os.FileMode.
func WithFileMode(mode os.FileMode) Option {
	type chmoder interface {
		Chmod(os.FileMode) error
	}

	return func(f File) (Option, error) {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}

		save := fi.Mode()

		switch f := f.(type) {
		case chmoder:
			if err := f.Chmod(mode); err != nil {
				return nil, err
			}

		default:
			return nil, ErrNotSupported
		}

		return WithFileMode(save), nil
	}
}

type observer interface {
	Observe(float64)
}

type copyConfig struct {
	runningTimeout time.Duration
	buffer         []byte

	bwScale    float64
	bwCount    int
	bwInterval time.Duration
	bwRunning  observer
	bwLifetime observer
}

// CopyOption defines a function that applies a value or setting for a specific files.Copy operation.
type CopyOption func(c *copyConfig) CopyOption

// WithWatchdogTimeout sets a running interval timeout,
// where if no copy progress is made during that time,
// the files.Copy will fail with a timeout error.
func WithWatchdogTimeout(timeout time.Duration) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.runningTimeout

		c.runningTimeout = timeout

		return WithWatchdogTimeout(save)
	}
}

// WithBuffer sets which buffer a files.Copy should use internally as temporary storage between the Read and Write.
func WithBuffer(buf []byte) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.buffer

		c.buffer = buf

		return WithBuffer(save)
	}
}

// WithBufferSize makes a new buffer of size bytes, which is used as temporary intermediate storage for the files.Copy.
func WithBufferSize(size int) CopyOption {
	if size < 0 {
		panic("cannot use a negative buffer size!")
	}

	return WithBuffer(make([]byte, size))
}

// WithMetricsScale sets the scale of reported Metrics, otherwise it is reported in bytes/second.
func WithMetricsScale(scale float64) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.bwScale

		c.bwScale = scale

		return WithMetricsScale(save)
	}
}

// WithBandwidthMetrics establishes a lifetime bandwidth metric for the files.Copy.
func WithBandwidthMetrics(total interface{ Observe(float64) }) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.bwLifetime

		c.bwLifetime = total

		return WithBandwidthMetrics(save)
	}
}

// WithIntervalBandwidthMetrics keeps a running list of n intervals and reports the bandwidth over this window.
func WithIntervalBandwidthMetrics(running interface{ Observe(float64) }, n int, interval time.Duration) CopyOption {
	return func(c *copyConfig) CopyOption {
		saveOb := c.bwRunning
		saveCount := c.bwCount
		saveDur := c.bwInterval

		c.bwRunning = running
		c.bwCount = n
		c.bwInterval = interval

		return WithIntervalBandwidthMetrics(saveOb, saveCount, saveDur)
	}
}
