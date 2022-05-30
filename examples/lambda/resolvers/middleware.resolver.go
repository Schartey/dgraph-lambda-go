package resolvers

import (
	"github.com/miko/dgraph-lambda-go/api"
)

type MiddlewareResolverInterface interface {
	Middleware_admin(mc *api.MiddlewareContext) *api.LambdaError
	Middleware_user(mc *api.MiddlewareContext) *api.LambdaError
}

type MiddlewareResolver struct {
	*Resolver
}

func (m *MiddlewareResolver) Middleware_admin(mc *api.MiddlewareContext) *api.LambdaError {
	return nil
}

func (m *MiddlewareResolver) Middleware_user(mc *api.MiddlewareContext) *api.LambdaError {
	return nil
}
