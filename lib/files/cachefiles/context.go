package cachefiles

import (
	"context"
	"time"
)

type (
	keyExpiration struct{}
	keyReentrance struct{}
)

// WithExpire returns a Context that includes information for the cache FileStore to expire buffers after the given timeout.
func WithExpire(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, keyExpiration{}, timeout)
}

// GetExpire returns the expiration timeout specified for the given Context.
func GetExpire(ctx context.Context) (time.Duration, bool) {
	timeout, ok := ctx.Value(keyExpiration{}).(time.Duration)

	return timeout, ok
}

// isReentrant returns either:
// a new sub-context with a reentrance key attached, along with a false bool,
// or the same context input with a true bool.
func isReentrant(ctx context.Context) (rctx context.Context, reentrant bool) {
	if v := ctx.Value(keyReentrance{}); v != nil {
		return ctx, true
	}

	return context.WithValue(ctx, keyReentrance{}, struct{}{}), false
}
