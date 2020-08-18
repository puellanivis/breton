package files

import (
	"errors"
	"os"
	"syscall"
)

// PathError is DEPRECATED, and returns an *os.PathError with appropriate fields set. DO NOT USE.
//
// This is a stop-gap quick-replace to remove `&os.PathError{ op, path, err }`.
// One should use the direct complex literal instruction instead.
//
// Deprecated: use &os.PathError{} directly.
func PathError(op, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

var (
	// ErrNotSupported should be returned, if a particular feature or option is not supported.
	ErrNotSupported = errors.New("not supported")

	// ErrNotDirectory should be returned, if a request is made to ReadDir a non-directory.
	ErrNotDirectory = syscall.ENOTDIR
)
