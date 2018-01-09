package files

import (
	"context"
	"net/url"
	"path/filepath"
)

func isPath(uri *url.URL) bool {
	if uri.IsAbs() {
		return false
	}

	if uri.User != nil {
		return false
	}

	if len(uri.Host)+len(uri.RawQuery)+len(uri.Fragment) > 0 {
		return false
	}

	if uri.ForceQuery {
		return false
	}

	return true
}

func getPath(uri *url.URL) string {
	if uri.RawPath != "" {
		return uri.RawPath
	}

	return uri.Path
}

func makePath(path string) *url.URL {
	return &url.URL{
		Path:    path,
		RawPath: path,
	}
}

func resolveFilename(ctx context.Context, uri *url.URL) *url.URL {
	if uri.IsAbs() {
		return uri
	}

	var path string

	if isPath(uri) {
		path = getPath(uri)

		if filepath.IsAbs(path) {
			return makePath(path)
		}
	}

	root, ok := getRoot(ctx)
	if !ok {
		return uri
	}

	if path != "" && isPath(root) {
		return makePath(filepath.Join(getPath(root), path))

	}

	return root.ResolveReference(uri)
}
