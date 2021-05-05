package resolver

import (
	"context"
)

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type HandlerFunc func(context.Context, []byte, AuthHeader) ([]byte, error)
