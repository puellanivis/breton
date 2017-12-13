package files

import (
	"context"
	"net/url"
	"path/filepath"
)

type key int

const (
	rootKey key = iota
)

// WithRootURL attaches a url.URL to a Context
// and is used as the resolution reference for any files.Open() using that context.
func WithRootURL(ctx context.Context, uri *url.URL) context.Context {
	uriCopy := uri
	uri = resolveFilename(ctx, uri)

	// we got the same URL back, clone it so that it stays immutable to the original uri passed in.
	if uriCopy == uri {
		uriCopy = new(url.URL)
		*uriCopy = *uri

		if uri.User != nil {
			uriCopy.User = new(url.Userinfo)
			*uriCopy.User = *uri.User // gotta copy this pointer struct also.
		}

		uri = uriCopy
	}
	
	return context.WithValue(ctx, rootKey, uri)
}

// WithRoot stores either a URL or a local path to use as a root point when resolving filenames.
func WithRoot(ctx context.Context, path string) (context.Context, error) {
	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
		return WithRootURL(ctx, makePath(path)), nil
	}

	uri, err := url.Parse(path)
	if err != nil {
		return ctx, err
	}

	return WithRootURL(ctx, uri), nil
}

func getRoot(ctx context.Context) (*url.URL, bool) {
	root, ok := ctx.Value(rootKey).(*url.URL)
	return root, ok
}

// GetRoot returns the currently attached string that is being used as the root for any invalid URLs.
func GetRoot(ctx context.Context) (string, bool) {
	root, ok := getRoot(ctx)
	if !ok {
		return "", false
	}

	if isPath(root) {
		return getPath(root), true
	}

	return root.String(), true
}
