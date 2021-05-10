package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type Resolver struct {
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

func (r *Resolver) WebhookFunc(typeName string, webhookFunc WebHookFunc) error {
	if webhookFunc == nil {
		return errors.New("WebhookFunc cannot be nil")
	}

	fmt.Printf("Loaded Webhook for %s\n", typeName)
	r.webhooks[typeName] = webhookFunc

	return nil
}

func (r *Resolver) ResolveFunc(resolver string, handlerFunc HandlerFunc) error {
	if handlerFunc == nil {
		return errors.New("HandlerFunc cannot be nil")
	}
	fmt.Printf("Loaded Resolver for %s\n", resolver)
	r.resolvers[resolver] = handlerFunc
	return nil
}

func (r *Resolver) Use(middleware MiddlewareFunc) error {
	if middleware == nil {
		return errors.New("Middlware cannot be nil")
	}
	resolverMiddleware := &ResolverMiddlewareFunc{resolver: "*", middlewareFunc: middleware}
	r.middleware = append(r.middleware, *resolverMiddleware)
	return nil
}

func (r *Resolver) UseOnResolver(resolver string, middleware MiddlewareFunc) error {
	if middleware == nil {
		return errors.New("Middleware cannot be nil")
	}
	resolverMiddleware := &ResolverMiddlewareFunc{resolver: resolver, middlewareFunc: middleware}
	r.middleware = append(r.middleware, *resolverMiddleware)
	return nil
}

func (r *Resolver) Resolve(ctx context.Context, dbody *DBody) ([]byte, error) {
	if dbody == nil {
		return nil, errors.New("DBody cannot be nil")
	}
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

		h, err = applyMiddleware(h, dbody.Resolver, r.middleware...)
		if err != nil {
			return nil, err
		}

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

func applyMiddleware(h HandlerFunc, resolver string, middleware ...ResolverMiddlewareFunc) (HandlerFunc, error) {
	for i := len(middleware) - 1; i >= 0; i-- {
		if middleware[i].resolver == "*" {
			if middleware[i].middlewareFunc == nil {
				return nil, errors.New("A middleware func was nil, cancelling resolution")
			}
			h = middleware[i].middlewareFunc(h)
		} else if middleware[i].resolver == resolver {
			if middleware[i].middlewareFunc == nil {
				return nil, errors.New("A middleware func was nil, cancelling resolution")
			}
			h = middleware[i].middlewareFunc(h)
		}
	}
	return h, nil
}
