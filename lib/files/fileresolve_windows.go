package files

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func resolveFileURL(uri *url.URL) (string, error) {
	if uri.User != nil {
		return "", os.ErrInvalid
	}

	if uri.Opaque != "" {
		name := uri.Opaque
		if !path.IsAbs(name) {
			return "", os.ErrInvalid
		}

		name = strings.TrimPrefix(name, "/")

		name, err := url.PathUnescape(name)
		if err != nil {
			return "", os.ErrInvalid
		}

		if !filepath.IsAbs(name) {
			return "", os.ErrInvalid
		}

		return filepath.Clean(filepath.FromSlash(name)), nil
	}

	name := uri.Path
	if !path.IsAbs(name) {
		return "", os.ErrInvalid
	}

	switch uri.Host {
	case "", ".":
		name = strings.TrimPrefix(name, "/")

		if !filepath.IsAbs(name) {
			return "", os.ErrInvalid
		}

		return filepath.Clean(filepath.FromSlash(name)), nil
	}

	name = filepath.Clean(filepath.FromSlash(name))

	return `\\` + uri.Host + name, nil
}
