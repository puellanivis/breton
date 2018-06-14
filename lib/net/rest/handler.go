package rest

import (
	"context"
)

type Handler interface {
	ServeREST(context.Context, []string, *Request) (interface{}, error)
}

type HandlerFunc func(context.Context, []string, *Request) (interface{}, error)

func (f HandlerFunc) ServeREST(ctx context.Context, path []string, req *Request) (interface{}, error) {
	return f(ctx, path, req)
}
