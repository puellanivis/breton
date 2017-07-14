package http

import (
	"net/http"

	"context"
)

type key int

const (
	clientKey key = iota
)

// WithClient attaches an http.Client to the Context, that will be used by this library as the http.Client
func WithClient(ctx context.Context, cl *http.Client) context.Context {
	return context.WithValue(ctx, clientKey, cl)
}

func getClient(ctx context.Context) (*http.Client, bool) {
	cl, ok := ctx.Value(clientKey).(*http.Client)
	return cl, ok
}
