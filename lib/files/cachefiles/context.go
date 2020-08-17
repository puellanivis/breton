package cachefiles

import (
	"context"
	"time"
)

type (
	expireKey     struct{}
	reentranceKey struct{}
)

// WithExpire returns a Context that includes information for the cache FileStore to expire buffers after the given timeout.
func WithExpire(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, expireKey{}, timeout)
}

// GetExpire returns the expiration timeout specified for the given Context.
func GetExpire(ctx context.Context) (time.Duration, bool) {
	timeout, ok := ctx.Value(expireKey{}).(time.Duration)

	return timeout, ok
}

// isReentrant returns either:
// a new sub-context with a reentrance key attached, along with a false bool,
// or the same context input with a true bool.
func isReentrant(ctx context.Context) (rctx context.Context, reentrant bool) {
	if v := ctx.Value(reentranceKey{}); v != nil {
		return ctx, true
	}

	return context.WithValue(ctx, reentranceKey{}, struct{}{}), false
}
