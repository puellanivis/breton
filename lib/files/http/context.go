package http

import (
	"net/http"
	"net/url"

	"context"
)

type key int

const (
	clientKey key = iota
	contentTypeKey
	contentKey
)

func WithClient(ctx context.Context, cl *http.Client) context.Context {
	return context.WithValue(ctx, clientKey, cl)
}

func getClient(ctx context.Context) (*http.Client, bool) {
	cl, ok := ctx.Value(clientKey).(*http.Client)
	return cl, ok
}

func WithQuery(ctx context.Context, vals url.Values) context.Context {
	body := []byte(vals.Encode())
	return WithContent(ctx, "application/x-www-form-urlencoded", body)
}

func WithContentType(ctx context.Context, ctype string) context.Context {
	return context.WithValue(ctx, contentTypeKey, ctype)
}

func getContentType(ctx context.Context) (string, bool) {
	ctype, ok := ctx.Value(contentTypeKey).(string)
	return ctype, ok
}

func WithContent(ctx context.Context, ctype string, data []byte) context.Context {
	return context.WithValue(WithContentType(ctx, ctype), contentKey, data)
}

func getContent(ctx context.Context) ([]byte, bool) {
	b, ok := ctx.Value(contentKey).([]byte)
	return b, ok
}
