package files

import (
	"errors"
	"os"
)

// PathError returns an *os.PathError with appropriate fields set. DO NOT USE.
//
// This is a stop-gap quick-replace to remove `&os.PathError{ op, path, err }`.
// One should use the direct complex literal instruction instead.
func PathError(op, path string, err error) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// ErrNotSupported should be returned, if a particular feature or option is not supported.
var ErrNotSupported = errors.New("not supported")
