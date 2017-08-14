package files

import (
	"context"
	"net/url"
)

type key int

const (
	rootKey key = iota
)

// WithRootURL attaches a url.URL to a Context
// and is used as the resolution reference for any files.Open() using that context.
func WithRootURL(ctx context.Context, uri *url.URL) context.Context {
	return context.WithValue(ctx, rootKey, uri)
}

// WithRoot parses the root as a URL, then attaches it to the Context as per WithRootURL.
func WithRoot(ctx context.Context, root string) (context.Context, error) {
	uri, err := url.Parse(root)
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

	return root.String(), true
}
