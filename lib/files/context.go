package files

import (
	"context"
)

type key int

const (
	rootKey key = iota
)

// WithRoot attaches a string to a Context that is used as the prefix to any files.Open() using that context that is not otherwise a valid URL.
func WithRoot(ctx context.Context, root string) context.Context {
	return context.WithValue(ctx, rootKey, root)
}

// GetRoot returns the currently attached string that is being used as the root for any invalid URLs.
func GetRoot(ctx context.Context) (string, bool) {
	root, ok := ctx.Value(rootKey).(string)
	return root, ok
}
