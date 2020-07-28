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
// DEPRECATED: use &os.PathError{} directly.
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

// NewInvalidURLError returns an error that formats as the given text,
// and errors.Is: files.ErrURLInvalid and os.ErrInvalid.
func NewInvalidURLError(reason string) error {
	return &invalidURLError{
		s: reason,
	}
}

var (
	// ErrURLInvalid should be returned, if the URL is syntactically valid, but semantically invalid.
	ErrURLInvalid = NewInvalidURLError("invalid url")

	// ErrURLCannotHaveAuthority should be returned, if the URL scheme does not allow a non-empty authority section.
	ErrURLCannotHaveAuthority = NewInvalidURLError("invalid url: cannot have authority")

	// ErrURLNoHost should be return, if the URL scheme requires the authority section to not specify a host.
	ErrURLNoHost = NewInvalidURLError("invalid url: scheme cannot have host in authority")

	// ErrURLHostRequired should be returned, if the URL scheme requires a host in the authority section.
	ErrURLHostRequired = NewInvalidURLError("invalid url: scheme requires host in authority")

	// ErrURLPathRequired should be returned, if the URL scheme requires a non-empty path.
	ErrURLPathRequired = NewInvalidURLError("invalid url: scheme requires non-empty path")
)
