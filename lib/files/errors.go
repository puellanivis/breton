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

type invalidURLError struct {
	s string
}

func (e *invalidURLError) Error() string {
	return e.s
}

func (e *invalidURLError) Unwrap() error {
	if e == ErrURLInvalid {
		return os.ErrInvalid
	}

	return ErrURLInvalid
}

func (e *invalidURLError) Is(target error) bool {
	switch target {
	case ErrURLInvalid, os.ErrInvalid:
		return true
	}

	return e == target
}

// NewInvalidURLError returns an error that formats as the given text,
// and where errors.Is will return true for both: files.ErrURLInvalid and os.ErrInvalid.
func NewInvalidURLError(reason string) error {
	return &invalidURLError{
		s: reason,
	}
}

// A set of Invalid URL Error to better identify and relate specific invalid URL details.
var (
	ErrURLInvalid = NewInvalidURLError("invalid url")

	ErrURLCannotHaveAuthority = NewInvalidURLError("invalid url: scheme cannot have an authority")
	ErrURLCannotHaveHost      = NewInvalidURLError("invalid url: scheme cannot have a host in authority")

	ErrURLHostRequired = NewInvalidURLError("invalid url: scheme requires non-empty host in authority")
	ErrURLPathRequired = NewInvalidURLError("invalid url: scheme requires non-empty path")
)
