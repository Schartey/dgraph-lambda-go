package resolver

import (
	"context"
)

type ResolverMiddlewareFunc struct {
	Resolver       string
	middlewareFunc MiddlewareFunc
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type HandlerFunc func(ctx context.Context, input []byte, parents []byte, authHeader AuthHeader) (interface{}, error)

type WebHookFunc func(ctx context.Context, event Event) error
