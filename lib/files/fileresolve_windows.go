package files

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func resolveFileURL(uri *url.URL) (string, error) {
	if uri.User != nil {
		return "", ErrURLInvalid
	}

	if name := uri.Opaque; name != "" {
		if !path.IsAbs(name) {
			// a path in Opaque must start with `/` and not with `%2f`.
			return "", ErrURLInvalid
		}

		name = strings.TrimPrefix(name, "/")

		name, err := url.PathUnescape(name)
		if err != nil {
			return "", ErrURLInvalid
		}

		if !filepath.IsAbs(name) {
			return "", ErrURLInvalid
		}

		return filepath.Clean(filepath.FromSlash(name)), nil
	}

	name := uri.Path
	if !path.IsAbs(name) {
		return "", ErrURLInvalid
	}

	switch uri.Host {
	case "", ".":
		name = strings.TrimPrefix(name, "/")

		if !filepath.IsAbs(name) {
			return "", ErrURLInvalid
		}

		return filepath.Clean(filepath.FromSlash(name)), nil
	}

	name = filepath.Clean(filepath.FromSlash(name))

	return `\\` + uri.Host + name, nil
}
