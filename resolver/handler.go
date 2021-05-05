package resolver

import (
	"context"

	"github.com/dgraph-io/dgo/v210"
)

type MiddlewareFunc func(HandlerFunc) HandlerFunc

type HandlerFunc func(context.Context, []byte, AuthHeader, *dgo.Dgraph) ([]byte, error)
