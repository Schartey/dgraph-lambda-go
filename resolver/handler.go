package resolver

import (
	"context"
)

type ResolverMiddlewareFunc struct {
	Resolver       string
	middlewareFunc MiddlewareFunc
}

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type HandlerFunc func(context.Context, []byte, []byte, AuthHeader) ([]byte, error)
