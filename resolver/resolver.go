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
	webhooks   map[string]WebHookFunc
}

func NewResolver() *Resolver {
	return &Resolver{
		resolvers: make(map[string]HandlerFunc),
		webhooks:  make(map[string]WebHookFunc),
	}
}

func (r *Resolver) WebHookFunc(typeName string, webhookFunc WebHookFunc) {
	fmt.Printf("Loaded Webhook for %s\n", typeName)
	r.webhooks[typeName] = webhookFunc
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
	if dbody.Resolver == "$webhook" {
		if r.webhooks[dbody.Event.TypeName] == nil {
			return nil, errors.New(fmt.Sprintf("Could not resolve webhook %s", dbody.Event.TypeName))
		}
		err := r.webhooks[dbody.Event.TypeName](ctx, dbody.Event)
		return nil, err
	} else {
		if r.resolvers[dbody.Resolver] == nil {
			return nil, errors.New(fmt.Sprintf("Could not resolve %s", dbody.Resolver))
		}

		args, err := json.Marshal(dbody.Args)
		if err != nil {
			fmt.Println("Could not marshal parents")
		}

		parents, err := json.Marshal(dbody.Parents)
		if err != nil {
			fmt.Println("Could not marshal parents")
		}

		h := r.resolvers[dbody.Resolver]

		h = applyMiddleware(h, dbody.Resolver, r.middleware...)

		res, err := h(ctx, args, parents, dbody.AuthHeader)
		if err != nil {
			return nil, err
		}
		resBytes, err := json.Marshal(res)
		if err != nil {
			return nil, err
		}
		return resBytes, nil
	}
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
