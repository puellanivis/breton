package cachefiles

import (
	"time"

	"context"
)

type key int

const (
	expireKey key = iota
)

func WithExpire(ctx context.Context, timeout time.Duration) context.Context {
	return context.WithValue(ctx, expireKey, timeout)
}

func GetExpire(ctx context.Context) (time.Duration, bool) {
	timeout, ok := ctx.Value(expireKey).(time.Duration)

	return timeout, ok
}
