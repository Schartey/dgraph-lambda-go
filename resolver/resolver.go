package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"plugin"
)

type Resolver struct {
	plugins    []*plugin.Plugin
	middleware []ResolverMiddlewareFunc
	resolvers  map[string]HandlerFunc
}

func NewResolver() *Resolver {
	return &Resolver{resolvers: make(map[string]HandlerFunc)}
}

func (r *Resolver) ResolveFunc(resolver string, handlerFunc HandlerFunc) {
	fmt.Printf("Loaded Resolver for %s\n", resolver)
	r.resolvers[resolver] = handlerFunc
}

func (r *Resolver) Use(middleware MiddlewareFunc) {
	resolverMiddleware := &ResolverMiddlewareFunc{Resolver: "*", middlewareFunc: middleware}
	r.middleware = append(r.middleware, *resolverMiddleware)
}

func (r *Resolver) UseOnResolver(resolver string, middleware MiddlewareFunc) {
	resolverMiddleware := &ResolverMiddlewareFunc{Resolver: resolver, middlewareFunc: middleware}
	r.middleware = append(r.middleware, *resolverMiddleware)
}

func (r *Resolver) Resolve(ctx context.Context, dbody *DBody) ([]byte, error) {
	args, err := json.Marshal(dbody.Args)
	if err != nil {
		fmt.Println("Could not marshal arguments")
	}
	parents, err := json.Marshal(dbody.Parents)
	if err != nil {
		fmt.Println("Could not marshal parents")
	}

	if r.resolvers[dbody.Resolver] == nil {
		return nil, errors.New(fmt.Sprintf("Could not resolve %s", dbody.Resolver))
	}

	h := r.resolvers[dbody.Resolver]

	h = applyMiddleware(h, dbody.Resolver, r.middleware...)

	res, err := h(ctx, args, parents, dbody.AuthHeader)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func applyMiddleware(h HandlerFunc, resolver string, middleware ...ResolverMiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		if middleware[i].Resolver == "*" {
			h = middleware[i].middlewareFunc(h)
		} else if middleware[i].Resolver == resolver {
			h = middleware[i].middlewareFunc(h)
		}
	}
	return h
}
