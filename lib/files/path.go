package files

import (
	"context"
	"net/url"
	"path/filepath"
)

func isPath(uri *url.URL) bool {
	switch {
	case uri.IsAbs():
		return false

	case uri.User != nil:
		return false

	case len(uri.Host)+len(uri.RawQuery)+len(uri.Fragment) > 0:
		return false

	case uri.ForceQuery:
		return false
	}

	return true
}

func resolveURL(ctx context.Context, uri *url.URL) *url.URL {
	if uri.IsAbs() {
		// short-circuit: If the url is absolute,
		// then we should never consider resolving it as a reference relative to root.
		return uri
	}

	if root, ok := getRoot(ctx); ok {
		switch {
		case !isPath(root):
			// If root is not a path-only URL,
			// then always resolve the uri as a reference.
			return root.ResolveReference(uri)

		case isPath(uri):
			// special-case: If both root and uri are wrapped simple paths,
			// then join their Paths through filepath.Join,
			// instead of using URL path handling.
			return &url.URL{
				Path: filepath.Join(root.Path, uri.Path),
			}
		}

		// root is a wrapped simple path, but uri is not,
		// there’s no good way to really join these two together.
		// fallthrough to not even considering root.
	}

	if isPath(uri) {
		uri.Path = filepath.Clean(uri.Path)
	}

	return uri
}

// parsePath will always return a non-nil `*url.URL`.
//
// If the path is an invalid URL, then we will return a wrapped simple path,
// which is simply a &url.URL{ Path: path }.
func parsePath(ctx context.Context, path string) *url.URL {
	if filepath.IsAbs(path) {
		// If for this architecture, path is an an absolute path
		// then we should only ever treat it as a wrapped simple path.
		return &url.URL{
			Path: filepath.Clean(path),
		}
	}

	uri, err := url.Parse(path)
	if err != nil {
		// If this path fails to parse as a URL, treat it like a wrapped simple path.
		uri = &url.URL{
			Path: path,
		}
	}

	// Since we do not `filepath.Clean` the wrapped simple path from above,
	// this function must assure that the wrapped simple path is cleaned before returning the url.
	return resolveURL(ctx, uri)
}
