// +build dragonflybsd freebsd linux netbsd openbsd solaris darwin

package files

import (
	"net/url"
	"path"
	"strconv"
)

func resolveFileHandle(num string) (uintptr, error) {
	fd, err := strconv.ParseUint(num, 0, strconv.IntSize)
	if err != nil {
		return uintptr(^fd), ErrURLInvalid
	}

	return uintptr(fd), nil
}

func resolveFileURL(uri *url.URL) (string, error) {
	if uri.User != nil {
		return "", ErrURLInvalid
	}

	switch uri.Host {
	case "", "localhost":
	default:
		return "", ErrURLInvalid
	}

	if name := uri.Opaque; name != "" {
		if !path.IsAbs(name) {
			// a path in Opaque must start with `/` and not with `%2f`.
			return "", ErrURLInvalid
		}

		name, err := url.PathUnescape(name)
		if err != nil {
			return "", ErrURLInvalid
		}

		return path.Clean(name), nil
	}

	name := uri.Path
	if !path.IsAbs(name) {
		return "", ErrURLInvalid
	}

	return path.Clean(name), nil
}
