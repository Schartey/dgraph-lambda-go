package resolver

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {

	t.Run("NewResolver", func(t *testing.T) {
		resolver := NewResolver()

		require.Nil(t, resolver.middleware)
		require.Len(t, resolver.middleware, 0)

		require.NotNil(t, resolver.webhooks)
		require.Len(t, resolver.webhooks, 0)

		require.NotNil(t, resolver.resolvers)
		require.Len(t, resolver.resolvers, 0)
	})

	t.Run("WebhookFunc Nil-Check", func(t *testing.T) {
		resolver := NewResolver()

		err := resolver.WebhookFunc("Test", nil)
		require.NotNil(t, err)
	})

	t.Run("WebhookFunc", func(t *testing.T) {
		resolver := NewResolver()

		webhookFunc := func(ctx context.Context, event Event) error {
			return nil
		}

		err := resolver.WebhookFunc("Test", webhookFunc)
		require.Nil(t, err)
		require.Len(t, resolver.webhooks, 1)
		require.Equal(t, reflect.ValueOf(webhookFunc).Pointer(), reflect.ValueOf(resolver.webhooks["Test"]).Pointer())
	})

	t.Run("ResolveFunc Nil-Check", func(t *testing.T) {
		resolver := NewResolver()

		err := resolver.ResolveFunc("Test", nil)
		require.NotNil(t, err)
	})

	t.Run("ResolveFunc", func(t *testing.T) {
		resolver := NewResolver()

		resolverFunc := func(ctx context.Context, input, parents []byte, authHeader AuthHeader) (interface{}, error) {
			return nil, nil
		}

		err := resolver.ResolveFunc("Test", resolverFunc)
		require.Nil(t, err)
		require.Len(t, resolver.resolvers, 1)
		require.Equal(t, reflect.ValueOf(resolverFunc).Pointer(), reflect.ValueOf(resolver.resolvers["Test"]).Pointer())
	})

	t.Run("Use Nil-Check", func(t *testing.T) {
		resolver := NewResolver()

		err := resolver.Use(nil)
		require.NotNil(t, err)
	})

	t.Run("Use", func(t *testing.T) {
		resolver := NewResolver()

		middlewareFunc := func(hf HandlerFunc) HandlerFunc {
			return hf
		}

		err := resolver.Use(middlewareFunc)
		require.Nil(t, err)
		require.Len(t, resolver.middleware, 1)
		require.Equal(t, "*", resolver.middleware[0].resolver)
		require.Equal(t, reflect.ValueOf(middlewareFunc).Pointer(), reflect.ValueOf(resolver.middleware[0].middlewareFunc).Pointer())
	})

	t.Run("UseOnResolver Nil-Check", func(t *testing.T) {
		resolver := NewResolver()

		err := resolver.UseOnResolver("Test", nil)
		require.NotNil(t, err)
	})

	t.Run("UseOnResolver", func(t *testing.T) {
		resolver := NewResolver()

		middlewareFunc := func(hf HandlerFunc) HandlerFunc {
			return hf
		}

		err := resolver.UseOnResolver("Test", middlewareFunc)
		require.Nil(t, err)
		require.Len(t, resolver.middleware, 1)
		require.Equal(t, "Test", resolver.middleware[0].resolver)
		require.Equal(t, reflect.ValueOf(middlewareFunc).Pointer(), reflect.ValueOf(resolver.middleware[0].middlewareFunc).Pointer())
	})

	t.Run("Resolve Nil-Check", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()
		res, err := resolver.Resolve(ctx, nil)
		require.NotNil(t, err)
		require.Nil(t, res)
	})

	t.Run("Resolve with empty dbody", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()
		res, err := resolver.Resolve(ctx, &DBody{})
		require.NotNil(t, err)
		require.Nil(t, res)
	})

	t.Run("Resolve without found webhook or resolver", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()
		res, err := resolver.Resolve(ctx, &DBody{Resolver: "Test"})
		require.NotNil(t, err)
		require.Nil(t, res)
	})

	t.Run("Resolve non existing webhook", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()
		res, err := resolver.Resolve(ctx, &DBody{Resolver: "$webhook"})
		require.NotNil(t, err)
		require.Nil(t, res)
	})

	t.Run("Resolve resolverFunc", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()

		jsonArgs := `{
			"key": "value"
		}`

		jsonParents := `[{
			"key": "value"
		}]`

		var interfaceArgs map[string]interface{}
		var interfaceParents []map[string]interface{}

		json.Unmarshal([]byte(jsonArgs), &interfaceArgs)
		json.Unmarshal([]byte(jsonParents), &interfaceParents)

		ah := AuthHeader{Key: "Key", Value: "Value"}

		resolverFuncCalled := false

		resolverFunc := func(ctx context.Context, input, parents []byte, authHeader AuthHeader) (interface{}, error) {
			resolverFuncCalled = true

			var resolvedArgs map[string]interface{}
			var resolvedParents []map[string]interface{}

			json.Unmarshal(input, &resolvedArgs)
			json.Unmarshal(parents, &resolvedParents)

			require.Equal(t, interfaceArgs, resolvedArgs)
			require.Equal(t, interfaceParents, resolvedParents)
			require.Equal(t, ah, authHeader)
			return resolvedArgs, nil
		}

		err := resolver.ResolveFunc("Test", resolverFunc)
		require.Nil(t, err)

		res, err := resolver.Resolve(ctx, &DBody{
			Resolver:    "Test",
			AccessToken: "AccessToken",
			Args:        interfaceArgs,
			Parents:     interfaceParents,
			AuthHeader:  ah,
			Event:       Event{}})
		require.Nil(t, err)
		require.NotNil(t, res)
		require.True(t, resolverFuncCalled)

		jsonResult, err := json.Marshal(interfaceArgs)
		require.Nil(t, err)

		require.Equal(t, jsonResult, res)
	})

	t.Run("Resolve webhook", func(t *testing.T) {
		ctx := context.Background()
		resolver := NewResolver()

		jsonParents := `[{
			"key": "value"
		}]`

		var interfaceParents []map[string]interface{}
		json.Unmarshal([]byte(jsonParents), &interfaceParents)

		ah := AuthHeader{Key: "Key", Value: "Value"}

		event := Event{
			TypeName:  "Test",
			CommitTs:  12039470,
			Operation: "Add",
			Add:       AddEventInfo{},
		}

		webhookFuncCalled := false

		webhookFunc := func(ctx context.Context, event Event) error {
			webhookFuncCalled = true
			return nil
		}

		err := resolver.WebhookFunc("Test", webhookFunc)
		require.Nil(t, err)

		res, err := resolver.Resolve(ctx, &DBody{
			Resolver:    "$webhook",
			AccessToken: "AccessToken",
			Parents:     interfaceParents,
			AuthHeader:  ah,
			Event:       event})
		require.Nil(t, err)
		require.Nil(t, res)
		require.True(t, webhookFuncCalled)
	})

	t.Run("applyMiddleware Nil-Check", func(t *testing.T) {
		dummyFunc := func(ctx context.Context, input, parents []byte, authHeader AuthHeader) (interface{}, error) {
			return nil, nil
		}

		var middleware []ResolverMiddlewareFunc

		middleware = append(middleware, ResolverMiddlewareFunc{resolver: "Test"})

		hf, err := applyMiddleware(dummyFunc, "Test", middleware...)
		require.NotNil(t, err)
		require.Nil(t, hf)
	})

	t.Run("applyMiddleware", func(t *testing.T) {
		dummyFunc := func(ctx context.Context, input, parents []byte, authHeader AuthHeader) (interface{}, error) {
			return nil, nil
		}

		middlewareFunc := func(hf HandlerFunc) HandlerFunc {
			return hf
		}

		var middleware []ResolverMiddlewareFunc

		middleware = append(middleware, ResolverMiddlewareFunc{resolver: "Test", middlewareFunc: middlewareFunc})

		hf, err := applyMiddleware(dummyFunc, "Test", middleware...)
		require.Nil(t, err)
		require.NotNil(t, hf)
		require.Equal(t, reflect.ValueOf(dummyFunc).Pointer(), reflect.ValueOf(hf).Pointer())
	})
}
