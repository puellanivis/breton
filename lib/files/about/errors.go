package aboutfiles

import (
	"errors"
	"os"
)

type errorURL struct {
	error
}

func (e errorURL) ReadAll() ([]byte, error) {
	return nil, e.error
}

func (e errorURL) ReadDir() ([]os.FileInfo, error) {
	return nil, e.error
}

func (e errorURL) Unwrap() error {
	return e.error
}

// ErrNoSuchHost defines an error, where a DNS host lookup failed to resolve.
var ErrNoSuchHost = errors.New("no such host")
