package http

import (
	"net/http"

	"context"
)

type key int

const (
	clientKey key = iota
	userAgentKey
)

// WithClient attaches an http.Client to the Context, that will be used by this library as the http.Client
func WithClient(ctx context.Context, cl *http.Client) context.Context {
	return context.WithValue(ctx, clientKey, cl)
}

func getClient(ctx context.Context) (*http.Client, bool) {
	cl, ok := ctx.Value(clientKey).(*http.Client)
	return cl, ok
}

// WithUserAgent attaches an string to the Context, that will be used by this library as the User-Agent in all headers
func WithUserAgent(ctx context.Context, ua string) context.Context {
	return context.WithValue(ctx, userAgentKey, ua)
}

func getUserAgent(ctx context.Context) (string, bool) {
	ua, ok := ctx.Value(userAgentKey).(string)
	return ua, ok
}
