package files

import (
	"errors"
	"os"
	"time"
)

// ErrNotSupported should be returned when a specific file.File given to an
// Option does not support the Option specified.
var ErrNotSupported = errors.New("option not supported")

// Option is a function that applies a specific option to a files.File, it
// returns an Option and and error. If error is not nil, then the Option
// returned will revert the option that was set. Since errors returned by
// Option arguments are discarded by Open(), and Create(), if you
// care about the error status of an Option you must apply it yourself
// after Open() or Create()
type Option func(File) (Option, error)

// applyOptions is a helper function to apply a range of options on an os.File
func applyOptions(f File, opts []Option) error {
	for _, opt := range opts {
		if _, err := opt(f); err != nil {
			return err
		}
	}

	return nil
}

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
	bufferSize int
	buffer []byte

	bwObserver observer
}

type CopyOption func(c *copyConfig) CopyOption

func WithWatchdogTimeout(timeout time.Duration) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.runningTimeout

		c.runningTimeout = timeout

		return WithWatchdogTimeout(save)
	}
}

func WithBuffer(buf []byte) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.buffer

		c.buffer = buf

		return WithBuffer(save)
	}
}

func WithBufferSize(size int) CopyOption {
	if size < 0 {
		panic("cannot use a negative buffer size!")
	}

	return WithBuffer(make([]byte, size))
}

func WithBandwidthMetrics(observer interface{ Observe(float64) }) CopyOption {
	return func(c *copyConfig) CopyOption {
		save := c.bwObserver

		c.bwObserver = observer

		return WithBandwidthMetrics(save)
	}
}
