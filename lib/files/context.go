package files

import (
	"context"
	"net/url"
)

type (
	rootKey struct{}
)

func withRoot(ctx context.Context, root *url.URL) context.Context {
	return context.WithValue(ctx, rootKey{}, root)
}

// WithRootURL attaches a url.URL to a Context
// and is used as the resolution reference for any files.Open() using that context.
func WithRootURL(ctx context.Context, uri *url.URL) context.Context {
	return withRoot(ctx, resolveURL(ctx, uri))
}

// WithRoot stores either a URL or a local path to use as a root point when resolving filenames.
func WithRoot(ctx context.Context, path string) (context.Context, error) {
	return withRoot(ctx, parsePath(ctx, path)), nil
}

func getRoot(ctx context.Context) (*url.URL, bool) {
	root, ok := ctx.Value(rootKey{}).(*url.URL)
	return root, ok
}

// GetRoot returns the currently attached string that is being used as the root for any invalid URLs.
func GetRoot(ctx context.Context) (string, bool) {
	root, ok := getRoot(ctx)
	if !ok {
		return "", false
	}

	if isPath(root) {
		return root.Path, true
	}

	return root.String(), true
}
