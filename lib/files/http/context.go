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

// WithClient attaches an http.Client to the Context, that will be used by this library as the http.Client
func WithClient(ctx context.Context, cl *http.Client) context.Context {
	return context.WithValue(ctx, clientKey, cl)
}

func getClient(ctx context.Context) (*http.Client, bool) {
	cl, ok := ctx.Value(clientKey).(*http.Client)
	return cl, ok
}

// WithQuery encodes a set of www form values, which is then attached to a Contex, and used by this libraray as a POST query to the http requests.
func WithQuery(ctx context.Context, vals url.Values) context.Context {
	body := []byte(vals.Encode())
	return WithContent(ctx, "application/x-www-form-urlencoded", body)
}

// WithContentType attaches a Content-Type value to a Context, which is used by this library during any POST requests. (The default GET requests will remain without a Content-Type.)
func WithContentType(ctx context.Context, ctype string) context.Context {
	return context.WithValue(ctx, contentTypeKey, ctype)
}

func getContentType(ctx context.Context) (string, bool) {
	ctype, ok := ctx.Value(contentTypeKey).(string)
	return ctype, ok
}

// WithContent attaches a byte slice to a Context, which is then used by this library as a body for a POST to any http requests.
// TODO: this shouldn’t be done as a context function, but rather a call on the file returned by an http open/create.
func WithContent(ctx context.Context, ctype string, data []byte) context.Context {
	return context.WithValue(WithContentType(ctx, ctype), contentKey, data)
}

func getContent(ctx context.Context) ([]byte, bool) {
	b, ok := ctx.Value(contentKey).([]byte)
	return b, ok
}
