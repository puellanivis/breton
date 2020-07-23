// +build dragonflybsd freebsd linux netbsd openbsd solaris darwin

package files

import (
	"net/url"
	"os"
	"path"
)

func resolveFileURL(uri *url.URL) (string, error) {
	if uri.User != nil {
		return "", os.ErrInvalid
	}

	switch uri.Host {
	case "", "localhost":
	default:
		return "", os.ErrInvalid
	}

	if uri.Path == "" {
		name := uri.Opaque
		if !path.IsAbs(name) {
			return "", os.ErrInvalid
		}

		name, err := url.PathUnescape(uri.Opaque)
		if err != nil {
			return "", os.ErrInvalid
		}

		return name, nil
	}

	name := uri.Path
	if !path.IsAbs(name) {
		return "", os.ErrInvalid
	}

	return path.Clean(name), nil
}
