package files

import (
	"context"
)

type key int

const (
	rootKey key = iota
)

func WithRoot(ctx context.Context, root string) context.Context {
	return context.WithValue(ctx, rootKey, root)
}

func GetRoot(ctx context.Context) (string, bool) {
	root, ok := ctx.Value(rootKey).(string)
	return root, ok
}
